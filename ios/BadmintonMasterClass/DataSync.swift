import Foundation

// MARK: - OSS Configuration
//
// The sync URL is built from two pieces:
//   bucket   — Aliyun OSS bucket name  (matches BMC_OSS_BUCKET env var from admin scripts)
//   endpoint — OSS endpoint hostname   (matches BMC_OSS_ENDPOINT, e.g. oss-cn-hangzhou.aliyuncs.com)
//
// Defaults can be overridden **without rebuilding** by adding keys to Info.plist:
//   BMC_OSS_BUCKET   — e.g. "my-prod-bucket"
//   BMC_OSS_ENDPOINT — e.g. "oss-cn-shanghai.aliyuncs.com"
//
// Resulting URL: https://{bucket}.{endpoint}/bmc.db

enum SyncConfig {
    static var bucket: String {
        Bundle.main.object(forInfoDictionaryKey: "BMC_OSS_BUCKET") as? String
            ?? "bmc-data"
    }

    static var endpoint: String {
        Bundle.main.object(forInfoDictionaryKey: "BMC_OSS_ENDPOINT") as? String
            ?? "oss-cn-hangzhou.aliyuncs.com"
    }

    static var remoteURL: URL {
        URL(string: "https://\(bucket).\(endpoint)/bmc.db")!
    }
}

// MARK: - ETag Storage

private enum ETagStore {
    private static let key = "bmc_sync_etag"

    static var lastETag: String? {
        get { UserDefaults.standard.string(forKey: key) }
        set { UserDefaults.standard.set(newValue, forKey: key) }
    }
}

// MARK: - DataSync

enum DataSync {
    private static var remoteURL: URL { SyncConfig.remoteURL }

    /// Async entry point for pull-to-refresh.
    static func syncDatabase() async {
        await performSync()
    }

    /// Fire-and-forget entry point (e.g. app launch).
    static func syncIfNeeded() {
        Task { await performSync() }
    }

    // MARK: - Core sync logic (single implementation)

    private static func performSync() async {
        await MainActor.run { SyncManager.shared.setSyncing() }

        do {
            var request = URLRequest(url: remoteURL)
            request.timeoutInterval = 30

            // Conditional fetch: send stored ETag so the server can return 304
            if let etag = ETagStore.lastETag {
                request.setValue(etag, forHTTPHeaderField: "If-None-Match")
            }

            let (tempURL, response) = try await URLSession.shared.download(for: request)

            guard let httpResponse = response as? HTTPURLResponse else {
                print("[DataSync] Invalid response type")
                await MainActor.run { SyncManager.shared.setFailed() }
                return
            }

            // 304 Not Modified — local data is already up-to-date
            if httpResponse.statusCode == 304 {
                print("[DataSync] 304 Not Modified — skipping download")
                await MainActor.run { SyncManager.shared.setSuccess() }
                return
            }

            guard (200...299).contains(httpResponse.statusCode) else {
                print("[DataSync] Server returned status \(httpResponse.statusCode)")
                await MainActor.run { SyncManager.shared.setFailed() }
                return
            }

            // Save the new ETag for next request
            if let newETag = httpResponse.value(forHTTPHeaderField: "ETag") {
                ETagStore.lastETag = newETag
            }

            // Move to a stable temporary location (the download temp file may be deleted)
            let stableTemp = FileManager.default.temporaryDirectory
                .appendingPathComponent(UUID().uuidString + ".db")
            try FileManager.default.moveItem(at: tempURL, to: stableTemp)

            await MainActor.run {
                Database.shared.replaceWith(downloadedDBAt: stableTemp)
                SyncManager.shared.setSuccess()
            }
        } catch {
            print("[DataSync] Download failed: \(error.localizedDescription)")
            await MainActor.run { SyncManager.shared.setFailed() }
        }
    }
}

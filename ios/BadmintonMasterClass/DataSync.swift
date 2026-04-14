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

enum DataSync {
    private static var remoteURL: URL { SyncConfig.remoteURL }

    /// Download the latest DB from the remote URL and replace the local copy.
    /// Reports progress via SyncManager. Failures are non-fatal — the app continues with local data.
    static func syncIfNeeded() {
        Task { @MainActor in
            SyncManager.shared.setSyncing()
        }

        let task = URLSession.shared.downloadTask(with: remoteURL) { tempURL, response, error in
            guard let tempURL = tempURL, error == nil else {
                print("[DataSync] Download failed: \(error?.localizedDescription ?? "unknown error")")
                Task { @MainActor in SyncManager.shared.setFailed() }
                return
            }

            // Verify we got a successful HTTP response
            if let httpResponse = response as? HTTPURLResponse,
               !(200...299).contains(httpResponse.statusCode) {
                print("[DataSync] Server returned status \(httpResponse.statusCode)")
                Task { @MainActor in SyncManager.shared.setFailed() }
                return
            }

            // Move to a stable temporary location (the download temp file may be deleted)
            let stableTemp = FileManager.default.temporaryDirectory
                .appendingPathComponent(UUID().uuidString + ".db")
            do {
                try FileManager.default.moveItem(at: tempURL, to: stableTemp)
            } catch {
                print("[DataSync] Failed to stage downloaded file: \(error)")
                Task { @MainActor in SyncManager.shared.setFailed() }
                return
            }

            DispatchQueue.main.async {
                Database.shared.replaceWith(downloadedDBAt: stableTemp)
                SyncManager.shared.setSuccess()
            }
        }
        task.resume()
    }
}

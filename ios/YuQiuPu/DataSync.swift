import Foundation

enum DataSync {
    private static let remoteURL = URL(string: "https://your-bucket.oss-cn-hangzhou.aliyuncs.com/yuqiupu.db")!

    /// Download the latest DB from the remote URL and replace the local copy.
    /// Failures are silently ignored — the app continues with local data.
    static func syncIfNeeded() {
        let task = URLSession.shared.downloadTask(with: remoteURL) { tempURL, response, error in
            guard let tempURL = tempURL, error == nil else {
                print("[DataSync] Download failed: \(error?.localizedDescription ?? "unknown error")")
                return
            }

            // Verify we got a successful HTTP response
            if let httpResponse = response as? HTTPURLResponse,
               !(200...299).contains(httpResponse.statusCode) {
                print("[DataSync] Server returned status \(httpResponse.statusCode)")
                return
            }

            // Move to a stable temporary location (the download temp file may be deleted)
            let stableTemp = FileManager.default.temporaryDirectory
                .appendingPathComponent(UUID().uuidString + ".db")
            do {
                try FileManager.default.moveItem(at: tempURL, to: stableTemp)
            } catch {
                print("[DataSync] Failed to stage downloaded file: \(error)")
                return
            }

            DispatchQueue.main.async {
                Database.shared.replaceWith(downloadedDBAt: stableTemp)
            }
        }
        task.resume()
    }
}

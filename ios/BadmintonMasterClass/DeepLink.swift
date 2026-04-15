import UIKit

/// Computes a native app deep link URL from a web URL and platform identifier.
/// Returns nil if the platform has no deep link scheme or the URL cannot be parsed.
enum DeepLink {

    /// Attempt to open the content in its native app; fall back to the web URL.
    /// - Parameters:
    ///   - sourceUrl: The original web URL string.
    ///   - sourcePlatform: Platform identifier (e.g. "bilibili", "douyin").
    ///   - fallback: Called with the original URL when the native app is not installed.
    static func open(sourceUrl: String, sourcePlatform: String, fallback: @escaping (URL) -> Void) {
        guard let webURL = URL(string: sourceUrl) else { return }

        if let deepURL = deepLinkURL(sourceUrl: sourceUrl, sourcePlatform: sourcePlatform) {
            UIApplication.shared.open(deepURL, options: [:]) { success in
                if !success {
                    fallback(webURL)
                }
            }
        } else {
            fallback(webURL)
        }
    }

    /// Pure computation: web URL + platform -> deep link URL (or nil).
    static func deepLinkURL(sourceUrl: String, sourcePlatform: String) -> URL? {
        switch sourcePlatform {
        case "bilibili":
            return bilibiliDeepLink(sourceUrl)
        case "youtube":
            return youtubeDeepLink(sourceUrl)
        case "xiaohongshu":
            return xiaohongshuDeepLink(sourceUrl)
        case "douyin":
            return douyinDeepLink(sourceUrl)
        case "wechat":
            // WeChat articles — no reliable deep link; always open in web view
            return nil
        default:
            return nil
        }
    }

    // MARK: - Platform-specific deep links

    /// bilibili.com/video/BVxxx -> bilibili://video/BVxxx
    /// b23.tv short links -> nil (opened in browser, which follows the redirect)
    private static func bilibiliDeepLink(_ urlString: String) -> URL? {
        guard let url = URL(string: urlString),
              let host = url.host else {
            return nil
        }
        // b23.tv short links redirect to the real URL; let the browser handle it.
        if host.contains("b23.tv") {
            return nil
        }
        guard host.contains("bilibili.com"),
              url.pathComponents.count >= 3,
              url.pathComponents[1] == "video" else {
            return nil
        }
        let bvid = url.pathComponents[2]
        return URL(string: "bilibili://video/\(bvid)")
    }

    /// youtube.com/watch?v=xxx -> vnd.youtube:xxx
    private static func youtubeDeepLink(_ urlString: String) -> URL? {
        guard let url = URL(string: urlString),
              let host = url.host,
              (host.contains("youtube.com") || host.contains("youtu.be")),
              let components = URLComponents(url: url, resolvingAgainstBaseURL: false) else {
            return nil
        }
        // youtube.com/watch?v=xxx
        if let videoId = components.queryItems?.first(where: { $0.name == "v" })?.value {
            return URL(string: "vnd.youtube:\(videoId)")
        }
        // youtu.be/xxx
        if host.contains("youtu.be"), url.pathComponents.count >= 2 {
            return URL(string: "vnd.youtube:\(url.pathComponents[1])")
        }
        return nil
    }

    /// xiaohongshu.com/explore/xxx -> xhsdiscover://item/xxx
    /// xiaohongshu.com/discovery/item/xxx -> xhsdiscover://item/xxx
    /// xhslink.com/xxx -> xhsdiscover://item/xxx (short URL)
    private static func xiaohongshuDeepLink(_ urlString: String) -> URL? {
        guard let url = URL(string: urlString),
              let host = url.host else {
            return nil
        }

        // xhslink.com short URLs — open directly; the app's URL scheme handles resolution
        if host.contains("xhslink.com"), url.pathComponents.count >= 2 {
            let shortCode = url.pathComponents[1]
            return URL(string: "xhsdiscover://item/\(shortCode)")
        }

        guard host.contains("xiaohongshu.com") else { return nil }

        // /explore/xxx
        if url.pathComponents.count >= 3, url.pathComponents[1] == "explore" {
            let noteId = url.pathComponents[2]
            return URL(string: "xhsdiscover://item/\(noteId)")
        }

        // /discovery/item/xxx
        if url.pathComponents.count >= 4,
           url.pathComponents[1] == "discovery",
           url.pathComponents[2] == "item" {
            let noteId = url.pathComponents[3]
            return URL(string: "xhsdiscover://item/\(noteId)")
        }

        return nil
    }

    /// douyin.com/video/xxx -> snssdk1128://feed?id=xxx
    private static func douyinDeepLink(_ urlString: String) -> URL? {
        guard let url = URL(string: urlString),
              let host = url.host,
              host.contains("douyin.com"),
              url.pathComponents.count >= 3,
              url.pathComponents[1] == "video" else {
            return nil
        }
        let videoId = url.pathComponents[2]
        return URL(string: "snssdk1128://feed?id=\(videoId)")
    }
}

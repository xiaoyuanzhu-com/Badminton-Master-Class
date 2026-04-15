package com.bmc.app.util

import android.content.ActivityNotFoundException
import android.content.Context
import android.content.Intent
import android.net.Uri
import androidx.browser.customtabs.CustomTabsIntent

/**
 * Computes native-app deep link URIs from web URLs and opens them,
 * falling back to a Custom Tab when the native app is not installed.
 */
object DeepLink {

    /**
     * Try to open the content in its native app.
     * Falls back to Chrome Custom Tab when no handler is found.
     */
    fun open(context: Context, sourceUrl: String, sourcePlatform: String) {
        val deepUri = deepLinkUri(sourceUrl, sourcePlatform)
        if (deepUri != null) {
            try {
                val intent = Intent(Intent.ACTION_VIEW, deepUri).apply {
                    flags = Intent.FLAG_ACTIVITY_NEW_TASK
                }
                context.startActivity(intent)
                return
            } catch (_: ActivityNotFoundException) {
                // App not installed — fall through to Custom Tab
            }
        }
        // Fallback: open in Custom Tab
        val customTab = CustomTabsIntent.Builder().build()
        customTab.launchUrl(context, Uri.parse(sourceUrl))
    }

    /**
     * Pure computation: web URL + platform -> deep link URI (or null).
     */
    fun deepLinkUri(sourceUrl: String, sourcePlatform: String): Uri? {
        return when (sourcePlatform) {
            "bilibili" -> bilibiliDeepLink(sourceUrl)
            "youtube" -> youtubeDeepLink(sourceUrl)
            "xiaohongshu" -> xiaohongshuDeepLink(sourceUrl)
            "douyin" -> douyinDeepLink(sourceUrl)
            "wechat" -> null // WeChat articles — no reliable deep link
            else -> null
        }
    }

    /** bilibili.com/video/BVxxx -> bilibili://video/BVxxx */
    /** b23.tv short links -> null (opened in Custom Tab, which follows the redirect) */
    private fun bilibiliDeepLink(urlString: String): Uri? {
        val uri = Uri.parse(urlString)
        val host = uri.host ?: return null
        // b23.tv short links redirect to the real URL; let the Custom Tab handle it.
        if (host.contains("b23.tv")) return null
        if (!host.contains("bilibili.com")) return null
        val segments = uri.pathSegments
        if (segments.size < 2 || segments[0] != "video") return null
        return Uri.parse("bilibili://video/${segments[1]}")
    }

    /** youtube.com/watch?v=xxx -> vnd.youtube:xxx */
    private fun youtubeDeepLink(urlString: String): Uri? {
        val uri = Uri.parse(urlString)
        val host = uri.host ?: return null
        if (host.contains("youtube.com")) {
            val videoId = uri.getQueryParameter("v") ?: return null
            return Uri.parse("vnd.youtube:$videoId")
        }
        if (host.contains("youtu.be")) {
            val segments = uri.pathSegments
            if (segments.isNotEmpty()) {
                return Uri.parse("vnd.youtube:${segments[0]}")
            }
        }
        return null
    }

    /**
     * xiaohongshu.com/explore/xxx -> xhsdiscover://item/xxx
     * xiaohongshu.com/discovery/item/xxx -> xhsdiscover://item/xxx
     * xhslink.com/xxx -> xhsdiscover://item/xxx (short URL)
     */
    private fun xiaohongshuDeepLink(urlString: String): Uri? {
        val uri = Uri.parse(urlString)
        val host = uri.host ?: return null
        val segments = uri.pathSegments

        // xhslink.com short URLs — open directly; the app's URL scheme handles resolution
        if (host.contains("xhslink.com") && segments.isNotEmpty()) {
            return Uri.parse("xhsdiscover://item/${segments[0]}")
        }

        if (!host.contains("xiaohongshu.com")) return null

        // /explore/xxx
        if (segments.size >= 2 && segments[0] == "explore") {
            return Uri.parse("xhsdiscover://item/${segments[1]}")
        }

        // /discovery/item/xxx
        if (segments.size >= 3 && segments[0] == "discovery" && segments[1] == "item") {
            return Uri.parse("xhsdiscover://item/${segments[2]}")
        }

        return null
    }

    /** douyin.com/video/xxx -> snssdk1128://feed?id=xxx */
    private fun douyinDeepLink(urlString: String): Uri? {
        val uri = Uri.parse(urlString)
        val host = uri.host ?: return null
        if (!host.contains("douyin.com")) return null
        val segments = uri.pathSegments
        if (segments.size < 2 || segments[0] != "video") return null
        return Uri.parse("snssdk1128://feed?id=${segments[1]}")
    }
}

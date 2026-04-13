package com.yuqiupu.app.data

import android.content.Context
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.io.File
import java.net.HttpURLConnection
import java.net.URL

object DataSync {

    private const val REMOTE_URL =
        "https://your-bucket.oss-cn-hangzhou.aliyuncs.com/yuqiupu.db"

    /**
     * Download the latest DB from the remote URL and replace the local copy.
     * Failures are silently ignored — the app continues with local data.
     */
    suspend fun syncIfNeeded(context: Context) {
        withContext(Dispatchers.IO) {
            try {
                val url = URL(REMOTE_URL)
                val connection = url.openConnection() as HttpURLConnection
                connection.connectTimeout = 10_000
                connection.readTimeout = 30_000

                try {
                    if (connection.responseCode !in 200..299) {
                        return@withContext
                    }

                    val tempFile = File(context.cacheDir, "yuqiupu_download.db")
                    connection.inputStream.use { input ->
                        tempFile.outputStream().use { output ->
                            input.copyTo(output)
                        }
                    }

                    Database.getInstance(context).replaceWith(tempFile)
                } finally {
                    connection.disconnect()
                }
            } catch (e: Exception) {
                // Silently ignore — continue with local data
                e.printStackTrace()
            }
        }
    }
}

package com.bmc.app.data

import android.content.Context
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.withContext
import java.io.File
import java.net.HttpURLConnection
import java.net.URL

/** Represents the current state of data synchronisation. */
sealed interface SyncState {
    data object Idle : SyncState
    data object Syncing : SyncState
    data object Success : SyncState
    data object Failed : SyncState
}

// ---------------------------------------------------------------------------
// OSS Configuration
//
// The sync URL is built from two pieces:
//   bucket   — Aliyun OSS bucket name  (matches BMC_OSS_BUCKET env var from admin scripts)
//   endpoint — OSS endpoint hostname   (matches BMC_OSS_ENDPOINT, e.g. oss-cn-hangzhou.aliyuncs.com)
//
// Override the defaults by adding buildConfigField entries in app/build.gradle.kts:
//   buildConfigField("String", "BMC_OSS_BUCKET",   "\"my-prod-bucket\"")
//   buildConfigField("String", "BMC_OSS_ENDPOINT", "\"oss-cn-shanghai.aliyuncs.com\"")
//
// Resulting URL: https://{bucket}.{endpoint}/bmc.db
// ---------------------------------------------------------------------------

object SyncConfig {
    /** OSS bucket name — change via BuildConfig or edit this default. */
    val bucket: String = try {
        com.bmc.app.BuildConfig::class.java.getField("BMC_OSS_BUCKET").get(null) as String
    } catch (_: Exception) { "bmc-data" }

    /** OSS endpoint — change via BuildConfig or edit this default. */
    val endpoint: String = try {
        com.bmc.app.BuildConfig::class.java.getField("BMC_OSS_ENDPOINT").get(null) as String
    } catch (_: Exception) { "oss-cn-hangzhou.aliyuncs.com" }

    val remoteUrl: String get() = "https://$bucket.$endpoint/bmc.db"
}

object DataSync {

    private val _state = MutableStateFlow<SyncState>(SyncState.Idle)

    /** Observable sync state for UI consumption. */
    val state: StateFlow<SyncState> = _state.asStateFlow()

    /** Reset state to idle (called after auto-dismiss delay). */
    fun resetState() {
        _state.value = SyncState.Idle
    }

    /**
     * Download the latest DB from the remote URL and replace the local copy.
     * Updates [state] so the UI can show progress. Failures are non-fatal —
     * the app continues with local data.
     */
    suspend fun syncIfNeeded(context: Context) {
        _state.value = SyncState.Syncing
        withContext(Dispatchers.IO) {
            try {
                val url = URL(SyncConfig.remoteUrl)
                val connection = url.openConnection() as HttpURLConnection
                connection.connectTimeout = 10_000
                connection.readTimeout = 30_000

                try {
                    if (connection.responseCode !in 200..299) {
                        _state.value = SyncState.Failed
                        return@withContext
                    }

                    val tempFile = File(context.cacheDir, "bmc_download.db")
                    connection.inputStream.use { input ->
                        tempFile.outputStream().use { output ->
                            input.copyTo(output)
                        }
                    }

                    Database.getInstance(context).replaceWith(tempFile)
                    _state.value = SyncState.Success
                } finally {
                    connection.disconnect()
                }
            } catch (e: Exception) {
                e.printStackTrace()
                _state.value = SyncState.Failed
            }
        }
    }
}

package com.bmc.app.data

import android.content.Context
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateMapOf
import androidx.compose.runtime.snapshots.SnapshotStateList
import androidx.compose.runtime.snapshots.SnapshotStateMap
import org.json.JSONArray
import org.json.JSONObject
import java.io.File

/**
 * Persists user-specific state (favorites, path progress) as JSON in app-internal storage.
 * Exposes Compose-observable collections so UI recomposes on changes.
 */
class UserState private constructor(context: Context) {

    private val file: File = File(context.applicationContext.filesDir, "user_state.json")

    /** Ordered list of favorited content IDs. */
    val favorites: SnapshotStateList<Int> = mutableStateListOf()

    /** path_id → set of completed step IDs. */
    val pathProgress: SnapshotStateMap<Int, MutableSet<Int>> = mutableStateMapOf()

    init {
        load()
    }

    // -- Favorites -------------------------------------------------------------

    fun isFavorite(contentId: Int): Boolean = contentId in favorites

    fun toggleFavorite(contentId: Int) {
        if (contentId in favorites) {
            favorites.remove(contentId)
        } else {
            favorites.add(contentId)
        }
        save()
    }

    // -- Path progress ---------------------------------------------------------

    fun isStepCompleted(pathId: Int, stepId: Int): Boolean {
        return pathProgress[pathId]?.contains(stepId) == true
    }

    fun toggleStepCompleted(pathId: Int, stepId: Int) {
        val steps = pathProgress.getOrPut(pathId) { mutableSetOf() }
        if (stepId in steps) {
            steps.remove(stepId)
        } else {
            steps.add(stepId)
        }
        // Trigger recomposition by replacing the entry
        pathProgress[pathId] = steps.toMutableSet()
        save()
    }

    // -- Persistence -----------------------------------------------------------

    private fun load() {
        if (!file.exists()) return
        try {
            val json = JSONObject(file.readText())

            // favorites
            val favArray = json.optJSONArray("favorites")
            if (favArray != null) {
                for (i in 0 until favArray.length()) {
                    favorites.add(favArray.getInt(i))
                }
            }

            // pathProgress
            val progressObj = json.optJSONObject("pathProgress")
            if (progressObj != null) {
                for (key in progressObj.keys()) {
                    val stepsArray = progressObj.getJSONArray(key)
                    val stepSet = mutableSetOf<Int>()
                    for (i in 0 until stepsArray.length()) {
                        stepSet.add(stepsArray.getInt(i))
                    }
                    pathProgress[key.toInt()] = stepSet
                }
            }
        } catch (e: Exception) {
            e.printStackTrace()
        }
    }

    private fun save() {
        try {
            val json = JSONObject()

            // favorites
            json.put("favorites", JSONArray(favorites.toList()))

            // pathProgress
            val progressObj = JSONObject()
            for ((pathId, steps) in pathProgress) {
                progressObj.put(pathId.toString(), JSONArray(steps.toList()))
            }
            json.put("pathProgress", progressObj)

            file.writeText(json.toString(2))
        } catch (e: Exception) {
            e.printStackTrace()
        }
    }

    // -- Singleton -------------------------------------------------------------

    companion object {
        @Volatile
        private var instance: UserState? = null

        fun getInstance(context: Context): UserState {
            return instance ?: synchronized(this) {
                instance ?: UserState(context).also { instance = it }
            }
        }
    }
}

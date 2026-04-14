package com.bmc.app.data

import android.content.Context
import android.database.sqlite.SQLiteDatabase
import com.bmc.app.models.Category
import com.bmc.app.models.ContentItem
import java.io.File
import java.io.FileOutputStream

class Database private constructor(context: Context) {

    private val appContext = context.applicationContext
    private val dbName = "bmc.db"
    private val dbFile: File get() = appContext.getDatabasePath(dbName)
    private var db: SQLiteDatabase? = null

    init {
        copyBundledDbIfNeeded()
        openDatabase()
    }

    // -- Setup ----------------------------------------------------------------

    private fun copyBundledDbIfNeeded() {
        if (dbFile.exists()) return

        dbFile.parentFile?.mkdirs()

        appContext.assets.open(dbName).use { input ->
            FileOutputStream(dbFile).use { output ->
                input.copyTo(output)
            }
        }
    }

    private fun openDatabase() {
        db = SQLiteDatabase.openDatabase(
            dbFile.path,
            null,
            SQLiteDatabase.OPEN_READONLY
        )
    }

    private fun closeDatabase() {
        db?.close()
        db = null
    }

    // -- Queries --------------------------------------------------------------

    fun categories(parentId: Int?): List<Category> {
        val results = mutableListOf<Category>()
        val database = db ?: return results

        val cursor = if (parentId != null) {
            database.rawQuery(
                "SELECT id, name, icon, sort_order, parent_id FROM categories WHERE parent_id = ? ORDER BY sort_order",
                arrayOf(parentId.toString())
            )
        } else {
            database.rawQuery(
                "SELECT id, name, icon, sort_order, parent_id FROM categories WHERE parent_id IS NULL ORDER BY sort_order",
                null
            )
        }

        cursor.use { c ->
            while (c.moveToNext()) {
                val pid = if (c.isNull(4)) null else c.getInt(4)
                results.add(
                    Category(
                        id = c.getInt(0),
                        name = c.getString(1),
                        icon = c.getString(2),
                        sortOrder = c.getInt(3),
                        parentId = pid
                    )
                )
            }
        }
        return results
    }

    fun contents(categoryId: Int): List<ContentItem> {
        val results = mutableListOf<ContentItem>()
        val database = db ?: return results

        val cursor = database.rawQuery(
            "SELECT id, title, summary, thumbnail_url, source_url, source_platform, author_name, category_id, sort_order FROM contents WHERE category_id = ? ORDER BY sort_order",
            arrayOf(categoryId.toString())
        )

        cursor.use { c ->
            while (c.moveToNext()) {
                results.add(
                    ContentItem(
                        id = c.getInt(0),
                        title = c.getString(1),
                        summary = c.getString(2),
                        thumbnailUrl = c.getString(3),
                        sourceUrl = c.getString(4),
                        sourcePlatform = c.getString(5),
                        authorName = c.getString(6),
                        categoryId = c.getInt(7),
                        sortOrder = c.getInt(8)
                    )
                )
            }
        }
        return results
    }

    fun searchContents(keyword: String): List<ContentItem> {
        val results = mutableListOf<ContentItem>()
        if (keyword.isBlank()) return results
        val database = db ?: return results

        val pattern = "%$keyword%"
        val cursor = database.rawQuery(
            "SELECT id, title, summary, thumbnail_url, source_url, source_platform, author_name, category_id, sort_order FROM contents WHERE title LIKE ? OR summary LIKE ? OR author_name LIKE ? ORDER BY sort_order",
            arrayOf(pattern, pattern, pattern)
        )

        cursor.use { c ->
            while (c.moveToNext()) {
                results.add(
                    ContentItem(
                        id = c.getInt(0),
                        title = c.getString(1),
                        summary = c.getString(2),
                        thumbnailUrl = c.getString(3),
                        sourceUrl = c.getString(4),
                        sourcePlatform = c.getString(5),
                        authorName = c.getString(6),
                        categoryId = c.getInt(7),
                        sortOrder = c.getInt(8)
                    )
                )
            }
        }
        return results
    }

    // -- Replace DB (for sync) ------------------------------------------------

    fun replaceWith(downloadedFile: File) {
        closeDatabase()

        try {
            if (dbFile.exists()) {
                dbFile.delete()
            }
            downloadedFile.copyTo(dbFile, overwrite = true)
            downloadedFile.delete()
        } catch (e: Exception) {
            e.printStackTrace()
        }

        openDatabase()
    }

    // -- Singleton ------------------------------------------------------------

    companion object {
        @Volatile
        private var instance: Database? = null

        fun getInstance(context: Context): Database {
            return instance ?: synchronized(this) {
                instance ?: Database(context).also { instance = it }
            }
        }
    }
}

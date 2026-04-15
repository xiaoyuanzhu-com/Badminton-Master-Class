package com.bmc.app.models

data class ContentItem(
    val id: Int,
    val title: String,
    val summary: String,
    val thumbnailUrl: String,
    val sourceUrl: String,
    val sourcePlatform: String,
    val authorName: String,
    val difficulty: String,
    val duration: String,
    val editorNotes: String,
    val categoryId: Int,
    val sortOrder: Int,
    val categoryName: String = ""
)

package com.bmc.app.models

data class ContentItem(
    val id: Int,
    val title: String,
    val summary: String,
    val thumbnailUrl: String,
    val sourceUrl: String,
    val sourcePlatform: String,
    val authorName: String,
    val categoryId: Int,
    val sortOrder: Int
)

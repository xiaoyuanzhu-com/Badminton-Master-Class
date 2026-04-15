package com.bmc.app.models

data class Category(
    val id: Int,
    val name: String,
    val icon: String,
    val sortOrder: Int,
    val parentId: Int?,
    val contentCount: Int = 0
)

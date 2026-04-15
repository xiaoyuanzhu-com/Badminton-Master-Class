package com.bmc.app.models

data class LearningPath(
    val id: Int,
    val title: String,
    val summary: String,
    val difficulty: String,
    val sortOrder: Int,
    val stepCount: Int = 0
)

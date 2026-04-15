package com.bmc.app.models

data class PathStep(
    val id: Int,
    val pathId: Int,
    val stepOrder: Int,
    val day: String,
    val title: String,
    val note: String
)

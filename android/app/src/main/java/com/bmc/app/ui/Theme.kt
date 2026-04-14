package com.bmc.app.ui

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

// DESIGN.md — Ink Black monochromatic palette
val InkBlack = Color(0xFF111111)
val CanvasWhite = Color(0xFFFFFFFF)
val Snow = Color(0xFFFAFAFA)
val LightGray = Color(0xFFF5F5F5)
val PressGray = Color(0xFFE5E5E5)
val SecondaryText = Color(0xFF707072)
val DisabledText = Color(0xFF9E9EA0)
val DividerColor = Color(0xFFCACACB)
val DarkSurface = Color(0xFF28282A)
val DeepCharcoal = Color(0xFF1F1F21)
val DarkPress = Color(0xFF39393B)
val ErrorRed = Color(0xFFD30005)
val SuccessGreen = Color(0xFF007D48)

private val BMCLightColorScheme = lightColorScheme(
    primary = InkBlack,
    onPrimary = CanvasWhite,
    primaryContainer = CanvasWhite,
    onPrimaryContainer = InkBlack,
    secondary = SecondaryText,
    onSecondary = CanvasWhite,
    secondaryContainer = LightGray,
    onSecondaryContainer = InkBlack,
    surface = CanvasWhite,
    onSurface = InkBlack,
    surfaceVariant = LightGray,
    onSurfaceVariant = SecondaryText,
    background = CanvasWhite,
    onBackground = InkBlack,
    error = ErrorRed,
    onError = CanvasWhite,
    outline = DividerColor,
    outlineVariant = PressGray,
)

private val BMCDarkColorScheme = darkColorScheme(
    primary = CanvasWhite,
    onPrimary = InkBlack,
    primaryContainer = DarkSurface,
    onPrimaryContainer = CanvasWhite,
    secondary = SecondaryText,
    onSecondary = InkBlack,
    secondaryContainer = DarkPress,
    onSecondaryContainer = CanvasWhite,
    surface = DeepCharcoal,
    onSurface = CanvasWhite,
    surfaceVariant = DarkSurface,
    onSurfaceVariant = SecondaryText,
    background = DeepCharcoal,
    onBackground = CanvasWhite,
    error = ErrorRed,
    onError = CanvasWhite,
    outline = DarkPress,
    outlineVariant = DarkSurface,
)

@Composable
fun BMCTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit
) {
    val colorScheme = if (darkTheme) BMCDarkColorScheme else BMCLightColorScheme

    MaterialTheme(
        colorScheme = colorScheme,
        content = content
    )
}

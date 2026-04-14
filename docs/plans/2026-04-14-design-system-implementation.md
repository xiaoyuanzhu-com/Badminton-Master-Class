# Design System Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Apply the Nike-inspired DESIGN.md to both iOS (SwiftUI) and Android (Jetpack Compose) native apps.

**Architecture:** The DESIGN.md defines an Ink Black (#111111) monochromatic UI with platform badge colors as the only chromatic accents. Android needs a custom Material 3 color scheme replacing dynamic colors. iOS needs exact hex badge colors and minor tint adjustments. Both platforms already have correct spacing and layout — this is primarily a color/theme pass.

**Tech Stack:** SwiftUI (iOS), Jetpack Compose + Material 3 (Android)

---

### Task 1: Android — Custom Color Scheme in Theme.kt

**Files:**
- Modify: `android/app/src/main/java/com/bmc/app/ui/Theme.kt`

**Step 1: Replace stock theme with DESIGN.md colors**

Replace the entire file with a custom color scheme:

```kotlin
package com.bmc.app.ui

import android.os.Build
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

// DESIGN.md — Ink Black monochromatic palette
private val InkBlack = Color(0xFF111111)
private val CanvasWhite = Color(0xFFFFFFFF)
private val Snow = Color(0xFFFAFAFA)
private val LightGray = Color(0xFFF5F5F5)
private val PressGray = Color(0xFFE5E5E5)
private val SecondaryText = Color(0xFF707072)
private val DisabledText = Color(0xFF9E9EA0)
private val Divider = Color(0xFFCACACB)
private val DarkSurface = Color(0xFF28282A)
private val DeepCharcoal = Color(0xFF1F1F21)
private val DarkPress = Color(0xFF39393B)
private val ErrorRed = Color(0xFFD30005)
private val SuccessGreen = Color(0xFF007D48)

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
    outline = Divider,
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
```

Key changes:
- Removed dynamic color — the app now has a consistent monochromatic identity on all devices
- Removed `dynamicColor` parameter — no longer needed
- Light scheme: white surfaces, Ink Black text, LightGray for surfaceVariant
- Dark scheme: DeepCharcoal surfaces, white text, DarkSurface for containers
- `outline` maps to Divider color (#CACACB) for HorizontalDivider

**Step 2: Update BMCTheme call site if `dynamicColor` was passed**

Check `MainActivity.kt` or wherever BMCTheme is called — remove any `dynamicColor = true` argument.

**Step 3: Commit**

```bash
git add android/app/src/main/java/com/bmc/app/ui/Theme.kt
git commit -m "feat(android): apply DESIGN.md color scheme to BMCTheme"
```

---

### Task 2: Android — TopAppBar Styling

**Files:**
- Modify: `android/app/src/main/java/com/bmc/app/ui/HomeScreen.kt`
- Modify: `android/app/src/main/java/com/bmc/app/ui/CategoryScreen.kt`

**Step 1: Update HomeScreen TopAppBar**

The TopAppBar currently uses `primaryContainer` / `onPrimaryContainer`. With the new theme, `primaryContainer` is already `CanvasWhite` and `onPrimaryContainer` is `InkBlack`, so the TopAppBar will automatically be white with black text. This is correct per DESIGN.md.

Verify: no code change needed — the theme handles it. If the TopAppBar renders correctly with the new theme, skip to Step 2.

**Step 2: Update search bar to use filled style**

Replace the `OutlinedTextField` in HomeScreen with a `TextField` (filled style) for a cleaner look matching DESIGN.md's search bar spec (#F5F5F5 background, no visible border):

In HomeScreen.kt, replace the OutlinedTextField block:

```kotlin
TextField(
    value = searchQuery,
    onValueChange = { searchQuery = it },
    modifier = Modifier
        .fillMaxWidth()
        .padding(horizontal = 16.dp, vertical = 8.dp),
    placeholder = { Text("搜索教程") },
    leadingIcon = {
        Icon(Icons.Default.Search, contentDescription = "搜索")
    },
    trailingIcon = {
        if (searchQuery.isNotEmpty()) {
            IconButton(onClick = { searchQuery = "" }) {
                Icon(Icons.Default.Clear, contentDescription = "清除")
            }
        }
    },
    singleLine = true,
    shape = RoundedCornerShape(12.dp),
    colors = TextFieldDefaults.colors(
        focusedContainerColor = LightGray,
        unfocusedContainerColor = LightGray,
        focusedIndicatorColor = Color.Transparent,
        unfocusedIndicatorColor = Color.Transparent,
    )
)
```

Add imports at top of HomeScreen.kt:
```kotlin
import androidx.compose.material3.TextField
import androidx.compose.material3.TextFieldDefaults
import androidx.compose.ui.graphics.Color
```

Remove the `OutlinedTextField` import.

Also add the LightGray constant at file level or import from Theme:
```kotlin
private val LightGray = Color(0xFFF5F5F5)
```

**Step 3: Commit**

```bash
git add android/app/src/main/java/com/bmc/app/ui/HomeScreen.kt
git commit -m "feat(android): search bar uses filled style with #F5F5F5 background"
```

---

### Task 3: Android — Sync Failed State Color

**Files:**
- Modify: `android/app/src/main/java/com/bmc/app/ui/HomeScreen.kt`

**Step 1: Add error color to sync failed state**

In the `SyncStatusBar` composable, the "同步失败" text currently uses `onSurfaceVariant`. Per DESIGN.md, failed state should use Error Red.

Change:
```kotlin
is SyncState.Failed -> {
    Text(
        text = "同步失败",
        style = MaterialTheme.typography.labelSmall,
        color = MaterialTheme.colorScheme.error
    )
}
```

**Step 2: Commit**

```bash
git add android/app/src/main/java/com/bmc/app/ui/HomeScreen.kt
git commit -m "fix(android): sync failed text uses error red per DESIGN.md"
```

---

### Task 4: iOS — Exact Platform Badge Hex Colors

**Files:**
- Modify: `ios/BadmintonMasterClass/CategoryView.swift`

**Step 1: Replace system colors with exact DESIGN.md hex values**

The iOS PlatformBadge currently uses SwiftUI system colors (`.pink`, `.red`, `.black`, `.green`) which don't match the exact hex values in DESIGN.md. Replace with exact colors:

```swift
private var badgeColor: Color {
    switch platform {
    case "bilibili": return Color(red: 0xFB/255, green: 0x72/255, blue: 0x99/255)   // #FB7299
    case "xiaohongshu": return Color(red: 0xFF/255, green: 0x24/255, blue: 0x42/255) // #FF2442
    case "douyin": return Color(red: 0x16/255, green: 0x18/255, blue: 0x23/255)      // #161823
    case "wechat": return Color(red: 0x07/255, green: 0xC1/255, blue: 0x60/255)      // #07C160
    case "youtube": return Color(red: 0xFF/255, green: 0x00/255, blue: 0x00/255)     // #FF0000
    default: return .gray
    }
}
```

**Step 2: Increase badge horizontal padding from 6 to 8**

DESIGN.md specifies 8dp horizontal padding for badges. Change in the PlatformBadge body:

```swift
.padding(.horizontal, 8)
```

**Step 3: Commit**

```bash
git add ios/BadmintonMasterClass/CategoryView.swift
git commit -m "feat(ios): exact DESIGN.md hex colors for platform badges"
```

---

### Task 5: iOS — Sync Failed State Color

**Files:**
- Modify: `ios/BadmintonMasterClass/HomeView.swift`

**Step 1: Use red for sync failed state**

In the SyncStatusBar, the `.failed` case currently uses `.secondary` foreground. Per DESIGN.md, it should use Error Red:

```swift
case .failed:
    Text("同步失败")
        .font(.caption)
        .foregroundStyle(.red)
        .frame(maxWidth: .infinity)
        .padding(.vertical, 6)
        .background(.ultraThinMaterial)
        .transition(.opacity)
```

**Step 2: Commit**

```bash
git add ios/BadmintonMasterClass/HomeView.swift
git commit -m "fix(ios): sync failed text uses red per DESIGN.md"
```

---

### Task 6: Verify & Final Commit

**Step 1: Review all changes**

```bash
git log --oneline -5
git diff HEAD~5..HEAD --stat
```

Verify:
- Android Theme.kt: custom Ink Black color scheme, no dynamic colors
- Android HomeScreen: filled search bar (#F5F5F5), error red on sync failure
- iOS CategoryView: exact hex platform badge colors, 8pt horizontal padding
- iOS HomeView: red sync failure text

**Step 2: Build verification (if possible)**

```bash
# Android
cd android && ./gradlew assembleDebug

# iOS
cd ios && xcodebuild -scheme BadmintonMasterClass -destination 'platform=iOS Simulator,name=iPhone 16' build
```

Note: builds require Mac mini with Android SDK and Xcode. If not available locally, mark for remote build.

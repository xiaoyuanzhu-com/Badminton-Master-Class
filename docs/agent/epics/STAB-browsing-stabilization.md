# STAB — Browsing Stabilization

## Meta
- Status: in-progress
- Parent roadmap: RM-4
- Created: 2026-04-18
- Source: Pre-execution audit of basic browsing on iOS + Android (BOSS-requested)

## Goal
Close all visible cracks in basic browsing on iOS and Android **before** expanding product scope into engagement features (HISTORY, FRESH, SHARE). Audit verdict: foundation is architecturally sound but degraded by data gaps, one latent crash, and small inconsistencies. Phase 0 lands as a tight bundle, then RM-4 product expansion proceeds.

## Audit Findings (2026-04-18)

### iOS
- **Critical:** `Database` not thread-safe — `replaceWith()` on MainActor races concurrent queries (`PathDetailView` uses `withTaskGroup`). Crash risk on pull-to-refresh during path browsing.
- **Critical:** 12/20 content items have empty `thumbnail_url` — gray placeholders dominate.
- **Critical:** 6/20 subcategories empty (劈吊, 抽挡, 挡网, 发接发, 柔韧性, 扑球) — 30% of category drill-downs are dead ends.
- **Polish:** No progress bar in `PathDetailView` (only on home cards).
- **Polish:** Three duplicate `DifficultyBadge` implementations (HomeView, CategoryView, PathDetailView).
- **Newly surfaced:** `contents()` and `pathStepContents()` queries don't JOIN categories — `categoryName` badge only appears in search results, never in browse.
- **Newly surfaced:** `HomeView.refreshable` reloads categories + paths but not favorites — stale favorite rows after sync.

### Android
- **Critical:** `LearningPathCard` (HomeScreen.kt:429) and `SearchPathRow` (HomeScreen.kt:494) render raw `path.difficulty` ("beginner") instead of localized "入门".
- **Critical:** Same 12/20 thumbnails missing, same 6 empty subcategories.
- **Polish:** No aggregate progress bar in `PathDetailScreen` (only per-step toggles).
- **Polish:** `UserState` saves synchronously on every tap — main-thread file I/O.
- **Polish:** `SyncConfig` uses reflection to read BuildConfig fields with blanket exception catch.
- **Newly surfaced:** First-launch shows "暂无内容" empty state during DB asset copy instead of a loading spinner.
- **Newly surfaced:** Build never verified on device — Mac mini has no JDK (RM-3 marked it BLOCKED).

### Both platforms
- **Content:** Zero items have `duration` or `editor_notes` populated — UI features are dead code for the entire dataset.
- **Web:** `search.html` uses Google blue header instead of Ink Black.

## Tasks

### 🔲 STAB-1: iOS Database thread safety
- Confidence: 🟢 High
- Change size: Medium (1-2 files, ~50 lines)
- Approach: Convert `Database` to a Swift actor, OR route `replaceWith` through `queryQueue` with a write lock
- File: `ios/BadmintonMasterClass/Database.swift`

### 🔲 STAB-2: Verify thread-safety fix
- Confidence: 🟢 High
- Change size: Small (test only)
- Scenario: concurrent sync + path detail load

### 🔲 STAB-3: Android path difficulty labels
- Confidence: 🟢 High
- Change size: Small (~10 lines)
- File: `android/app/src/main/java/.../HomeScreen.kt:429,494`
- Use existing `ContentDifficultyBadge` from `CategoryScreen.kt:359-379`

### 🔲 STAB-4: Path detail progress bar (iOS + Android)
- Confidence: 🟢 High
- Change size: Small (~20 lines, 2 files)
- iOS: `PathDetailView.swift:19-22`
- Android: `PathDetailScreen.kt`

### 🔲 STAB-5: Hide empty subcategories (iOS + Android)
- Confidence: 🟢 High
- Change size: Small
- Approach: filter at query layer in `Database` (subcategories with content_count > 0) OR at render layer
- Empty list: 劈吊, 抽挡, 挡网, 发接发, 柔韧性, 扑球

### 🔲 STAB-6: Backfill 12 missing thumbnails
- Confidence: 🟢 High
- Change size: Small (data only)
- Use Bilibili/YouTube/Xiaohongshu thumbnail URLs (already proven stable)

### 🔲 STAB-7: Backfill duration + editor_notes
- Confidence: 🟡 Medium (content quality matters)
- Change size: Medium
- 20 items × (duration + editor_notes) — needs human-quality editorial work

### 🔲 STAB-8: iOS browse-mode category JOIN
- Confidence: 🟢 High
- Change size: Small (~5 lines SQL)
- File: `ios/BadmintonMasterClass/Database.swift:105` (`contents` query) + path step contents query

### 🔲 STAB-9: iOS reload favorites on refresh
- Confidence: 🟢 High
- Change size: Small (~1 line)
- File: `ios/BadmintonMasterClass/Views/HomeView.swift:147-153`

### 🔲 STAB-10: Android first-launch spinner
- Confidence: 🟢 High
- Change size: Small
- Distinguish "loading" from "empty" state in `HomeScreen.kt`

### 🔲 STAB-11: Install JDK 17 on Mac mini + verify Android build
- Confidence: 🟢 High
- Change size: Small (infra)
- Command: `brew install openjdk@17` on macmini
- Then run `./gradlew assembleDebug` and verify

### 🔲 STAB-12: Web search.html header color
- Confidence: 🟢 High
- Change size: Small (CSS)

## Key Decisions

### Why Phase 0 instead of accepting RM-4 as drafted
- **Chose:** Pre-bundle the visible cracks before product expansion.
- **Reasoning:** Layering HISTORY/FRESH/SHARE on top of an app where 60% of thumbnails are gray and 30% of categories are dead ends would result in user testing where the new features get blamed for foundational problems. Cleaner signal if Phase 0 lands first.
- **Confidence:** 🟢 High — BOSS-validated direction.

### Why not parallelize Phase 0 across multiple agents
- **Chose:** Sequential execution per roadmap-diffusion skill ("each task lands on `main` before the next task begins").
- **Reasoning:** Some tasks touch the same files (e.g., STAB-4 + STAB-9 both edit iOS views). Sequential commits keep history clean and avoid worktree conflicts.
- **Confidence:** 🟢 High.

## Execution Summary
(populated as tasks complete)

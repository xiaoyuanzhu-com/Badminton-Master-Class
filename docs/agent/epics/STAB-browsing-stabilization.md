# STAB — Browsing Stabilization

## Meta
- Status: done (Phase 0 closed 2026-04-18)
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

### ✅ STAB-1: iOS Database thread safety — `17fa817`
- Confidence: 🟢 High
- Change size: Medium (1-2 files, ~50 lines)
- Approach: Convert `Database` to a Swift actor, OR route `replaceWith` through `queryQueue` with a write lock
- File: `ios/BadmintonMasterClass/Database.swift`

### ✅ STAB-2: Verify thread-safety fix — `17fa817`
- Confidence: 🟢 High
- Change size: Small (test only)
- Scenario: concurrent sync + path detail load

### ✅ STAB-3: Android path difficulty labels — `61e0a7b`
- Confidence: 🟢 High
- Change size: Small (~10 lines)
- File: `android/app/src/main/java/.../HomeScreen.kt:429,494`
- Use existing `ContentDifficultyBadge` from `CategoryScreen.kt:359-379`

### ✅ STAB-4: Path detail progress bar (iOS + Android) — `828efae`
- Confidence: 🟢 High
- Change size: Small (~20 lines, 2 files)
- iOS: `PathDetailView.swift:19-22`
- Android: `PathDetailScreen.kt`

### ✅ STAB-5: Hide empty subcategories (iOS + Android) — `13be107`
- Confidence: 🟢 High
- Change size: Small
- Approach: filter at query layer in `Database` (subcategories with content_count > 0) OR at render layer
- Empty list: 劈吊, 抽挡, 挡网, 发接发, 柔韧性, 扑球

### ✅ STAB-6: Backfill 12 missing thumbnails — `89c32c2`
- Confidence: 🟢 High
- Change size: Small (data only)
- Use Bilibili/YouTube/Xiaohongshu thumbnail URLs (already proven stable)

### ⏭️ STAB-7: Backfill duration + editor_notes — DEFERRED → RM-4 GROW EPIC
- Confidence: 🟡 Medium (content quality matters)
- Change size: Medium
- 20 items × (duration + editor_notes) — needs human-quality editorial work
- **Rationale:** Backfilling `duration` and `editor_notes` for the 20 existing items is content work that belongs in GROW, not in stabilization. The fields' absence hides UI rows; it does not break browsing. Surfaced during STAB-6: the source-of-truth content pipeline (`data/build.py`) is broken post the RM-3 taxonomy restructure (`47c650f`) — fixing it before any large-scale content work is now a GROW prerequisite.

### ✅ STAB-8: iOS browse-mode category JOIN — `ad18a27`
- Confidence: 🟢 High
- Change size: Small (~5 lines SQL)
- File: `ios/BadmintonMasterClass/Database.swift:105` (`contents` query) + path step contents query

### ✅ STAB-9: iOS reload favorites on refresh — `ad18a27`
- Confidence: 🟢 High
- Change size: Small (~1 line)
- File: `ios/BadmintonMasterClass/Views/HomeView.swift:147-153`

### ✅ STAB-10: Android first-launch spinner — `f8ba193`
- Confidence: 🟢 High
- Change size: Small
- Distinguish "loading" from "empty" state in `HomeScreen.kt`

### ✅ STAB-11: Install JDK 17 on Mac mini + verify Android build — `c082eb3`
- Confidence: 🟢 High
- Change size: Small (infra)
- PARTIAL — wrapper landed (`c082eb3`), full build verified via STAB-14 (`641832b`)

### ✅ STAB-12: Web search.html header color — `f8ba193`
- Confidence: 🟢 High
- Change size: Small (CSS)

### ✅ STAB-13: Android placeholder app icon — `f550b84`
- Confidence: 🟢 High
- Change size: Small (mipmap resources)
- Add placeholder `ic_launcher` mipmap drawables so the build does not fail on missing icon resources

### ✅ STAB-14: Android Kotlin compile blockers — `641832b`
- Confidence: 🟢 High
- Change size: Small
- Enable `BuildConfig` generation and add `material-icons-extended` dependency to unblock `assembleDebug`

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

11 tasks shipped across 9 commits; 1 task deferred to GROW.

| Commit | Description |
|--------|-------------|
| `17fa817` | STAB-1/2: iOS Database actor — eliminate thread-safety crash on pull-to-refresh |
| `828efae` | STAB-4: Aggregate progress bar on path detail screen (iOS + Android) |
| `61e0a7b` | STAB-3: Localize Android path difficulty labels (入门/进阶/精通) |
| `13be107` | STAB-5: Hide empty subcategories at query layer (iOS + Android) |
| `ad18a27` | STAB-8/9: iOS browse category JOIN + reload favorites on refresh |
| `f8ba193` | STAB-10/12: Android first-launch spinner + web search.html header Ink Black |
| `c082eb3` | STAB-11 (partial): Commit gradle wrapper (gradlew + gradle-wrapper.jar) |
| `4ff1dcc` | STAB-13 prep: Adaptive icon resources scaffolding |
| `f550b84` | STAB-13: Android placeholder app icon (mipmap/ic_launcher) |
| `641832b` | STAB-14: Enable BuildConfig + add material-icons-extended (unblocks assembleDebug) |
| `89c32c2` | STAB-6: Backfill thumbnail_url for 12 content items |

**Deferred:** STAB-7 (duration + editor_notes backfill) → RM-4 GROW EPIC. Prerequisite: fix `data/build.py` (GROW-0).

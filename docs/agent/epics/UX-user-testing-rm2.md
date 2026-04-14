# UX User Testing Report — RM-2 (Badminton Master Class)

**Date:** 2026-04-14
**Scope:** Content-as-code workflow, compiler, read-only web client, iOS app, Android app
**Persona:** Badminton enthusiast (app user) + content curator (managing content files)

---

## RM-1 Issues Resolved by RM-2

### Critical Issues — All 4 Resolved

| # | RM-1 Issue | RM-2 Resolution | Status |
|---|-----------|-----------------|--------|
| 1 | Android: No back navigation in CategoryScreen | `CategoryScreen.kt` now accepts `onBack` callback; `TopAppBar` has `navigationIcon` with `Icons.AutoMirrored.Filled.ArrowBack`. `BMCApp.kt` passes `navController.popBackStack()`. | **Resolved** |
| 2 | Android: Thumbnails never load (hardcoded placeholder) | Coil dependency added (`io.coil-kt:coil-compose:2.6.0` in `build.gradle.kts`). `ContentThumbnail` composable uses `AsyncImage` from Coil when `thumbnailUrl` is non-empty, falls back to placeholder. | **Resolved** |
| 3 | Sync always downloads full DB on every launch | Both platforms now implement ETag-based conditional sync. iOS stores ETag in `UserDefaults`, sends `If-None-Match`, handles `304 Not Modified`. Android uses `SharedPreferences` for ETag storage with the same logic. | **Resolved** |
| 4 | iOS: Database queries on main thread | iOS `Database.swift` now has a private `DispatchQueue` (`com.bmc.db.query`) with `userInitiated` QoS. Async wrappers (`categoriesAsync`, `contentsAsync`, `searchContentsAsync`) dispatch queries off-main. `HomeView` and `CategoryView` use `.task` and `await` instead of `onAppear`. | **Resolved** |

### Important Issues (5-13)

| # | RM-1 Issue | RM-2 Status | Notes |
|---|-----------|-------------|-------|
| 5 | No onboarding/first-launch explanation | **Not addressed** | Still no onboarding. |
| 6 | No indication content opens externally | **Partially addressed** | Deep linking now opens native apps when installed, making the transition smoother. But there is still no visual cue on the content row itself. |
| 7 | Search only queries content, not categories | **Partially addressed** — web client searches both content and people, but mobile apps still only search content titles/summaries/authors. Categories are not searchable on mobile. | |
| 8 | Search results lack category context | **Resolved on web** — the web search results show category name badges on each content item. **Not resolved on mobile** — `ContentRow` on iOS/Android still shows only title, summary, platform badge, and author. No category name. | |
| 9 | iOS: Sync status auto-dismisses too quickly | **Not addressed** | Same 2s/3s timers in `SyncManager.swift`. |
| 10 | Android: Pull-to-refresh naming misleading | **Partially addressed** | The sync is now truly conditional (ETag), so `syncIfNeeded` is more honestly named. However, the pull-to-refresh still calls `syncIfNeeded` rather than a forced sync, meaning pull-to-refresh on a 304 response just shows "already synced" with no way to force a full re-download. |
| 11 | Admin: Category dropdown flat, not hierarchical | **N/A** | Admin panel has been converted to a read-only web client. No more add/edit forms. Content management happens through content-as-code files. |
| 12 | Admin: No validation feedback on form errors | **N/A** | Admin panel is now read-only. Content validation is handled by `validate.py` in the content-as-code pipeline. |
| 13 | Android: Category name in route can break with special characters | **Not addressed** | `BMCApp.kt` line 42 still passes `category.name` directly in the navigation route: `"category/${category.id}/${category.name}"`. Names with `/`, `?`, or `#` will break navigation. Chinese characters work in practice, but the architecture is fragile. |

### Nice-to-Have Issues (14-20)

| # | RM-1 Issue | RM-2 Status |
|---|-----------|-------------|
| 14 | No content count on categories | **Resolved on web** — home page shows "N 个内容" per category. **Not on mobile.** |
| 15 | No loading state on first launch | **Not addressed** — still shows empty state briefly. |
| 16 | Seed data has no thumbnail URLs | **By design** — `build.py` comment says thumbnail_url is intentionally empty in the DB for now. Thumbnails stored as files in repo will be served from CDN in future. |
| 17 | iOS: Search on every keystroke (no debounce) | **Resolved** — `HomeView.swift` now uses `Task.sleep(nanoseconds: 300_000_000)` with cancellation for 300ms debounce. |
| 18 | Admin: Export triggers OSS upload synchronously | **N/A** — admin panel no longer has export. Upload is handled by `make upload` via `build.py`. |
| 19 | No dark mode-specific styling on iOS | **Not addressed** |
| 20 | Admin panel no mobile responsiveness | **Resolved** — web client uses responsive `max-width: 960px` with grid layouts (`grid-template-columns: repeat(auto-fill, minmax(280px, 1fr))`) that adapt to mobile. |

**Summary: 4/4 critical resolved. 3/9 important resolved or made N/A. 3/7 nice-to-have resolved or N/A.**

---

## New Issues Found in RM-2

### Critical

*None identified.* The RM-2 changes are architecturally sound.

### Important

#### N1. Android search has no debounce

**Severity:** Important
**File:** `android/app/src/main/java/com/bmc/app/ui/HomeScreen.kt`, lines 71-77

iOS now has 300ms debounce via `Task.sleep`. Android's `LaunchedEffect(searchQuery)` fires immediately on every character change. `LaunchedEffect` does cancel the previous coroutine when `searchQuery` changes, so rapid typing will not accumulate results, but each keystroke still triggers a DB query before it gets cancelled by the next. Should add `delay(300)` at the start of the `LaunchedEffect` body.

#### N2. Android Database queries run on the calling thread

**Severity:** Important
**File:** `android/app/src/main/java/com/bmc/app/data/Database.kt`

iOS moved queries to a background `DispatchQueue`. Android's `Database.kt` has no threading mechanism — `categories()`, `contents()`, and `searchContents()` execute raw SQL on whatever thread calls them. Currently they are called from `LaunchedEffect` blocks which run on the main dispatcher by default. The database is opened as `OPEN_READONLY`, but with a growing content library this will cause jank. Should use `withContext(Dispatchers.IO)` in callers or provide `suspend` wrappers.

#### N3. Content-as-code: person field is required in schema but curator might not know the author

**Severity:** Important
**File:** `data/content/schemas/content.schema.json`, line 8

The `content.schema.json` has `"person"` in the `"required"` array. This means every content item must reference a person slug. If a curator wants to add a video but doesn't know the author, they cannot add it without first creating a person file. The `build.py` compiler handles missing persons gracefully (lines 177-178: empty person_slug just sets person_id to None), but `validate.py` will reject the file because the schema requires it. The schema and compiler are inconsistent.

#### N4. Deep link: b23.tv short links won't parse correctly

**Severity:** Important
**Files:** `ios/BadmintonMasterClass/DeepLink.swift` line 51, `android/app/src/main/java/com/bmc/app/util/DeepLink.kt` line 55

Both platforms check `host.contains("b23.tv")` for Bilibili, but then require `pathComponents[1] == "video"`. A b23.tv short link (e.g., `https://b23.tv/abc123`) is a redirect — the actual URL path is just the short code, not `/video/BVxxx`. The deep link extraction will return nil for b23.tv links, falling back to the web view. This is not a crash, but the deep link feature silently fails for the most common Bilibili sharing format.

#### N5. Xiaohongshu deep link assumes /explore/ path only

**Severity:** Important
**Files:** `ios/BadmintonMasterClass/DeepLink.swift` line 85, `android/...DeepLink.kt` line 84

Xiaohongshu content URLs can also use `/discovery/item/` or just `/item/` paths, not only `/explore/`. The deep link logic requires `pathComponents[1] == "explore"`, which means alternative URL formats fall back to the browser. Shared Xiaohongshu links often use `xhslink.com` short URLs too, which are not handled at all.

#### N6. Web client: search results for content lack clickable source links

**Severity:** Important
**File:** `admin/templates/search.html`

The search results page shows content items with title linking to the detail page, platform badge, category badge, and author badge. But there is no direct "watch" or source link. A user must click through to the detail page first, then click the source link. On the home page and contents list, this is fine since it's a browse flow — but in search results, users typically want to go directly to the content.

#### N7. Category sort order determined by filesystem alphabetical order

**Severity:** Important
**File:** `data/build.py` lines 115-118

`build_categories` sorts directories by `entry.name` (filesystem alphabetical order). This means the display order of categories is determined by their English folder names (e.g., `attack`, `basics`, `defense`, `fitness`, `net-play`), not by any curated sort order. The README says "No sort_order — display ordering is app-level logic," but the compiler assigns `sort_order` based on alphabetical folder names. A curator cannot control the order categories appear without renaming folders (e.g., prefixing with numbers).

#### N8. Web client content list has no pagination

**Severity:** Minor (now), Important (at scale)
**File:** `admin/handlers.go` line 227

`contentsHandler` selects all matching contents with no LIMIT. As content grows, the `/contents` page will load every item at once. With 20 items this is fine; with 200+ it will be slow and unwieldy.

### Minor

#### N9. Android: `ContentRow` is defined in `CategoryScreen.kt`, used in `HomeScreen.kt` — shared correctly but unexpected location

**Severity:** Minor
**Files:** `android/...CategoryScreen.kt` (defines `ContentRow` as `internal`), `android/...HomeScreen.kt` (uses it)

`ContentRow`, `ContentThumbnail`, and `PlatformBadge` are defined in `CategoryScreen.kt` but used from both `CategoryScreen` and `HomeScreen`. They are `internal` visibility, so this works, but it is counterintuitive. A developer or curator looking for the shared content row would not expect to find it in `CategoryScreen.kt`.

#### N10. iOS `Database.replaceWith` runs on MainActor — could block UI during file operations

**Severity:** Minor
**File:** `ios/BadmintonMasterClass/DataSync.swift` lines 102-105

`Database.shared.replaceWith(downloadedDBAt:)` is called inside `await MainActor.run {}`. This method closes the database, performs file system operations (remove + move), and reopens the database — all on the main thread. For a small DB this is fast, but it could cause a brief freeze with a larger database.

#### N11. Web client templates duplicate the entire header/nav/CSS in every file

**Severity:** Minor
**Files:** All 7 HTML templates in `admin/templates/`

Every template contains the full `<header>`, `<nav>`, and `<style>` block (100+ lines each). A style or navigation change requires editing all 7 files. Go's `template.ParseFiles` supports `{{ template "header" }}` includes, which would eliminate this duplication.

#### N12. Compiler does not validate source_url uniqueness in the database

**Severity:** Minor
**File:** `data/build.py`

`validate.py` checks for duplicate `source_url` across content files (good). But the `contents` table in the compiled database has no UNIQUE constraint on `source_url`. If a bug in the compiler or a manual DB edit introduces a duplicate, the database allows it silently. The schema should have `UNIQUE(source_url)` as a safety net.

---

## Remaining Gaps

### Content Curator Experience

1. **No ingestion script yet** — The commit messages mention an "ingestion script" (`ingest.py`), but no such file exists in the repository. Curators must manually create JSON files following the schema. This is workable but error-prone for non-technical curators.

2. **No way to preview changes before building** — After editing JSON files, a curator must run `make build` and then either deploy the web client or install the mobile app to see their changes. A `make preview` target that serves the compiled DB locally via the web client would be valuable.

3. **Thumbnail pipeline incomplete** — `build.py` explicitly sets `thumbnail_url = ""` for all content, with a comment about future CDN upload. Every content item in the app shows a gray placeholder. This is the single biggest visual weakness of the app.

### Mobile App

4. **Search still doesn't find categories** — A user searching "步法" (footwork) finds individual videos but not the category itself. The web client's search finds people alongside content, but neither platform searches categories.

5. **No content count on mobile category rows** — The web client shows "N 个内容" per category. Mobile shows only icon + name. Users cannot gauge category depth.

6. **Sync failure has no retry mechanism** — If sync fails (network error, server down), the only option is pull-to-refresh or relaunch the app. There is no "Retry" button on the error state.

### Deep Linking

7. **Short URL formats not handled** — b23.tv (Bilibili), xhslink.com (Xiaohongshu), v.douyin.com (Douyin) are all common sharing formats that use redirects. None are handled by the deep link logic. The current implementation only works with canonical full URLs.

---

## Strengths

### Content-as-Code Architecture

The transition from a mutable admin panel with seed SQL to a content-as-code file system is the standout achievement of RM-2. The architecture is clean:

- **File structure maps directly to the UI hierarchy** — folder nesting = category tree. No IDs, no sort_order in the source files. A curator can understand the taxonomy by browsing the filesystem.
- **Schemas enforce correctness** — JSON Schema validation with both `jsonschema` library support and a manual fallback. The `additionalProperties: false` setting catches typos.
- **Cross-reference validation** — `validate.py` checks that person references resolve to actual files and that no two content items share a URL. This catches real data quality issues.
- **Single build command** — `make build` validates, compiles, and copies to both app bundles in one step. The pipeline is linear and predictable.
- **Clear separation** — content files are pure data (no code), the compiler is a pure function (files in, DB out), and the apps are pure consumers.

### Web Client

The conversion from a CRUD admin panel to a read-only web client is well-executed:

- **Clean, modern design** — Card-based layout with good typography, responsive grid, consistent color scheme. Looks professional without a CSS framework.
- **Complete navigation** — Home (category cards with subcategory chips), categories list, contents list with category filtering, content detail with breadcrumbs, people list, person detail with associated content, and full-text search across content and people.
- **Badges everywhere** — Platform, category, author, difficulty, and duration badges provide information density without clutter.
- **Search is comprehensive** — Searches across content title, summary, author, person name, and bio. Returns both content and people sections.
- **Test coverage** — `handlers_test.go` has 15 tests covering all endpoints, method validation, 404 handling, search, and auth.

### ETag Sync

The conditional sync implementation is correct and complete:

- First launch: no ETag stored, full download, ETag saved.
- Subsequent launches: sends `If-None-Match`, server returns 304 if unchanged, skipping download entirely.
- After content update: server returns 200 with new DB and new ETag.
- The implementation is consistent across iOS (URLSession + UserDefaults) and Android (HttpURLConnection + SharedPreferences).

### Deep Linking

The deep link architecture is well-designed:

- Pure computation functions (`deepLinkURL`/`deepLinkUri`) are separate from side effects (`open`), making them testable.
- Try native app first, fall back to web view gracefully.
- Platform-specific URL parsing handles the canonical URL formats correctly.
- WeChat is explicitly handled as "no deep link available" rather than silently failing.

### Platform Improvements

- **iOS search debounce** with proper task cancellation — 300ms delay, cancels previous task on new input, checks `Task.isCancelled` after sleep and after query.
- **Shared `ContentRow`** on Android — used by both `HomeScreen` and `CategoryScreen`, maintaining visual consistency.
- **Android back button** with proper `AutoMirrored` arrow icon for RTL support.
- **Coil integration** is clean — conditional rendering based on `thumbnailUrl.isNotEmpty()`.

---

## Recommendations for RM-3

### P0 — Fix Before Next Release

1. **Add Android search debounce** — Add `delay(300)` at the start of the search `LaunchedEffect` in `HomeScreen.kt`. One line, matches iOS behavior.
2. **Move Android DB queries off main thread** — Add `withContext(Dispatchers.IO)` in the `LaunchedEffect` blocks that call `Database.getInstance().categories()` etc., or add `suspend` wrappers to `Database.kt`.
3. **URL-encode category name in Android nav route** — Use `Uri.encode(category.name)` when building the route string, or (better) pass only the ID and look up the name from the database.
4. **Handle b23.tv Bilibili short links** — Either resolve the redirect first, or match the b23.tv format separately and open via `bilibili://` with the short code.

### P1 — High Value

5. **Build the ingestion script** — `ingest.py` should accept a URL, auto-detect platform, fetch metadata (title, author, thumbnail), create the person file if needed, and write the content JSON. This is the curator's primary workflow and should be frictionless.
6. **Implement thumbnail CDN pipeline** — Upload thumbnails from the repo to OSS/CDN during `make upload`. Rewrite `thumbnail_url` in the compiled DB to the CDN URL. This single change will dramatically improve the visual quality of both the app and web client.
7. **Make category sort order curator-controlled** — Add an optional `sort_order` field to `_technique.json`, or use a convention like numeric prefixes in folder names. Alphabetical-by-folder-name is too rigid.
8. **Add `make serve` target** — Run the web client locally against the compiled `bmc.db` for curator preview. Something like `cd admin && BMC_DB_PATH=../data/bmc.db go run .`.
9. **Show category name in mobile search results** — Add `category_name` to the mobile `ContentItem` model (join with categories table in search query) and display it in `ContentRow`.

### P2 — Quality of Life

10. **Deduplicate web template headers** — Extract the shared nav, header, and CSS into a base template.
11. **Add pagination to web content list** — `?page=1&per_page=20` with LIMIT/OFFSET.
12. **Add content counts to mobile category rows** — Query `COUNT(*)` per category and display alongside the name.
13. **Handle Xiaohongshu and Douyin short URL formats** in deep linking.
14. **Move `Database.replaceWith` off MainActor on iOS** — Perform file operations on a background queue, then reopen the database.
15. **Make person field optional in content schema** — Change from `required` to optional, allowing content without a known author.

### P3 — Polish

16. **Add a simple onboarding screen** — Single-page explanation on first launch.
17. **Add retry button to sync failure state** — Especially important for first-launch-without-network.
18. **Improve sync status bar persistence** — Show "sync failed" until the user dismisses it or a successful sync occurs.
19. **Add dark mode color refinements on iOS** — Match the intentionality of Android's Material 3 dynamic colors.
20. **Extract shared Android composables** — Move `ContentRow`, `ContentThumbnail`, `PlatformBadge` to their own file (e.g., `SharedComponents.kt`).

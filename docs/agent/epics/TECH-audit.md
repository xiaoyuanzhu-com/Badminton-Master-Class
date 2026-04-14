# Tech Audit — Badminton Master Class

**Date:** 2026-04-14
**Scope:** Codebase audit for refactoring needs, performance issues, maintainability gaps, and RM-2 readiness
**Inputs:** All source files across admin (Go), iOS (SwiftUI), Android (Compose), schema (SQLite)

---

## Summary

| Severity | Count |
|----------|-------|
| Major (needs own EPIC) | 5 |
| Minor (quick fix) | 12 |

---

## Major Findings (each needs its own EPIC in RM-2)

### MAJOR-1: SQLite-file-distribution model cannot support user-specific data

**Category:** Architecture
**Blocks:** Personal Learning State (favorites, history), Learning Paths progress

The current architecture is: admin panel writes to a single SQLite file, uploads it to Aliyun OSS, and both iOS/Android download the full file and replace their local copy. This is a one-way, server-to-client, full-replacement sync model.

**Problems:**
1. **No place for user data.** Favorites, watch history, and learning path progress are user-specific. They cannot live in the downloaded DB because every sync replaces the entire file, destroying local data.
2. **No merge strategy.** `Database.replaceWith()` on both platforms does a delete-and-replace. There is no mechanism to preserve client-side tables during a server DB update.
3. **Sync destroys in-flight state.** If a user is browsing when sync completes, the DB handle is closed and reopened. On iOS, `Database.shared` is a singleton accessed directly from views with no concurrency protection — this can crash.

**Required architectural change:** Split into two databases or two table namespaces:
- **Content DB** (server-authored, synced from OSS) — categories, contents
- **User DB** (local-only, never overwritten) — favorites, watch_history, learning_path_progress

This is a prerequisite for the Personal Learning State EPIC and must be designed first.

---

### MAJOR-2: iOS database queries run on the main thread

**Category:** Performance
**Blocks:** Scaling beyond 100 content items, general UI responsiveness

Every iOS database call is synchronous and invoked directly from SwiftUI view bodies or `.onAppear` closures:

- `HomeView.onAppear` calls `Database.shared.categories(parentId: nil)` — main thread
- `HomeView.onChange(of: searchText)` calls `Database.shared.searchContents(keyword:)` — main thread, on every keystroke
- `CategoryView.onAppear` calls both `categories()` and `contents()` — main thread
- `DataSync.syncIfNeeded()` calls `Database.shared.replaceWith()` via `DispatchQueue.main.async` — closes and reopens the database on the main thread

With 20 items this is imperceptible. At 1,000+ items, the LIKE-based search query on every keystroke will cause visible jank. At 10,000 items, category loading will stutter.

**Required change:** Introduce an async database layer. Options:
- Move `Database` queries to a background actor/queue and publish results via `@Published` properties or async/await
- Use a ViewModel pattern with `Task { }` wrappers
- Add search debouncing (currently fires on every character)

---

### MAJOR-3: Full database download on every sync (no conditional fetch)

**Category:** Performance / Scalability
**Blocks:** Scaling content library, mobile data efficiency

Both iOS and Android download the entire `bmc.db` file on every app launch and every pull-to-refresh, with no caching or conditional logic:

- iOS `DataSync.syncIfNeeded()` — unconditional `URLSession.shared.downloadTask(with: remoteURL)`
- Android `DataSync.syncIfNeeded()` — unconditional `HttpURLConnection` download
- No `ETag`, `If-Modified-Since`, or `Last-Modified` headers are checked
- No local version tracking to skip unnecessary downloads

At 20 items the DB is tiny. At 10,000 items with thumbnails metadata, the DB could be several MB, downloaded on every launch over mobile data.

**Required change:** Implement conditional sync:
- Server-side: ensure OSS object has `ETag`/`Last-Modified` headers (OSS provides these by default)
- Client-side: store the last `ETag` or `Last-Modified` value locally, send `If-None-Match` / `If-Modified-Since` on subsequent requests, handle `304 Not Modified`

---

### MAJOR-4: Android has no back navigation on CategoryScreen

**Category:** Code Quality / UX Bug
**Blocks:** Basic usability on Android

The Android `CategoryScreen` Scaffold's `TopAppBar` has no navigation icon (no back arrow). The `BMCApp` NavHost provides `onSubcategoryTap` for forward navigation but the `CategoryScreen` composable has no `onBack` callback and no access to `navController.popBackStack()`.

This is confirmed in the user testing report and is a P0 blocker.

**Note:** While this is a bug fix rather than an architectural change, it is grouped as major because it was identified as a critical usability blocker in user testing and is already tracked in the Platform Stability EPIC. Listing it here for completeness; it does not need a separate EPIC.

---

### MAJOR-5: Android thumbnail loading is a hardcoded placeholder (Coil not integrated)

**Category:** Dependency Health / Feature Gap
**Blocks:** Content presentation, visual differentiation of content items

The Android `ContentThumbnail()` composable (in `CategoryScreen.kt`) is a hardcoded gray box with a play icon. It does not accept a URL parameter and cannot display images:

```kotlin
// Placeholder thumbnail — remote image loading requires Coil (to be added later)
@Composable
internal fun ContentThumbnail() {
    Box(...)  // Always shows placeholder
}
```

Meanwhile, iOS has `AsyncImage` integration that loads thumbnails from URLs. This means Android content rows are visually indistinguishable from each other.

**Required change:** Add Coil dependency and wire `ContentThumbnail` to accept and load `thumbnailUrl`. This is already identified in the Platform Stability EPIC.

---

## Minor Findings (summary lines, quick fixes)

### MINOR-1: Duplicated `ContentRow` composable across Android screens

**Category:** Code Quality — Duplication
**Files:** `HomeScreen.kt` (lines 217-266, `SearchResultRow`) and `CategoryScreen.kt` (lines 169-218, `ContentRow`)

These two composables are nearly identical — same layout, same sub-components (`ContentThumbnail`, `PlatformBadge`), same styling. They should be consolidated into a single shared `ContentRow` composable.

---

### MINOR-2: Duplicated platform badge/display-name mapping across iOS and Android

**Category:** Code Quality — Cross-platform duplication
**Files:** iOS `CategoryView.swift` (`PlatformBadge`), Android `CategoryScreen.kt` (`PlatformBadge`)

Both platforms independently define the same platform-to-display-name mapping (`bilibili` -> `B站`, etc.) and color mapping. This is inevitable in cross-platform native development but worth noting — any new platform added to the schema's CHECK constraint must be updated in three places (schema, iOS, Android).

---

### MINOR-3: iOS search has no debounce

**Category:** Performance
**File:** `HomeView.swift` line 31-32

```swift
.onChange(of: searchText) { _, newValue in
    searchResults = Database.shared.searchContents(keyword: newValue)
}
```

Fires a synchronous SQLite LIKE query on every keystroke. Should add a 300ms debounce. Combined with MAJOR-2 (main thread queries), this is the highest-risk performance path in the app.

---

### MINOR-4: iOS `DataSync` has near-complete code duplication between `syncDatabase()` and `syncIfNeeded()`

**Category:** Code Quality — Duplication
**File:** `DataSync.swift`

`syncDatabase()` (async, for pull-to-refresh) and `syncIfNeeded()` (fire-and-forget, for launch) contain ~40 lines of identical download logic. The only difference is the continuation wrapper. `syncIfNeeded()` should call `syncDatabase()` internally via `Task`.

---

### MINOR-5: Admin panel has no input validation for content creation

**Category:** Security / Data Integrity
**File:** `handlers.go` lines 174-197

The POST handler for `/contents` accepts form values with no validation:
- `title` can be empty string
- `categoryID` silently defaults to 0 on parse failure (which may not match any category)
- `source_platform` validation relies solely on the SQLite CHECK constraint — a constraint violation returns a raw SQL error to the user
- `source_url` is not validated as a URL

---

### MINOR-6: Admin panel default credentials are hardcoded

**Category:** Security
**File:** `main.go` lines 72-73

```go
username := getEnv("BMC_ADMIN_USER", "admin")
password := getEnv("BMC_ADMIN_PASSWORD", "admin")
```

Default username/password is `admin`/`admin`. While env var override exists, there is no warning when defaults are used, and no enforcement of minimum password complexity. For a personal project this is acceptable, but should be documented.

---

### MINOR-7: OSS upload in export handler is synchronous and blocks the response

**Category:** Performance
**File:** `handlers.go` line 508

```go
tryUploadToOSS(dbPath)  // blocks until upload completes
```

The export handler calls `tryUploadToOSS()` synchronously before streaming the DB file to the browser. If OSS is slow or times out, the admin user waits. This should be a goroutine (`go tryUploadToOSS(dbPath)`).

---

### MINOR-8: Schema has no `UNIQUE` constraint on `source_url`

**Category:** Data Integrity
**File:** `data/schema.sql`

Nothing prevents inserting the same URL twice. As the admin workflow scales to 200+ items, accidental duplicates become likely. Adding `UNIQUE(source_url)` would catch this at the database level.

---

### MINOR-9: `schema_version` table allows multiple rows with no constraint

**Category:** Data Integrity
**File:** `migrate.go`

The `schema_version` table has no primary key or unique constraint on `version`. Each migration INSERT adds a new row. While `getSchemaVersion()` uses `MAX(version)` which works correctly, the table design is unconventional — a single-row pattern with `UPDATE` would be cleaner.

---

### MINOR-10: Android `SyncConfig` uses reflection to access `BuildConfig` fields

**Category:** Code Quality — Fragility
**File:** `DataSync.kt` lines 37-44

```kotlin
val bucket: String = try {
    com.bmc.app.BuildConfig::class.java.getField("BMC_OSS_BUCKET").get(null) as String
} catch (_: Exception) { "bmc-data" }
```

This uses reflection to read optional `BuildConfig` fields, catching all exceptions silently. If the `BuildConfig` fields are actually added, they would be accessed directly as `BuildConfig.BMC_OSS_BUCKET`. The current approach is unnecessarily fragile. Should either declare the fields in `build.gradle.kts` or use a simpler config mechanism.

---

### MINOR-11: No test coverage for iOS or Android

**Category:** Testability
**Files:** No test files exist under `ios/` or `android/`

The Go admin panel has good test coverage (handlers, migration, auth, OSS upload). Neither the iOS nor Android project has any tests. Key untested areas:
- Database query correctness
- Sync state transitions
- Search behavior
- Navigation routing

For RM-2, at minimum the database layer on each platform should have unit tests, especially before adding favorites/history tables.

---

### MINOR-12: Android dependencies are slightly outdated

**Category:** Dependency Health
**File:** `android/app/build.gradle.kts`

| Dependency | Current | Latest (approx.) |
|-----------|---------|-------------------|
| Compose BOM | 2024.09.00 | 2025.x+ |
| Kotlin Compiler Extension | 1.5.8 | 1.5.14+ (or Compose Compiler Gradle plugin) |
| Activity Compose | 1.8.2 | 1.9.x |
| Navigation Compose | 2.7.7 | 2.8.x |
| Lifecycle | 2.7.0 | 2.8.x |
| AGP | 8.2.2 | 8.5+ |

Also: `compileSdk = 34` and `targetSdk = 34` — should target API 35 for current Play Store requirements. `isMinifyEnabled = false` in release — should enable R8 for production builds.

---

## Schema Readiness Assessment

**Question:** Can the current schema support favorites, history, difficulty levels?

| Feature | Schema ready? | What's needed |
|---------|--------------|---------------|
| Favorites | No | New `favorites` table (user-local, separate from synced DB) |
| Watch history | No | New `watch_history` table (user-local, separate from synced DB) |
| Difficulty levels | Partially | Add `difficulty TEXT` column to `contents` + migration v2 |
| Learning paths | No | New `learning_paths` and `learning_path_steps` tables |
| Editor's notes | Partially | Could reuse `summary` field or add `editors_note TEXT` column |
| Content count per category | Yes | Can be computed with `COUNT(*)` query (no schema change) |
| "New" content badge | Yes | `created_at` column already exists |

The critical blocker is MAJOR-1: the two-database architecture must be designed before any user-specific tables can be added, because the current sync model replaces the entire DB file.

---

## Scalability Assessment

**What breaks at scale?**

| Scale | What breaks |
|-------|------------|
| 100 items | iOS search jank begins (no debounce + main thread LIKE queries) |
| 1,000 items | Full DB download on every launch becomes noticeable on slow connections (~1-5 MB). Admin panel flat category dropdown becomes unwieldy |
| 10,000 items | LIKE-based search without FTS becomes slow (~100ms+). Admin panel loads all contents in a single page (no pagination). Full DB download is unacceptable (~50+ MB) |

**Recommended mitigations:**
- Conditional sync (MAJOR-3) handles the download size issue
- FTS5 virtual table for search at 1,000+ items
- Admin panel pagination at 100+ items

---

## Recommended RM-2 EPIC Additions from This Audit

Based on this audit, the following technical EPICs should be added to the RM-2 plan alongside the product EPICs from the product research:

1. **EPIC: Two-Database Architecture** (MAJOR-1) — Split synced content DB from local user DB. Prerequisite for Personal Learning State. Effort: M.
2. **EPIC: Async Database Layer for iOS** (MAJOR-2) — Move all DB queries off the main thread. Effort: S-M.
3. **EPIC: Conditional Sync** (MAJOR-3) — ETag/If-Modified-Since for both platforms. Effort: S.

MAJOR-4 (Android back nav) and MAJOR-5 (Coil integration) are already covered by the Platform Stability EPIC from the product research.

The 12 minor items can be addressed as part of existing EPICs or as a single "tech debt cleanup" pass.

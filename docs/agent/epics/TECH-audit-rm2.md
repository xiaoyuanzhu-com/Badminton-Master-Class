# Tech Audit — Post-RM-2, Pre-RM-3

**Date:** 2026-04-14
**Scope:** Full codebase audit after RM-2 delivery, focused on readiness for RM-3 features (learning paths, thumbnail pipeline, user state)
**Inputs:** All source files, RM-3 product research (PROD-research-rm3.md), prior tech audit (TECH-audit.md)

---

## Summary

| Severity | Count |
|----------|-------|
| Major (needs own EPIC) | 4 |
| Minor (stat line) | 14 |

**Overall assessment:** RM-2 resolved three of the five prior major findings (MAJOR-2: iOS async DB, MAJOR-3: ETag sync, MAJOR-4/5: Android back nav + Coil). The codebase is clean and well-structured for its current scope. However, four architectural gaps will block or complicate RM-3 features if not addressed.

---

## What RM-2 Fixed (from prior audit)

| Prior Finding | Status | Evidence |
|---------------|--------|----------|
| MAJOR-2: iOS main-thread DB queries | **Fixed** | `Database.swift` has `queryQueue` + async wrappers; `HomeView` uses `.task` and `await` |
| MAJOR-3: No conditional sync | **Fixed** | Both platforms send `If-None-Match` with stored ETag, handle 304 |
| MAJOR-4: Android no back button | **Fixed** | `CategoryScreen.kt` has `onBack` callback and back arrow in TopAppBar |
| MAJOR-5: Android Coil not integrated | **Fixed** | `CategoryScreen.kt` uses `coil.compose.AsyncImage` with `thumbnailUrl` parameter |
| MINOR-1: Duplicated ContentRow | **Fixed** | Shared `ContentRow` composable in `CategoryScreen.kt` used by both screens |
| MINOR-3: iOS search no debounce | **Fixed** | `HomeView.swift` has 300ms `Task.sleep` debounce in `onChange` |
| MINOR-4: iOS DataSync duplication | **Fixed** | Single `performSync()` method, `syncIfNeeded` calls it via `Task` |

---

## Major Findings (each needs own EPIC in RM-3)

### MAJOR-1: Schema has no learning path tables — compiler and validator need extension

**Category:** Architecture / Schema
**Blocks:** PATH EPIC (learning paths)

The v2 schema has `categories`, `people`, and `contents`. RM-3 learning paths require:

1. **New tables:** `learning_paths` and `path_steps` (or equivalent) in the compiled DB.
2. **New content file type:** `data/content/paths/*.json` files do not exist yet. The compiler (`build.py`) only walks `techniques/` for content and `people/` for people. It has no concept of paths.
3. **New schema in `validate.py`:** The validator has no `path.schema.json` and no cross-reference validation (e.g., verifying that content slugs in path steps actually exist).
4. **Schema version bump:** `SCHEMA_VERSION` must go from 2 to 3. The admin panel's `migrate.go` needs a v3 migration.

**Impact:** This is not a refactor — it is new plumbing. But the pattern is well-established (people/categories show how to add a new entity), so the effort is moderate.

**Recommendation:** Create a dedicated **EPIC: SCHEMA-V3** covering:
- `path.schema.json` definition
- `build.py` extension to compile paths
- `validate.py` extension to validate paths
- `migrate.go` v3 migration
- Admin panel path display pages
- Mobile model classes for paths

---

### MAJOR-2: Android DB queries run on the main thread

**Category:** Performance
**Blocks:** Scaling beyond 100 items, learning path step lists

RM-2 fixed this for iOS but **not for Android**. All Android database calls are synchronous and invoked from `LaunchedEffect` composables on the main dispatcher:

- `HomeScreen.kt` line 68: `categories = Database.getInstance(context).categories(parentId = null)` — runs on `LaunchedEffect(Unit)` which defaults to `Dispatchers.Main`
- `HomeScreen.kt` line 75: `searchResults = Database.getInstance(context).searchContents(searchQuery)` — main thread, no debounce delay
- `CategoryScreen.kt` line 63-65: `db.categories()` and `db.contents()` — main thread

The `LaunchedEffect` coroutine scope is `Dispatchers.Main` by default. These calls touch SQLite directly without `withContext(Dispatchers.IO)`.

**Impact:** With 27 items this works. At 100+ items with LIKE search, this will cause frame drops. Learning path step lists (which join multiple tables) will be heavier queries.

**Recommendation:** Already identified as FIX-2 in RM-3 research. Wrap all `Database` calls in `withContext(Dispatchers.IO)`. This is small effort but critical — classify as part of the FIX EPIC, not a standalone EPIC.

**Reclassification:** This is a **major finding** that belongs in the **FIX EPIC** (not its own EPIC). Listing it as major because it affects every screen.

---

### MAJOR-3: Android search has no debounce

**Category:** Performance
**Blocks:** Search usability at scale

iOS search has a 300ms debounce (`Task.sleep(nanoseconds: 300_000_000)`). Android search fires a DB query on every keystroke:

```kotlin
// HomeScreen.kt line 71-76
LaunchedEffect(searchQuery) {
    if (searchQuery.isBlank()) {
        searchResults = emptyList()
    } else {
        searchResults = Database.getInstance(context).searchContents(searchQuery)
    }
}
```

No `delay()` call. Combined with MAJOR-2 (main thread queries), this means every keystroke triggers a synchronous LIKE query on the main thread.

**Recommendation:** Already identified as FIX-1 in RM-3 research. Add `delay(300)` before the search call. Part of the FIX EPIC.

**Reclassification:** Like MAJOR-2, this is a **major finding** belonging in the **FIX EPIC**.

---

### MAJOR-4: No user-state persistence layer exists

**Category:** Architecture
**Blocks:** STATE EPIC (favorites, path progress)

The prior audit's MAJOR-1 (two-database architecture) was deferred from RM-2. The RM-3 research correctly recommends a simpler JSON-file approach for user state. However, **zero infrastructure exists** for client-side state:

- No `UserState` class on either platform
- No local file path convention for user data
- No serialization/deserialization pattern
- The `Database.replaceWith()` method replaces the content DB without touching other files (good), but there is no documented contract about which files are "sync-safe" vs. "user-owned"

**Impact:** Favorites and path progress both depend on this. It is the foundation for the STATE EPIC.

**Recommendation:** The STATE EPIC in the RM-3 research already covers this (STATE-1 through STATE-7). This finding confirms that STATE needs to be a distinct EPIC, not absorbed into PATH. The JSON-file approach is the right call.

---

## Minor Findings (stat lines)

### MINOR-1: `content.schema.json` marks `person` as required, but compiler handles it as optional

**Category:** Schema inconsistency
**File:** `data/content/schemas/content.schema.json` line 8 — `"required": ["title", "source_url", "source_platform", "person"]`
**File:** `data/build.py` line 174 — `person_slug = data.get("person", "")`

The schema says `person` is required, but the compiler gracefully handles its absence. This means validation will reject content files without a person, even though the DB and compiler support it. The RM-3 research identifies this as FIX-5.

**Fix:** Remove `person` from the `required` array in `content.schema.json`.

---

### MINOR-2: Thumbnail URL is always empty in compiled DB

**Category:** Feature gap (known)
**File:** `data/build.py` line 171 — `thumbnail_url = ""  # empty for now; future: OSS URL`

The compiler detects `{slug}.png` files (line 170) but never writes the URL. The `ingest.py` script downloads thumbnails. But even when thumbnail files exist alongside content JSON, the DB `thumbnail_url` column is always empty.

**Fix:** Part of the THUMB EPIC (THUMB-1). The pipeline is partially built; just needs the last step.

---

### MINOR-3: Android nav route passes category name unencoded in URL

**Category:** Bug
**File:** `BMCApp.kt` line 43 — `navController.navigate("category/${category.id}/${category.name}")`

Category names contain Chinese characters (e.g., "基础技术"). These are interpolated directly into the nav route string without URL encoding. If a category name contains `/` or other special characters, navigation will break.

**Fix:** Already identified as FIX-3 in RM-3 research. Pass only `category.id` and look up the name from the DB in `CategoryScreen`.

---

### MINOR-4: Duplicated platform display-name/color mapping across three surfaces

**Category:** Code quality — cross-platform duplication
**Files:** iOS `CategoryView.swift` (PlatformBadge), Android `CategoryScreen.kt` (PlatformBadge), Go `handlers.go` (platformLabel func)

Each surface independently maps `bilibili` -> `B站`, etc. Any new platform added to the schema's CHECK constraint must be updated in four places (schema, iOS, Android, Go). This is inherent to native cross-platform development but worth tracking.

**Fix:** No action needed now. Could be automated if platform list grows significantly.

---

### MINOR-5: `technique.schema.json` has no `sort_order` field

**Category:** Schema gap
**File:** `data/content/schemas/technique.schema.json` — only has `name` and `icon`

The RM-3 research proposes FIX-7: add `sort_order` to `_technique.json` so curators control display order. Currently, category sort order is determined by the compiler's filesystem traversal order (alphabetical by directory name), which is fragile and not curator-friendly.

**Fix:** Add optional `sort_order` integer field to technique schema. Update compiler to use it when present, falling back to alphabetical.

---

### MINOR-6: Web admin panel templates have no shared base template

**Category:** Code quality — web client
**Files:** All `admin/templates/*.html`

Each template is a standalone HTML file. Adding a nav bar, CSS, or footer requires editing every template. The RM-3 research identifies this as WEBPOL-1.

**Fix:** Extract shared header/nav/CSS into Go template blocks. Part of WEBPOL EPIC.

---

### MINOR-7: Web admin panel has no pagination

**Category:** Scalability — web client
**File:** `admin/handlers.go` — `contentsHandler` loads all contents in one query

At 27 items this is fine. At 100+ items (RM-3 target), the contents page will be a long scroll with no pagination.

**Fix:** Add `?page=1&per_page=20` query params. Part of WEBPOL EPIC.

---

### MINOR-8: `schema_version` table still allows multiple rows with no constraint

**Category:** Data integrity (cosmetic)
**File:** `data/build.py` line 74, `admin/migrate.go` line 73

The `schema_version` table has no primary key or unique constraint. Each migration inserts a new row. `getSchemaVersion()` uses `MAX(version)` which works, but the table grows one row per version. This was noted in the prior audit and remains unchanged.

**Fix:** Low priority. Could switch to a single-row `UPDATE` pattern when writing v3 migration.

---

### MINOR-9: No `UNIQUE` constraint on `source_url` in schema

**Category:** Data integrity
**File:** `data/schema.sql` line 26

The validator catches duplicate URLs across content files (line 180-185 of `validate.py`), but the compiled DB has no `UNIQUE(source_url)` constraint. If the compiler ever has a bug, duplicates can enter the DB silently.

**Fix:** Add `UNIQUE` constraint to `source_url` in schema. Low priority since validator already catches this.

---

### MINOR-10: Android `SyncConfig` still uses reflection for BuildConfig fields

**Category:** Code quality — fragility
**File:** `DataSync.kt` lines 37-44

Uses `BuildConfig::class.java.getField(...)` with a catch-all exception handler. Should either declare `buildConfigField` entries in `build.gradle.kts` or use a simpler constants-based approach.

**Fix:** Small cleanup task. Replace reflection with direct `BuildConfig` fields or hardcoded defaults.

---

### MINOR-11: No test coverage for iOS or Android

**Category:** Testability
**Files:** No test files exist under `ios/` or `android/`

The Go admin panel has solid test coverage (20+ tests covering handlers, migration, auth). Neither mobile platform has any tests. Key untested areas:
- Database query correctness
- Sync state transitions (ETag handling, 304 responses)
- Deep link URL computation
- Search debounce behavior

**Fix:** Before RM-3 adds path progress and user state, at minimum add unit tests for the database layer and deep link computation on each platform. This could be a task within the FIX EPIC.

---

### MINOR-12: No tests for Python build pipeline

**Category:** Testability
**Files:** `data/build.py`, `data/content/validate.py`, `data/ingest.py`

The compiler, validator, and ingestion script have zero test files. As RM-3 extends all three (path validation, path compilation, thumbnail URL writing), this becomes riskier.

**Fix:** Add at least a smoke test that runs `build.py` on the existing content and verifies the output DB has expected row counts and schema.

---

### MINOR-13: `ingest.py` does not write `thumbnail_url` into content JSON

**Category:** Feature gap (partial implementation)
**File:** `data/ingest.py` — `write_content()` (line 307-337)

The ingestion script fetches thumbnail URLs from page metadata (line 157) and downloads thumbnail images (line 504-505), but the content JSON written by `write_content()` does not include a `thumbnail_url` field. The content schema also has no `thumbnail_url` field.

This creates a gap: thumbnails are downloaded as files alongside the content JSON, but the compiler has no way to discover them and write URLs into the DB.

**Fix:** Part of the THUMB EPIC. Either add `thumbnail_url` to the content schema and write it in `ingest.py`, or have the compiler derive URLs from adjacent image files.

---

### MINOR-14: Android dependencies are outdated (from prior audit, unchanged)

**Category:** Dependency health
**File:** `android/app/build.gradle.kts`

| Dependency | Current | Notes |
|-----------|---------|-------|
| Compose BOM | 2024.09.00 | 18+ months old |
| Kotlin Compiler Extension | 1.5.8 | Should align with newer Compose |
| compileSdk / targetSdk | 34 | Play Store now requires 35 |
| Coil | 2.6.0 | Coil 3.x is current |
| `isMinifyEnabled = false` in release | -- | Should enable R8 for production |

**Fix:** Dependency bump pass. Medium effort due to potential Compose API changes.

---

## Schema Readiness for RM-3

| RM-3 Feature | Schema Ready? | What's Needed |
|--------------|---------------|---------------|
| Learning paths (display) | No | `learning_paths` table, `path_steps` table, v3 migration |
| Learning paths (files) | No | `path.schema.json`, `data/content/paths/` directory |
| Path progress tracking | No | User-state JSON file on each platform (not in DB) |
| Favorites | No | User-state JSON file on each platform |
| Thumbnails from URL | Partially | `thumbnail_url` column exists in DB but always empty; compiler needs to write it |
| Content freshness | Partially | `created_at` column exists but is always `datetime('now')` from compile time, not content creation time |
| Category sort order | Partially | `sort_order` column exists in DB; `_technique.json` files need `sort_order` field |

---

## RM-3 Readiness Assessment by Component

### Content-as-code pipeline (build.py, validate.py, ingest.py)

**Verdict: Solid foundation, needs extension, not refactoring.**

The pipeline pattern is well-designed:
- Walk directories, load JSON, validate against schema, compile to SQLite
- Adding learning paths follows the same pattern: new directory, new schema, new compiler function
- `ingest.py` is feature-complete for single-URL ingestion; it correctly handles all 5 platforms

Gaps to close:
1. Compiler needs `build_paths()` function (parallel to `build_people()`, `build_categories()`)
2. Validator needs path schema and cross-reference checks
3. Thumbnail URL needs to flow from ingest -> content JSON -> compiler -> DB

### iOS app

**Verdict: Ready for RM-3 with moderate effort.**

Strengths:
- Async DB layer already works
- Clean SwiftUI navigation with `NavigationStack`
- Search debounce is implemented
- `ContentRow` is reusable for path step lists

Gaps:
- No `LearningPath` or `PathStep` model types
- No tab bar or navigation section for paths (currently single-screen with categories)
- No user-state persistence layer

### Android app

**Verdict: Needs FIX EPIC first, then ready for RM-3.**

Strengths:
- Compose navigation with typed arguments
- Shared `ContentRow` composable
- Coil image loading integrated
- DataSync with ETag works

Gaps:
- DB queries on main thread (FIX-2) — must fix before adding path queries
- Search has no debounce (FIX-1)
- Nav route URL encoding issue (FIX-3)
- No `LearningPath` model
- No user-state persistence layer

### Web client (Go admin panel)

**Verdict: Functional but will need template work for paths.**

Strengths:
- Good test coverage
- Migration system handles schema upgrades cleanly
- Search works across contents and people

Gaps:
- No path-related pages
- No shared template base (each page is standalone HTML)
- No pagination (will matter at 100+ items)

---

## Recommended EPIC Classification

### Major findings -> EPICs

| Finding | Recommended EPIC | Rationale |
|---------|-----------------|-----------|
| MAJOR-1: Schema/compiler/validator need path support | **SCHEMA-V3** (new) or absorb into **PATH** EPIC | Could be the first phase of PATH. The schema work is a prerequisite that PATH-3 and PATH-4 already cover. Recommend keeping it within PATH rather than a separate EPIC. |
| MAJOR-2: Android main-thread DB queries | **FIX** EPIC (FIX-2) | Small effort, must ship first |
| MAJOR-3: Android search no debounce | **FIX** EPIC (FIX-1) | Small effort, must ship first |
| MAJOR-4: No user-state layer | **STATE** EPIC | Already scoped in RM-3 research |

### Minor findings -> placement

| Finding | Where to Address |
|---------|-----------------|
| MINOR-1: person required in schema | FIX EPIC (FIX-5) |
| MINOR-2: thumbnail_url always empty | THUMB EPIC (THUMB-1) |
| MINOR-3: Android nav URL encoding | FIX EPIC (FIX-3) |
| MINOR-4: Platform mapping duplication | No action (inherent to cross-platform) |
| MINOR-5: No sort_order in technique schema | FIX EPIC (FIX-7) |
| MINOR-6: No shared web template | WEBPOL EPIC (WEBPOL-1) |
| MINOR-7: No web pagination | WEBPOL EPIC (WEBPOL-2) |
| MINOR-8: schema_version multi-row | Opportunistic fix during v3 migration |
| MINOR-9: No UNIQUE on source_url | Opportunistic fix during v3 migration |
| MINOR-10: Android BuildConfig reflection | FIX EPIC (minor task) |
| MINOR-11: No mobile tests | FIX EPIC or standalone test task |
| MINOR-12: No Python pipeline tests | INGEST or PATH EPIC (add alongside new features) |
| MINOR-13: ingest.py missing thumbnail_url | THUMB EPIC (THUMB-1) |
| MINOR-14: Android deps outdated | FIX EPIC (dependency bump) |

---

## Key Takeaway

The RM-2 codebase is in good shape. The three major RM-2 fixes (iOS async DB, ETag sync, Android Coil/back nav) resolved the most critical prior findings. The remaining gaps are:

1. **FIX EPIC must ship first** — Android main-thread queries + no debounce are the highest-risk items.
2. **PATH can absorb the schema work** — No need for a separate SCHEMA-V3 EPIC; PATH-3 and PATH-4 already scope it.
3. **STATE is correctly scoped as its own EPIC** — The JSON-file approach is right; building it separately from PATH keeps dependencies clean.
4. **Test coverage is the biggest long-term debt** — Zero tests on mobile and Python. Not blocking RM-3 but increasing risk with each feature addition.

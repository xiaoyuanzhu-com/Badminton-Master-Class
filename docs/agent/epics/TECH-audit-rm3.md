# Tech Audit — Post-RM-3

**Date:** 2026-04-15
**Scope:** Full codebase audit after RM-3 delivery, focused on tech debt, performance, architecture, and readiness for RM-4
**Inputs:** All source files across iOS (SwiftUI), Android (Compose), Web (Go), Python build pipeline, previous audits (TECH-audit.md, TECH-audit-rm2.md)

---

## Summary

| Severity | Count |
|----------|-------|
| Major (needs own EPIC or priority fix) | 5 |
| Minor (stat line) | 11 |

**Overall assessment:** RM-3 delivered substantial new functionality (schema v3, learning paths, user state persistence, thumbnail pipeline, enhanced search). The prior RM-2 audit's four major findings have all been addressed. The codebase is well-structured, but RM-3's rapid growth introduced new architectural concerns: a thread-safety issue in iOS's Database singleton, N+1 query patterns in path detail loading, and an inefficient content-slug resolution in the build pipeline. Android is in much better shape than pre-RM-3 (DB queries moved to IO, search debounce added, URL encoding fixed).

---

## What RM-3 Fixed (from prior audit)

| Prior Finding | Status | Evidence |
|---------------|--------|----------|
| RM2-MAJOR-1: Schema/compiler need path support | **Fixed** | Schema v3: `learning_paths`, `path_steps`, `path_step_contents` tables; `build.py` has `build_paths()`; `validate.py` validates paths with cross-reference checks; `path.schema.json` exists |
| RM2-MAJOR-2: Android DB queries on main thread | **Fixed** | `HomeScreen.kt` line 92-97 uses `withContext(Dispatchers.IO)`, `CategoryScreen.kt` line 72-77 same; all screens use IO dispatcher |
| RM2-MAJOR-3: Android search no debounce | **Fixed** | `HomeScreen.kt` line 116: `delay(300)` before search; `LaunchedEffect(searchQuery)` auto-cancels on new keystroke |
| RM2-MAJOR-4: No user-state layer | **Fixed** | `UserState.swift` (iOS) and `UserState.kt` (Android) — JSON file persistence with favorites and path progress |
| RM2-MINOR-1: person required in schema | **Fixed** | `content.schema.json` line 7: `"required": ["title", "source_url", "source_platform"]` — person removed from required |
| RM2-MINOR-2: thumbnail_url always empty | **Fixed** | `build.py` line 199: `thumbnail_url = data.get("thumbnail_url", "")` reads from content JSON; `backfill_thumbnails.py` fills URLs |
| RM2-MINOR-3: Android nav URL encoding | **Fixed** | `BMCApp.kt` line 44: `URLEncoder.encode(category.name, "UTF-8")`, line 61-62: `URLDecoder.decode()` on receipt |
| RM2-MINOR-5: No sort_order in technique schema | **Fixed** | `technique.schema.json` has `sort_order` field; `build.py` line 155: `explicit_order = data.get("sort_order", sort_order)` |

---

## Major Findings

### MAJOR-1: iOS Database singleton is not thread-safe

**Category:** Concurrency / Crash Risk
**Severity:** High
**Files:** `ios/BadmintonMasterClass/Database.swift`

The `Database` class uses a single `OpaquePointer? (db)` handle accessed from both the main thread (via `replaceWith()` in `DataSync.swift` line 103: `await MainActor.run { Database.shared.replaceWith(...) }`) and the background `queryQueue` (all `*Async` methods dispatch to `Self.queryQueue`).

**Problems:**
1. **Race condition on `replaceWith()`.** When sync completes, `replaceWith()` calls `closeDatabase()` then `openDatabase()` on the main thread. If a background query is in flight on `queryQueue`, it will use a closed or nil `db` pointer — this is undefined behavior with the SQLite C API and can crash.
2. **No locking or actor isolation on `db`.** The `db` property is a plain mutable `OpaquePointer?` with no synchronization. `queryQueue` reads it concurrently while `replaceWith()` writes it from MainActor.
3. **PathDetailView makes concurrent queries via `withTaskGroup`** (line 97-107), meaning multiple background queries hit the same `db` handle simultaneously. While SQLite in WAL mode supports concurrent reads, the current code uses the default journal mode and a single connection with no WAL pragma.

**Impact:** With infrequent syncs and a small dataset, this rarely triggers. But it is a latent crash: if a user pulls-to-refresh while viewing a path detail (which fires 3+ concurrent queries), the timing window for a crash opens.

**Required fix:** Either:
- Make `Database` an actor (Swift actor isolation eliminates the race automatically)
- Route all operations (including `replaceWith`) through `queryQueue` with a serial discipline
- At minimum, add a lock (`NSLock` or `os_unfair_lock`) around `db` access

---

### MAJOR-2: `build_content_slug_map()` re-reads all JSON files and issues N queries

**Category:** Build Pipeline Performance
**Severity:** Medium
**File:** `data/build.py` lines 234-263

After `build_contents()` inserts all content items, `build_content_slug_map()` walks the entire `TECHNIQUES_DIR` again, re-reads every JSON file, and then for each file runs a `SELECT id FROM contents WHERE source_url = ?` query. With 20 content files this is invisible. At 200+ items this doubles the build time unnecessarily.

**Problems:**
1. **Full re-walk of the filesystem** — `os.walk()` iterates the same directory tree already walked by `build_contents()`.
2. **N SELECT queries** — one per content file to map slug to DB ID.
3. **Dead code** — the first loop (lines 237-245) contains only `pass` and does nothing. It is a leftover from incremental development.

**Required fix:** Return the slug-to-ID mapping directly from `build_contents()` by collecting `{slug: cur.lastrowid}` during the insert loop. This eliminates the entire function, the filesystem re-walk, and all N queries.

---

### MAJOR-3: N+1 query pattern in path step content loading (both platforms)

**Category:** Performance
**Severity:** Medium
**Files:** Android `PathDetailScreen.kt` lines 64-76, iOS `PathDetailView.swift` lines 94-108, Go `handlers.go` lines 303-334

All three platforms load path detail with an N+1 pattern:
1. One query for all steps: `SELECT ... FROM path_steps WHERE path_id = ?`
2. N queries for contents: for each step, `SELECT ... FROM path_step_contents JOIN contents WHERE step_id = ?`

With 3 paths averaging 8 steps each, this is 1 + 8 = 9 queries per path view. At 20 steps, it is 21 queries.

iOS partially mitigates this with `withTaskGroup` (concurrent queries), but the total query count is unchanged. Android runs them sequentially in a single `withContext(Dispatchers.IO)` block (lines 70-74: `for (step in loadedSteps) { contentsMap[step.id] = db.pathStepContents(step.id) }`).

**Required fix:** Add a `pathAllStepContents(pathId:)` method that does a single query with a JOIN across `path_steps`, `path_step_contents`, and `contents`, returning all step contents for a given path at once. Then group by step ID in application code.

---

### MAJOR-4: `admin/schema.sql` is stale — still labeled v2, missing v3 tables

**Category:** Schema Drift
**Severity:** Medium
**Files:** `admin/schema.sql` (line 1: `-- admin/schema.sql — v2`), `data/schema.sql` (line 1: `-- data/schema.sql — v2`)

Both schema reference files are labeled v2 and lack the `learning_paths`, `path_steps`, and `path_step_contents` tables. The actual v3 schema lives only inside `build.py`'s `SCHEMA_SQL` constant and `migrate.go`'s migration v3.

**Problems:**
1. `admin/schema.sql` is embedded via `//go:embed schema.sql` but never used (the Go admin uses `migrateDB()` which applies migrations incrementally). So this file is purely documentation — but it is misleading documentation.
2. `data/schema.sql` is the "reference" schema but it is incomplete.
3. New developers or tools referencing these files will see an outdated schema.

**Required fix:** Update both `admin/schema.sql` and `data/schema.sql` to match the full v3 schema from `build.py`. Update the version comment to `v3`.

---

### MAJOR-5: Android `UserState` saves to disk synchronously on every mutation

**Category:** Performance / ANR Risk
**Severity:** Medium
**File:** `android/app/src/main/java/com/bmc/app/data/UserState.kt` lines 34-39, 49-58, 93-110

Every call to `toggleFavorite()` or `toggleStepCompleted()` calls `save()` immediately, which does `file.writeText(json.toString(2))` — a synchronous file write. These methods are called from Compose composables on the main thread (e.g., `CategoryScreen.kt` line 304: `userState.toggleFavorite(item.id)`).

**Contrast with iOS:** The iOS `UserState` uses `Combine`'s `.debounce(for: .milliseconds(300))` to coalesce rapid mutations and only saves after 300ms of inactivity. This is the correct pattern.

**Impact:** On a modern device, writing a few hundred bytes is fast. But on slower devices or if the user rapidly taps multiple favorites, each tap blocks the main thread for a disk write. At scale (large favorites list + rapid interaction), this can cause frame drops or ANR.

**Required fix:** Add debounced save — either:
- Post the save to a coroutine with `Dispatchers.IO` and a short debounce delay
- Use `SharedPreferences` (which handles async writes natively via `apply()`)

---

## Minor Findings

### MINOR-1: iOS has three identical `DifficultyBadge` implementations

**Category:** Code Quality — Duplication
**Files:** `HomeView.swift` (`DifficultyBadge`, lines 256-287), `CategoryView.swift` (`ContentDifficultyBadge`, lines 196-227), `PathDetailView.swift` (`DifficultyBadgeInline`, lines 114-145)

Three separate structs with identical `displayName` and `badgeColor` switch statements, identical view bodies. Only the struct name differs. Should be consolidated into a single `DifficultyBadge` in a shared file.

---

### MINOR-2: Android difficulty badge shows raw English strings instead of Chinese labels

**Category:** UX Bug
**File:** `HomeScreen.kt` lines 428-434 (LearningPathCard) — `Text(text = path.difficulty, ...)` renders "beginner" instead of "入门"

The `ContentDifficultyBadge` in `CategoryScreen.kt` correctly maps difficulty to Chinese labels. But `LearningPathCard` and `SearchPathRow` in `HomeScreen.kt` display the raw difficulty string. This is an inconsistency: path cards show "beginner" while content rows show "入门".

**Fix:** Use `ContentDifficultyBadge(difficulty = path.difficulty)` composable in path cards and search rows, or extract the mapping to a shared function.

---

### MINOR-3: Duplicated platform badge/difficulty mapping across four surfaces

**Category:** Code Quality — Cross-platform duplication
**Files:** iOS `CategoryView.swift` (PlatformBadge, ContentDifficultyBadge), Android `CategoryScreen.kt` (PlatformBadge, ContentDifficultyBadge), Go `handlers.go` (platformLabel, difficultyLabel template funcs)

This is inherent to native cross-platform development. Tracking for completeness — any new platform or difficulty level must be updated in all four codebases.

---

### MINOR-4: `schema_version` table still allows multiple rows

**Category:** Data Integrity (cosmetic)
**Files:** `build.py` line 101-103, `migrate.go` line 104-111

Both the compiler and migration system insert a new row for each version. The table has no primary key or unique constraint. `getSchemaVersion()` uses `MAX(version)` which works but is unconventional. This was identified in both previous audits and remains unchanged.

**Fix:** Low priority. Could switch to a single-row `UPDATE` pattern in a future migration.

---

### MINOR-5: No `UNIQUE` constraint on `source_url` in compiled DB

**Category:** Data Integrity
**File:** `build.py` SCHEMA_SQL (line 58: `source_url TEXT NOT NULL`)

The validator catches duplicate URLs across content files, but the compiled DB has no `UNIQUE(source_url)` constraint. A compiler bug could silently insert duplicates.

**Fix:** Add `UNIQUE` to `source_url` in the schema. Low priority since validator catches this pre-compilation.

---

### MINOR-6: Android `SyncConfig` still uses reflection for BuildConfig

**Category:** Code Quality — Fragility
**File:** `DataSync.kt` lines 37-44

Uses `BuildConfig::class.java.getField(...)` with catch-all exception handler. This was identified in both previous audits and remains unchanged.

**Fix:** Replace with direct `BuildConfig` fields (add `buildConfigField` entries in `build.gradle.kts`) or just use hardcoded constants.

---

### MINOR-7: No test coverage for iOS or Android

**Category:** Testability
**Files:** No test files exist under `ios/` or `android/`

The Go admin panel has 30+ test cases covering all handlers, migration (fresh/idempotent/upgrade), and auth. Neither mobile platform has any tests. This was identified in both previous audits and remains unchanged.

Key untested areas that have grown in RM-3:
- Database query correctness (especially the new path-related queries with JOINs)
- UserState persistence (JSON serialization/deserialization round-trip)
- Deep link URL computation (5 platform handlers)
- Search debounce and concurrent query cancellation

**Risk:** Each release adds more untested code. The path-related queries with multi-table JOINs are the highest risk area — a schema mismatch or column ordering error would not be caught until runtime.

---

### MINOR-8: No tests for Python build pipeline

**Category:** Testability
**Files:** `data/build.py`, `data/content/validate.py`, `data/ingest.py`, `data/backfill_thumbnails.py`

Four Python scripts with zero tests. `build.py` now has 5 build functions (people, categories, contents, slug map, paths) and an OSS upload. `validate.py` validates 4 schema types with cross-reference checks. None of this is tested.

**Fix:** Add at minimum a smoke test: run `build.py`, open the resulting `bmc.db`, and verify row counts and schema version.

---

### MINOR-9: Android dependencies are outdated (3rd consecutive audit)

**Category:** Dependency Health
**File:** `android/app/build.gradle.kts`

| Dependency | Current | Notes |
|-----------|---------|-------|
| Compose BOM | 2024.09.00 | ~18 months old |
| Kotlin Compiler Extension | 1.5.8 | Tied to older Compose |
| compileSdk / targetSdk | 34 | Play Store requires 35 for new submissions |
| Coil | 2.6.0 | Coil 3.x is current |
| Activity Compose | 1.8.2 | 1.9.x available |
| Navigation Compose | 2.7.7 | 2.8.x available |
| `isMinifyEnabled = false` in release | -- | Should enable R8 for production |

This has been noted in all three audits. The risk increases as the codebase grows — eventually a dependency bump will require migration effort proportional to the version gap.

---

### MINOR-10: Web admin panel still has no shared base template

**Category:** Code Quality — Web
**Files:** 9 standalone HTML templates in `admin/templates/`

Each template is a complete HTML document. Adding shared navigation, CSS, or a footer requires editing all 9 files. This was identified in the RM-2 audit.

---

### MINOR-11: Web admin panel still has no pagination

**Category:** Scalability — Web
**File:** `admin/handlers.go` — `contentsHandler` loads all contents in one query with no `LIMIT`/`OFFSET`

At the current 20 items this is fine. At 100+ items (RM-4 target with more ingested content), the contents page will be an unusable scroll.

---

## Schema Audit

### v3 Schema Assessment

The v3 schema is well-designed:

| Aspect | Assessment |
|--------|-----------|
| `learning_paths` table | Clean, minimal columns (title, summary, difficulty, sort_order). No issues. |
| `path_steps` table | `day` column is `INTEGER` (nullable) — correct for optional day numbering. `step_order` provides explicit ordering. |
| `path_step_contents` junction table | Proper many-to-many with `sort_order`. Has index on `step_id`. |
| Index coverage | `idx_path_steps_path` covers the main query pattern (steps by path). `idx_path_step_contents_step` covers content lookup by step. |
| Missing index | No index on `path_step_contents.content_id` — needed if reverse lookup ("which paths contain this content?") is ever required. Not needed for current queries. |

### Data Integrity Gaps

1. **No `UNIQUE` on `(path_id, step_order)` in `path_steps`** — duplicate step orders within a path are silently accepted. The compiler avoids this via `enumerate()`, but the admin panel could create duplicates.
2. **No `UNIQUE` on `(step_id, content_id)` in `path_step_contents`** — same content could be linked to the same step twice.
3. **No `ON DELETE CASCADE`** — deleting a learning path does not cascade-delete its steps or step contents. Not an issue for the current read-only mobile model, but the admin panel would need manual cleanup.

---

## Performance Audit

### Query Patterns by Screen

| Screen | Queries | Performance Risk |
|--------|---------|-----------------|
| Home (iOS/Android) | 3: categories + paths + favorites | Low — simple indexed queries |
| Category detail | 2: subcategories + contents | Low — indexed on `category_id` |
| Path detail | 1 + N: steps + N step-content queries | **Medium** — N+1 pattern (MAJOR-3) |
| Search | 2: contents LIKE + paths LIKE | Low at current scale; LIKE without FTS degrades at 1000+ items |
| Sync | 1 HTTP + 1 DB replace | Low — ETag prevents unnecessary downloads |

### Scalability Projections

| Scale | What breaks |
|-------|------------|
| 50 items, 5 paths | Nothing — current performance is good |
| 200 items, 10 paths | Path detail with 15+ steps fires 16 queries; noticeable on older devices |
| 500 items | LIKE search starts to show latency (~50ms per query). Admin panel contents page becomes unwieldy |
| 1000+ items | FTS5 becomes necessary. Admin needs pagination. Full DB download ~2-5 MB |

---

## Platform-Specific Findings

### iOS

**Deployment target:** iOS 17 (implied by use of `ContentUnavailableView`, `onChange(of:initial:)` with two-parameter closure, `.searchable`). The target bump is clean — no deprecated API usage found.

**Strengths added in RM-3:**
- Clean async/await pattern with `withTaskGroup` for concurrent path content loading
- Proper `Task` cancellation in search debounce
- `UserState` with debounced auto-save via Combine

**Concerns:**
- MAJOR-1 (thread-safety) is the primary risk
- `URL: @retroactive Identifiable` in `SafariView.swift` is a Swift 5.9+ pattern and works, but may trigger warnings in future Swift versions

### Android

**Min SDK:** 26 (Android 8.0). Target SDK 34. Play Store now requires target 35 for new app submissions.

**Strengths added in RM-3:**
- All DB queries properly dispatched to `Dispatchers.IO`
- Search debounce via `delay(300)` with coroutine auto-cancellation
- URL encoding/decoding for navigation arguments

**Concerns:**
- MAJOR-5 (synchronous UserState saves)
- MINOR-2 (raw difficulty string in path cards)
- MINOR-6 (BuildConfig reflection — 3rd audit flagging this)
- MINOR-9 (outdated dependencies — 3rd audit flagging this)

### Go Admin Panel

**Strengths added in RM-3:**
- Complete learning path CRUD: list, detail with steps and linked contents
- Good test coverage: 30+ tests including path handlers
- Clean migration system (v1 -> v2 -> v3 with idempotent re-runs)

**Concerns:**
- MAJOR-4 (stale schema.sql files)
- N+1 query in `pathDetailHandler` (same pattern as mobile)
- No pagination (MINOR-11)
- No shared templates (MINOR-10)

### Python Build Pipeline

**Strengths added in RM-3:**
- `build_paths()` follows established patterns
- `validate.py` has cross-reference validation for path content slugs
- `backfill_thumbnails.py` is a well-designed one-shot utility

**Concerns:**
- MAJOR-2 (inefficient slug map building)
- Zero tests (MINOR-8)

---

## What RM-3 Fixed vs What Persists

### Fixed in RM-3 (across all three audits)

| Finding | First Identified | Audit # | Fixed In |
|---------|-----------------|---------|----------|
| iOS main-thread DB | Audit 1 MAJOR-2 | 1 | RM-2 |
| No conditional sync | Audit 1 MAJOR-3 | 1 | RM-2 |
| Android no back button | Audit 1 MAJOR-4 | 1 | RM-2 |
| Android no Coil | Audit 1 MAJOR-5 | 1 | RM-2 |
| Android main-thread DB | Audit 2 MAJOR-2 | 2 | RM-3 |
| Android search no debounce | Audit 2 MAJOR-3 | 2 | RM-3 |
| No user state layer | Audit 2 MAJOR-4 | 2 | RM-3 |
| person required in schema | Audit 2 MINOR-1 | 2 | RM-3 |
| thumbnail_url empty | Audit 2 MINOR-2 | 2 | RM-3 |
| Android nav URL encoding | Audit 2 MINOR-3 | 2 | RM-3 |
| No sort_order in technique | Audit 2 MINOR-5 | 2 | RM-3 |

### Persisting across audits (3rd time identified)

| Finding | First Identified | Risk Level |
|---------|-----------------|------------|
| No mobile tests | Audit 1 MINOR-11 | Growing — more untested code each release |
| No Python pipeline tests | Audit 2 MINOR-12 | Growing — pipeline is more complex now |
| Android deps outdated | Audit 1 MINOR-12 | Growing — version gap widens |
| schema_version multi-row | Audit 1 MINOR-9 | Static — cosmetic |
| No UNIQUE on source_url | Audit 1 MINOR-8 | Static — validator catches duplicates |
| BuildConfig reflection | Audit 1 MINOR-10 | Static — works but fragile |
| No web templates base | Audit 2 MINOR-6 | Growing — 9 templates now |
| No web pagination | Audit 2 MINOR-7 | Growing — more content each release |

---

## Recommended Tech EPICs for RM-4

### 1. EPIC: iOS Database Thread Safety

**Source:** MAJOR-1
**Effort:** S (1-2 days)
**Priority:** High — latent crash risk

Convert `Database` to a Swift actor, or add locking discipline. Route `replaceWith()` through the same serial queue as queries. This is small effort with high safety payoff.

### 2. EPIC: N+1 Query Elimination

**Source:** MAJOR-3
**Effort:** S (1 day per platform)
**Priority:** Medium — affects path detail performance

Add `pathAllStepContents(pathId:)` to all three platforms' Database layers. Single query replaces N queries. Also fix the Go admin's `pathDetailHandler`.

### 3. EPIC: Build Pipeline Optimization + Tests

**Source:** MAJOR-2, MINOR-8
**Effort:** S-M (2-3 days)
**Priority:** Medium — pipeline gets slower as content grows

- Eliminate `build_content_slug_map()` by collecting slug->ID during `build_contents()`
- Add smoke tests for `build.py` and `validate.py`
- Clean up the dead code in `build_content_slug_map()`

### 4. EPIC: Schema Drift Cleanup

**Source:** MAJOR-4, MINOR-4, MINOR-5
**Effort:** S (half day)
**Priority:** Low — documentation-only issue

- Update `admin/schema.sql` and `data/schema.sql` to v3
- Optionally add `UNIQUE` constraints on `source_url`, `(path_id, step_order)`, `(step_id, content_id)`
- Optionally switch `schema_version` to single-row pattern

### 5. EPIC: Android Maintenance Pass

**Source:** MAJOR-5, MINOR-2, MINOR-6, MINOR-9
**Effort:** M (3-5 days due to dependency bump)
**Priority:** Medium — targetSdk 35 required for Play Store

- Debounce UserState saves (MAJOR-5)
- Fix difficulty badge display (MINOR-2)
- Replace BuildConfig reflection with direct fields (MINOR-6)
- Bump all dependencies including compileSdk/targetSdk to 35 (MINOR-9)
- Enable R8 minification for release builds

### 6. EPIC: Test Coverage (ongoing)

**Source:** MINOR-7 (mobile), MINOR-8 (Python)
**Effort:** M (4-6 days across all platforms)
**Priority:** Medium — risk increases with each release

- iOS: Unit tests for Database queries, UserState round-trip, DeepLink computation
- Android: Same scope
- Python: Smoke test for build.py, validation edge cases
- Estimate: ~60% coverage of critical paths with 2 days per platform

---

## Key Takeaway

The RM-3 codebase is in good shape overall. The team resolved all four prior major findings and delivered a significant feature set (learning paths, user state, thumbnails, enhanced search). The five new major findings are moderate in severity — none are blocking for users today, but MAJOR-1 (iOS thread safety) is a latent crash that should be fixed before the next release. The persistent lack of test coverage (now flagged for the third consecutive audit) is the largest long-term risk: as the codebase grows, each untested feature addition increases the probability of a regression going undetected.

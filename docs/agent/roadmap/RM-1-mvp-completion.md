# RM-1 MVP Completion

## Meta
- Created: 2026-04-14
- Status: completed

## Plan
> Complete the MVP to make BMC usable end-to-end: full admin CRUD, working data sync, core UX improvements, and real content.

### ADMIN — Complete the admin panel
- [x] ADMIN-1: Add edit/delete for categories [🟢 | Medium]
- [x] ADMIN-2: Add edit/delete for contents [🟢 | Medium]
- [x] ADMIN-3: Add basic auth (password protection) [🟢 | Small]
- [x] ADMIN-4: Upload exported DB to Aliyun OSS on export [🟡 | Medium]

### SYNC — Make data sync actually work
- [x] SYNC-1: Configure real OSS bucket URL in both apps [🟢 | Small]
- [x] SYNC-2: Add sync status indicator in iOS [🟢 | Small]
- [x] SYNC-3: Add sync status indicator in Android [🟢 | Small]

### UX — Core user experience improvements
- [x] UX-1: Add search across all content (iOS + Android) [🟢 | Medium]
- [x] UX-2: Add pull-to-refresh to trigger manual sync (iOS) [🟢 | Small]
- [x] UX-3: Add pull-to-refresh to trigger manual sync (Android) [🟢 | Small]
- [x] UX-4: Add content thumbnails in list views (iOS + Android) [🟡 | Medium]
- [x] UX-5: Empty state handling (iOS + Android) [🟢 | Small]

### DATA — Content & data enrichment
- [x] DATA-1: Add more seed data (at least 5 categories with real content) [🟢 | Medium]

### DX — Developer experience
- [x] DX-1: Add admin handler tests for edit/delete [🟢 | Small]
- [x] DX-2: Add schema migration support [🟡 | Medium]

## Execution Results

### ADMIN

#### ✅ ADMIN-1: Add edit/delete for categories
- Confidence: 🟢 High
- Change size: Medium (5 files, +423 lines)
- Result: tests pass, commit `14981bf`
- Key decisions: Edit via GET/POST pattern, delete with dependency check

#### ✅ ADMIN-2: Add edit/delete for contents
- Confidence: 🟢 High
- Change size: Medium (5 files)
- Result: tests pass, commit `e575a69`
- Key decisions: Category dropdown and platform dropdown in edit form, 8 new tests

#### ✅ ADMIN-3: Add basic auth (password protection)
- Confidence: 🟢 High
- Change size: Small (2 files)
- Result: tests pass, commit `822cb6e`
- Key decisions: HTTP Basic Auth via env vars (BMC_ADMIN_USER, BMC_ADMIN_PASSWORD), timing-safe comparison

#### ✅ ADMIN-4: Upload exported DB to Aliyun OSS on export
- Confidence: 🟡 Medium
- Change size: Medium (3 files)
- Result: tests pass, commit `fbadd90`
- Key decisions: Graceful skip when OSS not configured, Aliyun OSS SDK integration

### SYNC

#### ✅ SYNC-1: Configure real OSS bucket URL in both apps
- Confidence: 🟢 High
- Change size: Small (2 files)
- Result: commit `97ed714`
- Key decisions: SyncConfig pattern on both platforms, reads from Info.plist (iOS) / BuildConfig (Android)

#### ✅ SYNC-2+3: Add sync status indicators
- Confidence: 🟢 High
- Change size: Medium (6 files)
- Result: commit `f3ec03d`
- Key decisions: SyncManager (iOS) / SyncState sealed interface (Android), auto-dismiss timers, Chinese labels

### UX

#### ✅ UX-1: Add search across all content
- Confidence: 🟢 High
- Change size: Medium (5 files)
- Result: commit `4f67952`
- Key decisions: SQL LIKE on title/summary/author_name, .searchable (iOS) / OutlinedTextField (Android)

#### ✅ UX-2+3: Add pull-to-refresh
- Confidence: 🟢 High
- Change size: Small (3 files)
- Result: commit `fcff90e`
- Key decisions: .refreshable (iOS) / PullToRefreshBox (Android), leverages existing sync state

#### ✅ UX-4: Add content thumbnails
- Confidence: 🟡 Medium
- Change size: Medium (3 files)
- Result: commit `56f5995`
- Key decisions: AsyncImage on iOS, placeholder-only on Android (Coil not added — deferred to RM-2)

#### ✅ UX-5: Empty state handling
- Confidence: 🟢 High
- Change size: Small (4 files)
- Result: commit `faba670`
- Key decisions: ContentUnavailableView (iOS) / centered Column (Android), Chinese labels

### DATA

#### ✅ DATA-1: Enrich seed data
- Confidence: 🟢 High
- Change size: Medium (4 files)
- Result: commit `93a58be`
- Key decisions: 6 top-level categories, 20 subcategories, 20 tutorials, mixed platforms/authors

### DX

#### ✅ DX-1: Admin handler tests for edit/delete
- Confidence: 🟢 High
- Change size: Small (included in ADMIN-1 and ADMIN-2)
- Result: covered by ADMIN task commits

#### ✅ DX-2: Schema migration support
- Confidence: 🟡 Medium
- Change size: Medium (3 files)
- Result: commit `90a6fb8`
- Key decisions: schema_version table, sequential migrations, auto-detect pre-migration DBs, 3 new tests

### Bug Fixes

#### ✅ Android Compose BOM version fix
- Result: commit `96ea202`
- Issue: BOM 2024.02.00 incompatible with PullToRefreshBox, updated to 2024.09.00

## Decisions Needed
None — all tasks completed.

## E2E & User Testing
- **33 Go admin tests pass**
- 2 bugs found and fixed (Compose BOM, unused import)
- **4 critical issues** identified for RM-2: Android back button, Android thumbnails, full DB download, iOS main thread queries
- **9 important issues**: no onboarding, no external link indicator, search doesn't find categories, etc.
- Full report: docs/agent/epics/UX-user-testing.md

## Product Research
- No Chinese competitor does cross-platform tutorial curation by technique
- BMC needs: personal state (favorites/history), editorial voice, content freshness
- Path to retention requires the app to become "mine" — not just a link directory
- Content library needs to grow from 20 to 100+ tutorials with real thumbnails
- Full report: docs/agent/epics/PROD-product-research.md

## Tech Audit
- **5 major findings**: two-DB architecture (critical blocker), iOS main thread queries, no conditional sync, Android back button, Android thumbnails
- **12 minor findings**: code duplication, no search debounce, no input validation, default credentials, etc.
- Key insight: two-database architecture must ship before favorites/history
- Full report: docs/agent/epics/TECH-audit.md

## Executive Review
- Review slide: docs/agent/roadmap/RM-1-review.html

## Next Roadmap
- Proposed: docs/agent/roadmap/RM-2-product-foundation.md
- 47 tasks across 10 EPICs (STAB, SYNC2, ARCH, SCHEMA, FAV, SRCH, PRSNT, HIST, DIFF, ADMIN2)
- Key themes: platform stability, two-DB architecture, personal learning state

## Handoff
N/A — roadmap completed in single session.

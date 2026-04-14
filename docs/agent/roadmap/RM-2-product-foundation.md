# RM-2 Product Foundation

## Meta
- Created: 2026-04-14
- Status: planned
- Depends on: RM-1 MVP Completion

## Plan
> Fix critical stability issues, build the two-database architecture that unlocks user features, and ship the highest-impact product improvements — favorites, better search, and content presentation polish. Goal: make BMC good enough that a real badminton player keeps it installed after the first week.

### P0 — Critical fixes (ship immediately)

#### STAB — Platform Stability
Android back button, Android thumbnails, and iOS main-thread queries are table-stakes quality issues that cause immediate churn.

- [ ] STAB-1: Add back navigation icon to Android CategoryScreen TopAppBar [🟢 | Small]
- [ ] STAB-2: Integrate Coil and wire ContentThumbnail to load thumbnailUrl on Android [🟢 | Medium]
- [ ] STAB-3: Move all iOS Database queries off the main thread (async/await + ViewModel pattern) [🟡 | Medium]
- [ ] STAB-4: Add 300ms search debounce on iOS to prevent per-keystroke queries [🟢 | Small]
- [ ] STAB-5: Consolidate duplicated ContentRow composable across Android HomeScreen and CategoryScreen [🟢 | Small]

#### SYNC2 — Smart Sync
Current sync downloads the full DB on every launch, wasting bandwidth and making startup sluggish. Conditional downloads fix this with minimal effort since OSS already provides ETag/Last-Modified headers.

- [ ] SYNC2-1: iOS — store last ETag locally, send If-None-Match on sync, handle 304 Not Modified [🟢 | Small]
- [ ] SYNC2-2: Android — store last ETag locally, send If-None-Match on sync, handle 304 Not Modified [🟢 | Small]
- [ ] SYNC2-3: Rename syncIfNeeded to reflect actual behavior; deduplicate iOS DataSync methods [🟢 | Small]

---

### P1 — Architectural foundation (enables future features)

#### ARCH — Two-Database Architecture
**Critical path.** The current sync model replaces the entire DB file, destroying any local data. Favorites and history cannot exist until content (synced) and user data (local) live in separate databases. This EPIC must ship before FAV or HIST.

- [ ] ARCH-1: Design two-database schema — content DB (synced from OSS) and user DB (local-only, never overwritten) [🟡 | Medium]
- [ ] ARCH-2: iOS — implement separate content DB and user DB with independent SQLite connections [🟡 | Medium]
- [ ] ARCH-3: Android — implement separate content DB and user DB with independent SQLite connections [🟡 | Medium]
- [ ] ARCH-4: Update sync logic on both platforms to replace only the content DB, preserving user DB [🟡 | Medium]
- [ ] ARCH-5: Add concurrency protection for iOS Database singleton (prevent crash if sync fires while user is browsing) [🟡 | Small]
- [ ] ARCH-6: Add unit tests for database layer on iOS and Android (query correctness, sync state transitions) [🟡 | Medium]

#### SCHEMA — Schema Evolution
Add fields that product features need. Difficulty levels enable filtering; editor_notes strengthens the editorial voice that is BMC's core differentiator; duration helps users choose content by time commitment.

- [ ] SCHEMA-1: Add difficulty_level TEXT column to contents table (beginner/intermediate/advanced) via migration v2 [🟢 | Small]
- [ ] SCHEMA-2: Add editor_notes TEXT column to contents table via migration v2 [🟢 | Small]
- [ ] SCHEMA-3: Add duration TEXT column to contents table via migration v2 [🟢 | Small]
- [ ] SCHEMA-4: Add UNIQUE constraint on source_url to prevent duplicate content entries [🟢 | Small]
- [ ] SCHEMA-5: Update admin panel forms to support new fields (difficulty dropdown, editor_notes textarea, duration input) [🟢 | Medium]
- [ ] SCHEMA-6: Backfill existing seed data with difficulty levels and editor's notes [🟢 | Medium]

---

### P2 — High-value product features

#### FAV — Favorites & Bookmarks
**Depends on: ARCH.** The single highest-leverage feature for retention. Without personal state, BMC is a read-only directory with zero switching cost. With favorites, it becomes "my badminton study notebook."

- [ ] FAV-1: Create favorites table in user DB (content_id, created_at) [🟢 | Small]
- [ ] FAV-2: iOS — add heart/bookmark icon on content rows; tap to toggle favorite [🟢 | Medium]
- [ ] FAV-3: Android — add heart/bookmark icon on content rows; tap to toggle favorite [🟢 | Medium]
- [ ] FAV-4: iOS — add "My Favorites" section or tab showing saved tutorials [🟢 | Medium]
- [ ] FAV-5: Android — add "My Favorites" section or tab showing saved tutorials [🟢 | Medium]

#### SRCH — Enhanced Search
Search is how returning users navigate. Today it misses categories entirely and returns results without context.

- [ ] SRCH-1: Include category names in search index on both platforms (searching "步法" surfaces the footwork category) [🟢 | Medium]
- [ ] SRCH-2: Show parent category name in search results on both platforms [🟢 | Small]
- [ ] SRCH-3: Add search debounce on Android (300ms) to match iOS [🟢 | Small]
- [ ] SRCH-4: Encode category name in Android navigation route to prevent breakage with special characters [🟢 | Small]

#### PRSNT — Content Presentation
First impression drives retention. Quick wins with outsized impact on perceived quality.

- [ ] PRSNT-1: Show content count on category rows (e.g., "杀球 (3)") on both platforms [🟢 | Small]
- [ ] PRSNT-2: Add external link indicator on content rows (external-link icon or "在B站观看" text) [🟢 | Small]
- [ ] PRSNT-3: Populate seed data with real thumbnail URLs from source platforms [🟢 | Medium]
- [ ] PRSNT-4: Add loading spinner for initial data load on first launch (both platforms) [🟢 | Small]
- [ ] PRSNT-5: Display difficulty level badge on content rows once SCHEMA-1 ships [🟢 | Small]

---

### P3 — Engagement features

#### HIST — Watch History
**Depends on: ARCH.** Track what the user has viewed. Enables future features like progress tracking and recommendations.

- [ ] HIST-1: Create watch_history table in user DB (content_id, watched_at) [🟢 | Small]
- [ ] HIST-2: iOS — record watch event when user opens an external link [🟢 | Small]
- [ ] HIST-3: Android — record watch event when user opens an external link [🟢 | Small]
- [ ] HIST-4: iOS — show visual indicator (checkmark or opacity change) on watched items [🟢 | Small]
- [ ] HIST-5: Android — show visual indicator (checkmark or opacity change) on watched items [🟢 | Small]
- [ ] HIST-6: Add "Recently Watched" section to home screen on both platforms [🟡 | Medium]

#### DIFF — Difficulty Levels
**Depends on: SCHEMA-1.** A beginner searching for 杀球 tutorials needs fundamentally different content than an advanced player. Difficulty filtering eliminates this friction.

- [ ] DIFF-1: iOS — add filter chips on category screen (All / Beginner / Intermediate / Advanced) [🟢 | Medium]
- [ ] DIFF-2: Android — add filter chips on category screen (All / Beginner / Intermediate / Advanced) [🟢 | Medium]
- [ ] DIFF-3: Include difficulty level in search results display [🟢 | Small]

#### ADMIN2 — Admin Panel Improvements
Speed of content library growth is directly tied to admin workflow efficiency.

- [ ] ADMIN2-1: Make OSS upload asynchronous in export handler (goroutine instead of blocking call) [🟢 | Small]
- [ ] ADMIN2-2: Add server-side input validation with clear error messages for content and category forms [🟡 | Medium]
- [ ] ADMIN2-3: Make admin panel mobile-responsive (responsive table layout, touch-friendly controls) [🟡 | Medium]
- [ ] ADMIN2-4: Show hierarchical category dropdown in "Add Content" form (indent subcategories under parents) [🟢 | Small]

---

## Dependency Graph

```
STAB ──┐
SYNC2 ─┤
       ├──> ARCH ──> FAV
SCHEMA ┤          ──> HIST
       │
       ├──> SRCH (independent)
       ├──> PRSNT (independent, PRSNT-5 depends on SCHEMA-1)
       ├──> DIFF (depends on SCHEMA-1)
       └──> ADMIN2 (independent)
```

**Ship order:**
1. STAB + SYNC2 (P0 — unblock everything)
2. ARCH + SCHEMA (P1 — architectural foundation, in parallel)
3. FAV + SRCH + PRSNT (P2 — user-facing features, after ARCH lands)
4. HIST + DIFF + ADMIN2 (P3 — engagement and tooling)

## Execution Results

## Decisions Needed
- ARCH: Should user DB use a separate SQLite file, or a separate schema/table prefix within the same file with selective sync? Separate file is simpler but means two open connections.
- SCHEMA: Should editor_notes be a separate field or repurpose the existing summary field? Separate field preserves the distinction between video description and editorial commentary.
- FAV: Should favorites sync across devices (requires account system) or remain device-local? Device-local is simpler and sufficient for RM-2.

## E2E & User Testing

## Product Research

## Tech Audit

## Executive Review

## Next Roadmap
Candidates for RM-3: Learning Paths, Content Detail Preview, Fresh Content Notifications, Onboarding, Share & Invite Flow, FTS5 search, admin pagination, content bulk import.

## Handoff

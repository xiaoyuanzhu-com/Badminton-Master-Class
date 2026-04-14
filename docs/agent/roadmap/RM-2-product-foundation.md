# RM-2 Product Foundation

## Meta
- Created: 2026-04-14
- Status: planning
- Depends on: RM-1 MVP Completion

## Plan
> Pivot to content-as-code: content lives as plain files in the repo, compiles to SQLite at build time, and an agent handles ingestion from URLs. The web panel becomes a read-only client like iOS/Android. Fix critical stability issues and add deep linking for smooth UX. Goal: make content management effortless and the browsing experience polished.

### P0 — The Pivot: Content-as-Code

#### FILE — Content File Schema & Migration
Content moves from SQLite seed data to structured files in `data/`. Git becomes the source of truth — diffable, reviewable, agent-writable. The folder hierarchy mirrors the category taxonomy.

```
data/
  content/
    techniques/
      _category.json            # { "name": "基本功", "icon": "🏸", "sort_order": 1 }
      forehand/
        _category.json          # { "name": "正手", "icon": "💪", "sort_order": 1 }
        forehand-clear.json     # content item
        forehand-clear.png      # thumbnail (optional, matched by filename)
      backhand/
        _category.json
        backhand-clear.json
        backhand-clear.png
    doubles/
      _category.json
      positioning/
        _category.json
        ...
  schema.json                   # JSON Schema for validation
  build.sh                      # compiles to bmc.db
```

- [ ] FILE-1: Design and document the file schema — `_category.json` for categories, `{slug}.json` for content items, `{slug}.png` for thumbnails [🟢 | Small]
- [ ] FILE-2: Create JSON Schema (`data/schema.json`) for validating content and category files [🟢 | Small]
- [ ] FILE-3: Migrate existing seed.sql data into the new file structure [🟢 | Medium]
- [ ] FILE-4: Add a linter/validator script that checks all files: valid JSON, valid platform enum, thumbnails exist if referenced, no duplicate source_urls [🟢 | Medium]
- [ ] FILE-5: Remove seed.sql (replaced by file structure); keep schema.sql as reference for the compiler target [🟢 | Small]

Content item JSON schema:
```json
{
  "title": "反手高远球三步学会",
  "summary": "从零开始学反手高远球",
  "source_url": "https://bilibili.com/video/BV1xxx",
  "source_platform": "bilibili",
  "author_name": "惠程俊",
  "difficulty": "beginner",
  "duration": "12:30",
  "editor_notes": "最清晰的反手教学之一",
  "sort_order": 1
}
```

#### BUILD — SQLite Compiler
A build script walks `data/content/` and produces `bmc.db`. This replaces the admin export flow. The compiled DB is the artifact that gets uploaded to OSS and bundled in apps.

- [ ] BUILD-1: Write compiler script (Go or Python) that walks `data/content/`, reads all `_category.json` and `*.json` files, produces `bmc.db` matching the existing schema [🟢 | Medium]
- [ ] BUILD-2: Handle thumbnails — embed as blob or generate URL mapping; decide on thumbnail distribution strategy (bundled in DB vs. separate CDN) [🟡 | Medium]
- [ ] BUILD-3: Auto-copy compiled `bmc.db` to `ios/BadmintonMasterClass/Resources/` and `android/app/src/main/assets/` [🟢 | Small]
- [ ] BUILD-4: Add OSS upload step — after compile, upload `bmc.db` to configured OSS bucket (reuse existing OSS logic from admin) [🟢 | Small]
- [ ] BUILD-5: Integrate with git hooks or Makefile — `make build` compiles + copies + validates [🟢 | Small]

#### WEB — Convert Admin to Web Client
The admin panel becomes a read-only web client — a third browsing surface alongside iOS and Android. It reads from the compiled SQLite, same as the apps. No more CRUD handlers.

- [ ] WEB-1: Strip all create/edit/delete handlers; keep list and detail views as read-only [🟢 | Medium]
- [ ] WEB-2: Add a content detail page (title, summary, thumbnail, editor's notes, platform badge, link to source) [🟢 | Medium]
- [ ] WEB-3: Add search on the web client (reuse the SQL LIKE pattern from mobile) [🟢 | Small]
- [ ] WEB-4: Serve from compiled `bmc.db` instead of the admin's live DB [🟢 | Small]
- [ ] WEB-5: Clean up auth — read-only client doesn't need basic auth (or make it optional for private preview) [🟢 | Small]

### P0 — Stability & UX Polish (carry over)

#### STAB — Platform Stability
Table-stakes quality fixes.

- [ ] STAB-1: Add back navigation icon to Android CategoryScreen TopAppBar [🟢 | Small]
- [ ] STAB-2: Integrate Coil and wire ContentThumbnail to load thumbnailUrl on Android [🟢 | Medium]
- [ ] STAB-3: Move all iOS Database queries off the main thread (async/await + ViewModel pattern) [🟡 | Medium]
- [ ] STAB-4: Add 300ms search debounce on iOS to prevent per-keystroke queries [🟢 | Small]
- [ ] STAB-5: Consolidate duplicated ContentRow composable across Android HomeScreen and CategoryScreen [🟢 | Small]

#### JUMP — Smart Deep Linking
Open content in the native app (Bilibili, Douyin, etc.) instead of a web view. Computed at runtime from `source_url` + `source_platform`.

- [ ] JUMP-1: Add deep link URL computation from source_url + source_platform (bilibili://, snssdk1128://, xhsdiscover://, youtube://) — shared utility on each platform [🟢 | Small]
- [ ] JUMP-2: iOS — try opening deep link via UIApplication.open; fall back to SFSafariViewController if app not installed [🟢 | Small]
- [ ] JUMP-3: Android — try deep link Intent; fall back to Custom Tab if app not installed [🟢 | Small]
- [ ] JUMP-4: Add WeChat article handling (weixin:// or direct web fallback) [🟡 | Small]

#### SYNC2 — Smart Sync
Conditional downloads to avoid re-downloading unchanged DB on every launch.

- [ ] SYNC2-1: iOS — store last ETag locally, send If-None-Match on sync, handle 304 Not Modified [🟢 | Small]
- [ ] SYNC2-2: Android — store last ETag locally, send If-None-Match on sync, handle 304 Not Modified [🟢 | Small]
- [ ] SYNC2-3: Deduplicate sync code on both platforms [🟢 | Small]

---

### P1 — Content Ingestion Agent

#### INGEST — Agent-Powered Content Ingestion
The primary way new content enters the system. Share a URL → agent fetches metadata → writes files → commits. Human reviews the diff.

- [ ] INGEST-1: Build agent definition that accepts a URL and target category path [🟢 | Medium]
- [ ] INGEST-2: Platform-specific metadata fetchers — extract title, author, thumbnail, duration from Bilibili, Douyin, Xiaohongshu, YouTube page/API [🟡 | Large]
- [ ] INGEST-3: Auto-detect platform from URL domain [🟢 | Small]
- [ ] INGEST-4: Download and save thumbnail as `{slug}.png` next to the content JSON [🟢 | Small]
- [ ] INGEST-5: Generate slug from title (pinyin or transliteration) [🟢 | Small]
- [ ] INGEST-6: If category path not provided, suggest best-fit category based on title/content keywords [🟡 | Medium]
- [ ] INGEST-7: Validate output with the linter from FILE-4 before committing [🟢 | Small]

---

### P2 — User Features

#### ARCH — Two-Database Architecture
**Critical path for favorites/history.** Content DB is now a build artifact (synced from OSS). User DB is local-only (favorites, history). Sync replaces only the content DB.

- [ ] ARCH-1: Design two-database schema — content DB (compiled, synced) and user DB (local, never overwritten) [🟡 | Medium]
- [ ] ARCH-2: iOS — implement separate content DB and user DB with independent SQLite connections [🟡 | Medium]
- [ ] ARCH-3: Android — implement separate content DB and user DB with independent SQLite connections [🟡 | Medium]
- [ ] ARCH-4: Update sync logic on both platforms to replace only the content DB, preserving user DB [🟡 | Medium]
- [ ] ARCH-5: Add concurrency protection for iOS Database singleton [🟡 | Small]

#### FAV — Favorites & Bookmarks
**Depends on: ARCH.** Personal state that makes BMC "my" app.

- [ ] FAV-1: Create favorites table in user DB (content_id, created_at) [🟢 | Small]
- [ ] FAV-2: iOS — add heart icon on content rows; tap to toggle favorite [🟢 | Medium]
- [ ] FAV-3: Android — add heart icon on content rows; tap to toggle favorite [🟢 | Medium]
- [ ] FAV-4: iOS — add "My Favorites" section or tab [🟢 | Medium]
- [ ] FAV-5: Android — add "My Favorites" section or tab [🟢 | Medium]

#### SRCH — Enhanced Search
- [ ] SRCH-1: Include category names in search results on both platforms [🟢 | Medium]
- [ ] SRCH-2: Show parent category name in search results [🟢 | Small]
- [ ] SRCH-3: Add search debounce on Android (300ms) [🟢 | Small]
- [ ] SRCH-4: Encode category name in Android navigation route [🟢 | Small]

#### PRSNT — Content Presentation
- [ ] PRSNT-1: Show content count on category rows [🟢 | Small]
- [ ] PRSNT-2: Add external link indicator (platform icon + "在B站观看") [🟢 | Small]
- [ ] PRSNT-3: Display difficulty badge and duration on content rows [🟢 | Small]
- [ ] PRSNT-4: Display editor's notes in content detail or expanded row [🟢 | Small]

---

### P3 — Engagement Features

#### HIST — Watch History
**Depends on: ARCH.**

- [ ] HIST-1: Create watch_history table in user DB [🟢 | Small]
- [ ] HIST-2: iOS — record watch event when opening external link [🟢 | Small]
- [ ] HIST-3: Android — record watch event when opening external link [🟢 | Small]
- [ ] HIST-4: Show visual indicator on watched items (both platforms) [🟢 | Small]
- [ ] HIST-5: Add "Recently Watched" section to home screen [🟡 | Medium]

#### DIFF — Difficulty Levels
- [ ] DIFF-1: iOS — add filter chips (All / Beginner / Intermediate / Advanced) [🟢 | Medium]
- [ ] DIFF-2: Android — add filter chips [🟢 | Medium]
- [ ] DIFF-3: Include difficulty in search results [🟢 | Small]

---

## Dependency Graph

```
FILE ──> BUILD ──> WEB (reads compiled DB)
                ──> OSS upload ──> mobile sync
STAB ──┐
SYNC2 ─┤ (independent, ship early)
JUMP ──┘

INGEST (depends on FILE + BUILD for validation/compilation)

ARCH ──> FAV
     ──> HIST

SRCH, PRSNT, DIFF (independent of ARCH, can ship in parallel)
```

**Ship order:**
1. FILE + BUILD + WEB + STAB + JUMP + SYNC2 (P0 — the pivot + stability)
2. INGEST (P1 — agent ingestion, depends on FILE/BUILD)
3. ARCH + FAV + SRCH + PRSNT (P2 — user features)
4. HIST + DIFF (P3 — engagement)

## Execution Results

## Decisions Needed
- BUILD-2: Thumbnails — bundle as blobs in SQLite, or keep as separate files served from CDN/OSS? Blobs keep distribution simple (single file) but increase DB size. CDN is better at scale but adds infra.
- ARCH: Separate SQLite file vs. separate tables in same file for user DB? Separate file is simpler for sync (replace content DB without touching user DB).
- INGEST-2: Which platforms need API access vs. HTML scraping? Bilibili has a public API; others may need page scraping or yt-dlp.

## E2E & User Testing

## Product Research

## Tech Audit

## Executive Review

## Next Roadmap
Candidates for RM-3: Learning Paths, Curated Playlists, Onboarding Flow, Share & Invite, FTS5 Search, Content Freshness Notifications, Bulk Import Tool.

## Handoff

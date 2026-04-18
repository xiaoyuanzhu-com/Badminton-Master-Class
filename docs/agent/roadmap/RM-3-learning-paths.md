# RM-3 Learning Paths

## Meta
- Created: 2026-04-14
- Status: planning
- Depends on: RM-2 Product Foundation

## Plan
> Turn BMC from a link directory into a structured learning companion. Ship bug fixes from RM-2 testing, build a thumbnail pipeline so the app looks finished, introduce learning paths as content-as-code (the differentiating feature), add lightweight user state for favorites and progress tracking, grow the library to 100+ items, and polish discovery and presentation. Goal: a badminton player can follow a "Beginner 30-Day Plan" built from free community videos — something no Chinese-language product offers today.

### P0 — Bug Fixes & Quick Wins

#### FIX — RM-2 Testing Bug Fixes
Ship the critical fixes surfaced by RM-2 user testing and tech audit. These are prerequisites — without them, new features land on a shaky foundation.

- [ ] FIX-1: Android search debounce — add `delay(300)` in `HomeScreen.kt` `LaunchedEffect` to match iOS behavior [🟢 | Small]
- [ ] FIX-2: Android DB queries off main thread — wrap all `Database` calls in `withContext(Dispatchers.IO)` across `HomeScreen.kt` and `CategoryScreen.kt` [🟢 | Small]
- [ ] FIX-3: URL-encode category name in Android nav route (or pass ID only and look up name from DB) [🟢 | Small]
- [ ] FIX-4: Handle b23.tv Bilibili short links in deep linking — resolve redirect or match short format separately [🟢 | Small]
- [ ] FIX-5: Make `person` field optional in `content.schema.json` (remove from `required` array; compiler already handles null) [🟢 | Small]
- [ ] FIX-6: Handle Xiaohongshu `/discovery/` path and `xhslink.com` short URLs in deep linking [🟢 | Small]
- [ ] FIX-7: Add `sort_order` field to `technique.schema.json` so curators control category display order [🟢 | Small]

**Confidence:** 🟢 High — all tasks are small, well-scoped, and clearly identified by testing.
**Total effort:** Small. Ship first, before any new features.

#### THUMB — Thumbnail Pipeline
Gray placeholders are the single biggest visual weakness. Surface platform thumbnail URLs in the compiled DB so content items display real images.

- [ ] THUMB-1: Add `thumbnail_url` field to `content.schema.json` and update `ingest.py` to write it from platform metadata [🟢 | Small]
- [ ] THUMB-2: Update `build.py` to write `thumbnail_url` from content JSON into the compiled DB (currently hardcoded empty) [🟢 | Small]
- [ ] THUMB-3: Quick win — use source platform thumbnail URLs directly in the DB (Bilibili/YouTube URLs are stable and public) [🟢 | Small]
- [ ] THUMB-4: CDN pipeline — during `make upload`, download thumbnails and upload to OSS; rewrite URLs in DB to CDN paths [🟡 | Medium]

**Confidence:** 🟢 High for THUMB-1 through THUMB-3 (straightforward data plumbing). 🟡 Medium for THUMB-4 (CDN infra dependency).
**Total effort:** Small for the quick-win path (THUMB-1/2/3). Medium if CDN pipeline is included.
**Ship THUMB-3 first** as the quick win. CDN upload (THUMB-4) can follow when needed for platforms with unstable thumbnail URLs.

---

### P1 — The Big Features

#### PATH — Learning Paths as Content-as-Code
The single highest-value feature BMC can build. No Chinese-language product offers free, structured badminton learning paths built from community content. A path is just another file in the repo — a JSON array of content slugs with editorial notes. The same build pipeline compiles it.

- [ ] PATH-1: Design path file schema (`data/content/schemas/path.schema.json`) — title, summary, difficulty, ordered steps with day/title/note/content-slugs [🟢 | Small]
- [ ] PATH-2: Update `validate.py` to validate path files — content slugs must exist, no duplicates, schema compliance [🟢 | Small]
- [ ] PATH-3: Update `build.py` to compile `learning_paths` and `path_steps` tables into `bmc.db`; bump `SCHEMA_VERSION` to 3 [🟡 | Medium]
- [ ] PATH-4: Update `admin/migrate.go` with v3 migration for `learning_paths` and `path_steps` tables [🟢 | Small]
- [ ] PATH-5: Create 2–3 starter paths: "Beginner 30-Day Plan", "Smash Mastery", "Doubles Positioning" [🟡 | Medium — content work]
- [ ] PATH-6: Web client — learning paths list page and path detail page with step-by-step view [🟡 | Medium]
- [ ] PATH-7: iOS — learning paths section on home screen; path detail view with step list [🟡 | Medium]
- [ ] PATH-8: Android — learning paths section on home screen; path detail view with step list [🟡 | Medium]

**Confidence:** 🟡 Medium — schema/compiler work follows established patterns (🟢), but UI on three surfaces and content authoring add scope.
**Total effort:** Large (sum of parts). The schema/compiler phase (PATH-1 through PATH-4) is Medium; UI phase (PATH-6 through PATH-8) is Medium; content (PATH-5) is parallel.
**Depends on:** FIX (stable base), THUMB (paths look bad with gray placeholders).

#### STATE — Lightweight User State
Instead of the full two-database architecture from RM-2, ship a minimal personal state system using a JSON file in the app's documents directory. Supports favorites and path progress. Upgrade to SQLite user DB later if needed.

- [ ] STATE-1: iOS — `UserState` class that reads/writes a JSON file in documents directory (favorites list, path progress map) [🟢 | Small]
- [ ] STATE-2: Android — `UserState` class with the same JSON-file approach in app-internal storage [🟢 | Small]
- [ ] STATE-3: iOS — heart icon on content rows, tap to toggle favorite, persists to UserState [🟡 | Medium]
- [ ] STATE-4: Android — heart icon on content rows, tap to toggle favorite, persists to UserState [🟡 | Medium]
- [ ] STATE-5: iOS — "My Favorites" section accessible from home screen [🟢 | Small]
- [ ] STATE-6: Android — "My Favorites" section accessible from home screen [🟢 | Small]
- [ ] STATE-7: Path progress tracking — mark steps as completed, persists to UserState JSON [🟡 | Medium]
- [ ] STATE-8: Visual progress indicator on path cards (e.g., "4/12 完成") [🟢 | Small]
- [ ] STATE-9: Verify sync (`Database.replaceWith`) does not touch the user-state file (different file path — should be safe, but verify) [🟢 | Small]

**Confidence:** 🟢 High for STATE-1/2 (simple file I/O). 🟡 Medium for favorites UI (STATE-3/4) and progress tracking (STATE-7/8) — touches multiple views.
**Total effort:** Medium.
**Depends on:** PATH (progress tracking needs paths to exist). STATE-1/2 and favorites (STATE-3 through STATE-6) can start in parallel with PATH.

---

### P2 — Content & Discovery

#### GROW — Content Library Growth
BMC needs to grow from 27 to 100+ items to be credible. Some categories have only 2–3 items — not enough to demonstrate editorial value. This is a content operations effort that runs in parallel with engineering.

- [ ] GROW-1: Define target content plan — 20+ items per top-level category (120+ total across techniques, doubles, fitness, etc.) [🟢 | Small — planning]
- [ ] GROW-2: Batch ingestion sessions using `ingest.py` — add 10–20 items per session [🟡 | Medium — content work]
- [ ] GROW-3: Write genuine editor's notes for every item (not video descriptions — *why* this was selected) [🟡 | Medium — content work]
- [ ] GROW-4: Add 2–3 new people (creators) with full profiles [🟢 | Small — content work]
- [ ] GROW-5: Ensure every difficulty level is represented in every category [🟢 | Small — content work]

**Confidence:** 🟢 High — the ingestion pipeline exists; this is operational effort.
**Total effort:** Medium (calendar time, not engineering complexity).
**Depends on:** THUMB (so new content has visible thumbnails from day one).

#### SRCH — Enhanced Search
More important now with a larger library and learning paths to find.

- [ ] SRCH-1: Include category names in mobile search results (join `categories` table in search query) [🟡 | Medium]
- [ ] SRCH-2: Show category name badge in mobile search result rows [🟢 | Small]
- [ ] SRCH-3: Include learning paths in search results [🟢 | Small]
- [ ] SRCH-4: Web client — add direct source link in search results (N6 from RM-2 testing) [🟢 | Small]

**Confidence:** 🟢 High — well-understood SQL and UI changes.
**Total effort:** Small to Medium.
**Depends on:** PATH (for SRCH-3, paths must exist to be searchable).

#### PRSNT — Content Presentation
Surface the rich metadata that already exists in the content files but is not yet displayed on mobile.

- [ ] PRSNT-1: Show content count on mobile category rows (e.g., "12 个内容") [🟢 | Small]
- [ ] PRSNT-2: Display difficulty badge and duration on content rows [🟢 | Small]
- [ ] PRSNT-3: Display editor's notes in content detail or expanded row [🟢 | Small]
- [ ] PRSNT-4: Add external link indicator on content rows (platform icon + "在B站观看") [🟢 | Small]

**Confidence:** 🟢 High — data already in the DB, just needs UI.
**Total effort:** Small.

---

### P3 — Polish

#### ONBOARD — First-Run Experience
Still no onboarding after two iterations. More valuable now that learning paths give users a clear starting action.

- [ ] ONBOARD-1: Single-screen onboarding: value proposition + "Start a Learning Path" CTA [🟢 | Small]
- [ ] ONBOARD-2: Optional skill level selection (beginner/intermediate/advanced) → suggest a matching learning path [🟡 | Medium]

**Confidence:** 🟡 Medium — design decisions needed for ONBOARD-2.
**Total effort:** Small to Medium.
**Depends on:** PATH (onboarding points users to learning paths).

#### SHARE — Share Flow
Enable the word-of-mouth moment: "Follow this 30-day plan."

- [ ] SHARE-1: Share a content item (deep link to BMC or fallback to source URL) on iOS and Android [🟡 | Medium]
- [ ] SHARE-2: Share a learning path as a card image (for WeChat Moments) [🟡 | Medium]
- [ ] SHARE-3: "Invite a practice partner" flow [🟢 | Small]

**Confidence:** 🟡 Medium — platform share sheet APIs are straightforward, but WeChat card image generation (SHARE-2) is non-trivial.
**Total effort:** Medium.
**Depends on:** PATH (sharing paths is the high-value scenario), STATE (share progress like "4/12 完成").

#### WEBPOL — Web Client Polish
The web client works but needs structural improvements to scale with more content and content types.

- [ ] WEBPOL-1: Extract shared header/nav/CSS into Go base template (`{{ template "header" }}`) [🟡 | Medium]
- [ ] WEBPOL-2: Add pagination to content list (`?page=1&per_page=20`) [🟢 | Small]
- [ ] WEBPOL-3: Learning paths section on web home page [🟡 | Medium]

**Confidence:** 🟢 High — standard web patterns.
**Total effort:** Medium.
**Depends on:** PATH (for WEBPOL-3, paths must be in the DB).

---

## Dependency Graph

```
FIX ─────────────────────────────────────────────> (ship first, independent)
THUMB ───────────────────────────────────────────> (ship alongside FIX)

PATH ──> PATH-1..4 (schema/compiler) ──> PATH-5 (content) ──> PATH-6..8 (UI)
     \                                                        /
      ── depends on: FIX (stable base), THUMB (visual quality)

STATE ──> STATE-1..2 (JSON persistence layer)
      ──> STATE-3..6 (favorites — can start in parallel with PATH)
      ──> STATE-7..8 (path progress — depends on PATH UI)
      ──> STATE-9 (sync safety verification)

GROW ──> depends on: THUMB (thumbnails from day one)
     ──> runs in parallel with PATH and STATE (content work, not engineering)

SRCH ──> SRCH-3 depends on PATH (paths must exist to search)
PRSNT ──> independent, can ship anytime after FIX

ONBOARD ──> depends on PATH (onboarding points to learning paths)
SHARE ──> depends on PATH + STATE (share paths with progress)
WEBPOL ──> WEBPOL-3 depends on PATH
```

## Ship Order

| Phase | EPICs | Goal | Rough Timing |
|-------|-------|------|--------------|
| **Phase 1** | FIX + THUMB | Stable base, real thumbnails, no gray placeholders | Week 1–2 |
| **Phase 2** | PATH (schema/compiler/content) + GROW (start batch ingestion) | Learning paths exist as data; library growing toward 100 items | Week 3–4 |
| **Phase 3** | PATH (UI on all 3 surfaces) + STATE (persistence + favorites) | Users can browse paths, save favorites, see thumbnails | Week 5–6 |
| **Phase 4** | STATE (path progress) + SRCH + PRSNT | Track progress through paths; richer discovery and presentation | Week 7–8 |
| **Stretch** | ONBOARD + SHARE + WEBPOL | First-run flow, word-of-mouth sharing, web polish | If time permits |

**Critical path:** FIX → THUMB → PATH (schema) → PATH (UI) → STATE (progress)

---

## Execution Results

### FIX (8/8 ✅)
- FIX-0 `b2831d5` | FIX-1 `b55eb27` | FIX-2 `eb8236c` | FIX-3 `c33b974`
- FIX-4 `10d7ae0` | FIX-5 `ddfef95` | FIX-6 `02ef5c9` | FIX-7 `d19d81c`

### THUMB (3/3 ✅)
- THUMB-1/2/3 `f01fd1e`

### PATH (8/8 ✅)
- PATH-1/2/3/4 `73b1cdc` | PATH-5 `63b1231` | PATH-6 `6470636` | PATH-7 `861f66c` | PATH-8 `cbd4f35`

### STATE (9/9 ✅)
- STATE-1/3/5 `5b01ade` | STATE-2/4/6 `c9df82c` | STATE-7/8 `cf54e3d` | STATE-9 verified safe

### SRCH (4/4 ✅)
- SRCH-1/2/3 `c0834ac` — Category badges, learning path search (iOS + Android)
- SRCH-4 `a22f3e3` — Source links in web search results

### PRSNT (4/4 ✅)
- PRSNT-1/2/3/4 `ee4a8fb` — Content counts, difficulty badges, editor's notes, platform indicators (iOS + Android)

## Decisions Needed
- THUMB: Quick-win (platform URLs in DB) vs. CDN pipeline? Recommend shipping THUMB-3 first, CDN later. Bilibili and YouTube thumbnail URLs are stable.
- PATH: How many starter paths at launch? 2–3 seems right — one per difficulty level.
- STATE: JSON file vs. UserDefaults/SharedPreferences for small state? JSON file is more flexible and portable; recommend that.
- GROW: Target 100 or 120+ items? Depends on content availability per category. Define per-category minimums.
- SHARE: Is WeChat card image generation worth the effort in RM-3, or defer to RM-4?

## E2E & User Testing

### E2E Results
- Python validate + build: ✅ PASS (26 categories, 4 people, 20 contents, 3 paths, schema v3)
- Go admin tests: ✅ 29/29 PASS
- iOS build: ✅ PASS (iPhone 16, iOS 18.1)
- Android build: ⚠️ BLOCKED (no JDK on Mac mini; needs `brew install openjdk@17`)
- DB integrity: ✅ 3 paths, 24 steps, all FK references valid, 8/20 thumbnails
- No bugs found in code. Zero fixes needed.

### User Testing Findings (as intermediate badminton player)
**Critical:**
- Zero content items have `duration` or `editor_notes` populated — UI features are dead code for current dataset
- Only 8/20 (40%) content items have thumbnails

**Important:**
- No onboarding or skill-level filter
- Empty subcategories visible (7 of 26 have no content)
- Search doesn't query category names or editor notes
- Web search doesn't include learning paths (mobile does)
- Android shows raw English difficulty strings on path cards ("beginner" not "入门")
- No progress bar inside path detail view (only on home screen cards)
- No dedicated My Favorites page
- Web templates have inconsistent header colors (search.html uses Google blue, not Ink Black)

**Nice-to-have:**
- No "app not installed" messaging for deep link fallback
- No "continue where you left off" in paths
- No congratulations on path completion

## Product Research
See `docs/agent/epics/PROD-research-rm4.md`. Top 5 proposed EPICs:
1. GROW — Content scale-up (20→120+) [P0]
2. HISTORY — Watch history & engagement tracking [P0]
3. FRESH — Content freshness & re-engagement signals [P1]
4. SHARE — Social sharing & organic growth [P1]
5. STREAK — Practice streaks & daily goals [P2]

## Tech Audit
See `docs/agent/epics/TECH-audit-rm3.md`. 5 major findings:
1. iOS Database singleton not thread-safe (race condition on sync + query)
2. `build_content_slug_map()` inefficient (re-reads all files, N queries)
3. N+1 query pattern in path detail loading (all 3 platforms)
4. `admin/schema.sql` and `data/schema.sql` stale (still v2)
5. Android UserState saves synchronously on every mutation (no debounce)
11 minor findings. 6 proposed tech EPICs for RM-4.

## Executive Review
Link to review slide: docs/agent/roadmap/RM-3-review.html

## Next Roadmap
Link to proposed next roadmap: docs/agent/roadmap/RM-4-content-and-engagement.md

## Handoff

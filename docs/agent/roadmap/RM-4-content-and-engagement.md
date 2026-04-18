# RM-4 Content & Engagement

## Meta
- Created: 2026-04-15
- Status: executing
- Depends on: RM-3 Learning Paths

## Plan
> Fill the library to 120+ items so the category structure delivers on its promises, fix the critical bugs and metadata gaps surfaced by RM-3 user testing, resolve the highest-risk tech debt (iOS thread safety, Android UserState debounce, schema drift, N+1 queries), then layer on engagement features — watch history, freshness signals, social sharing — that give users reasons to come back and share. Goal: transform BMC from a well-built tool into a product that grows through content scale, engagement loops, and word-of-mouth sharing.

---

### P0 — Phase 0: Browsing Stabilization (BLOCKING)

A pre-execution audit of basic browsing on both iOS and Android (2026-04-18) found that while the foundation is architecturally sound, several visible cracks would erode user trust if shown today: 60% missing thumbnails, 30% empty subcategory dead-ends, English difficulty labels on Android, no progress bar in path detail, and a latent iOS crash via Database race condition. Phase 0 closes these before product expansion begins.

#### STAB — Browsing Stabilization
Resequenced subset of META + BUGFIX + IOSFIX, plus 4 newly-surfaced gaps from the audit. All small. Lands as a tight bundle before HISTORY/FRESH/SHARE.

- [ ] STAB-1: IOSFIX-1 — Convert iOS `Database` to actor or route all ops through serial queue (eliminate `replaceWith` race) [🟢 | Medium]
- [ ] STAB-2: IOSFIX-2 — Verify with concurrent sync + path detail load test [🟢 | Small]
- [ ] STAB-3: BUGFIX-1 — Android path difficulty labels: replace raw English with `ContentDifficultyBadge` in `LearningPathCard` and `SearchPathRow` [🟢 | Small]
- [ ] STAB-4: BUGFIX-3 — Add progress bar to path detail view on iOS and Android [🟢 | Small]
- [ ] STAB-5: BUGFIX-4 — Hide empty subcategories from browse UI on iOS and Android (filter at query or render layer) [🟢 | Small]
- [ ] STAB-6: META-3 — Backfill `thumbnail_url` for the 12 content items currently missing one [🟢 | Small]
- [ ] STAB-7: META-1/2 — Backfill `duration` and write genuine `editor_notes` for all 20 existing items [🟡 | Medium — content work]
- [ ] STAB-8: NEW — iOS `contents()` and `pathStepContents()` queries: JOIN categories so `categoryName` populates in browse mode (currently only set in search) [🟢 | Small]
- [ ] STAB-9: NEW — iOS `HomeView.refreshable` reloads `favoriteItems` after sync (prevents stale rows) [🟢 | Small]
- [ ] STAB-10: NEW — Android first-launch loading spinner during DB asset copy (avoid showing "暂无内容" empty state) [🟢 | Small]
- [ ] STAB-11: NEW — Install JDK 17 on Mac mini, then verify Android build compiles and runs on device [🟢 | Small]
- [ ] STAB-12: BUGFIX-2 — Web `search.html` header color alignment (Ink Black) [🟢 | Small]

**Confidence:** 🟢 High — all tasks are small, well-scoped, and identified by audit.
**Total effort:** Small-Medium bundle. Lands first, before any RM-4 product expansion.
**Source:** Audit on 2026-04-18 covering both platforms (see `docs/agent/epics/STAB-browsing-stabilization.md`).

---

### P0 — Critical Fixes from RM-3 Testing (deferred — superseded by Phase 0)
> The original META, BUGFIX, IOSFIX, SCHEMA EPICs below are preserved for context. STAB above subsumes their critical items. SCHEMA remains as its own follow-up.

#### META — Content Metadata Backfill
RM-3 user testing revealed that `duration`, `editor_notes`, and `thumbnail_url` fields are mostly empty — making the UI features that display them dead code. This must be fixed before any new content work.

- [ ] META-1: Backfill `duration` for all 20 existing content items by checking source videos [🟢 | Small]
- [ ] META-2: Write genuine `editor_notes` for all 20 existing items — "why this was selected" not "what this video is about" [🟡 | Medium]
- [ ] META-3: Backfill remaining 12 missing `thumbnail_url` values (currently 8/20 populated) [🟢 | Small]

**Confidence:** 🟢 High — straightforward content work using existing tooling.
**Total effort:** Small-Medium.

#### BUGFIX — RM-3 User Testing Bug Fixes
Ship the bugs and inconsistencies caught during RM-3 user testing. These erode trust and must be fixed before new features.

- [ ] BUGFIX-1: Fix Android difficulty label localization — path cards show "beginner" instead of "入门"; use `ContentDifficultyBadge` composable or extract mapping to shared function [🟢 | Small]
- [ ] BUGFIX-2: Fix web template styling inconsistency — `search.html` uses Google blue header instead of Ink Black; align all templates [🟢 | Small]
- [ ] BUGFIX-3: Add progress bar to path detail views on iOS and Android (currently only on home screen cards) [🟢 | Small]
- [ ] BUGFIX-4: Hide empty subcategories from browse UI, or show a "coming soon" indicator instead of blank screens [🟢 | Small]

**Confidence:** 🟢 High — all issues well-identified with known file locations.
**Total effort:** Small.

---

### P0 — Content Scale-Up

#### GROW — Content Library Growth (20 → 120+)
7 of 26 subcategories are empty — 27% of the app is broken promises. The taxonomy implies breadth that the content does not deliver. Growing to 120+ items with full coverage is the single highest-leverage investment for RM-4.

- [ ] GROW-1: Define per-category content targets — minimum 3 items per subcategory, 8+ for popular categories (basics, attack) [🟢 | Small — planning]
- [ ] GROW-2: Fill 7 empty subcategories: `slice-drop`, `net-block`, `drive-block`, `pounce`, `rotation`, `serve-receive`, `flexibility` (14+ items minimum) [🟡 | Medium — content]
- [ ] GROW-3: Deepen basics (grip, serve, clear, footwork) and attack (smash, drop) to 8+ items each [🟡 | Medium — content]
- [ ] GROW-4: Add new top-level category: **Tactics** (战术) with subcategories for match analysis, rally patterns, game IQ [🟡 | Medium — content + schema]
- [ ] GROW-5: Add new top-level category: **Gear** (装备) with subcategories for rackets, strings, shoes [🟢 | Small — content + schema]
- [ ] GROW-6: Add 6+ new creators with full person profiles (grow from 4 to 10+) [🟢 | Small — content]
- [ ] GROW-7: Write genuine editor's notes for all new items [🟡 | Medium — content]
- [ ] GROW-8: Add 1-2 new learning paths using newly ingested content (e.g., "Defensive Fundamentals", "Gear Guide for Beginners") [🟡 | Medium — content]

**Confidence:** 🟢 High — ingestion pipeline exists and is proven; this is operational effort.
**Total effort:** Large (calendar time, not engineering complexity). Runs continuously in parallel with engineering work.
**Depends on:** META (existing items must have complete metadata before adding more).

---

### P1 — Product EPICs

#### HISTORY — Watch History & Engagement Tracking
The most impactful missing feature. With 120+ items and learning paths, users need passive tracking to remember where they left off. Favorites (active intent) + watch history (passive tracking) + path progress (structured progression) create a complete engagement loop.

- [ ] HISTORY-1: Extend UserState JSON to include `watch_history` array (content slug + timestamp) [🟢 | Small]
- [ ] HISTORY-2: iOS — record watch event when opening external link; show "Recently Watched" section on home [🟡 | Medium]
- [ ] HISTORY-3: Android — same as HISTORY-2 [🟡 | Medium]
- [ ] HISTORY-4: Visual indicator (checkmark or reduced opacity) on already-watched items in all list views [🟢 | Small]
- [ ] HISTORY-5: "Continue Learning" — on path detail screen, highlight the next unwatched step [🟢 | Small]
- [ ] HISTORY-6: Web client — watch history section on home page [🟢 | Small]

**Confidence:** 🟡 Medium — UserState extension follows established patterns (🟢), but UI touches multiple views across 3 surfaces.
**Total effort:** Medium.
**Depends on:** STATE (RM-3, already shipped).
**Unlocks:** STREAK (streaks need engagement tracking), FRESH ("continue learning" needs history), SHARE (share progress).

#### FRESH — Content Freshness & Re-engagement
A content app that looks identical every visit is a dead app. With GROW adding 100+ items, freshness signals create a pull to re-engage.

- [ ] FRESH-1: Add `created_at` field to content schema (or derive from git commit date during build); write to compiled DB [🟢 | Small]
- [ ] FRESH-2: "New" badge on content items added within last 14 days (iOS, Android, web) [🟢 | Small]
- [ ] FRESH-3: "New This Week" section on home screen, showing items added in the last 7 days [🟡 | Medium]
- [ ] FRESH-4: Track last-open timestamp in UserState; show app badge count of new items since last open (iOS, Android) [🟡 | Medium]
- [ ] FRESH-5: Compiler: generate `content_changelog` table from git history (added/updated items per week) [🟡 | Medium]

**Confidence:** 🟡 Medium — schema/compiler work is straightforward; FRESH-5 (git history parsing) adds complexity.
**Total effort:** Medium.
**Depends on:** GROW (freshness signals without fresh content are meaningless).
**Unlocks:** Push notifications (future), email digest (future).

#### SHARE — Social Sharing & Organic Growth
Badminton players organize in WeChat groups and share tips constantly. Without a share flow, BMC's best content gets shared as raw Bilibili URLs, bypassing BMC entirely. Sharing is the primary organic acquisition channel for a niche app.

- [ ] SHARE-1: iOS — share button on content detail: native share sheet with BMC deep link + fallback source URL [🟡 | Medium]
- [ ] SHARE-2: Android — same as SHARE-1 [🟡 | Medium]
- [ ] SHARE-3: iOS — share learning path as card image (title, step count, progress if applicable, QR code to BMC) [🟡 | Medium]
- [ ] SHARE-4: Android — same as SHARE-3 [🟡 | Medium]
- [ ] SHARE-5: Web — share URL with Open Graph meta tags so links preview nicely in WeChat/iMessage [🟢 | Small]
- [ ] SHARE-6: Deep link handler: receiving a shared BMC link opens the correct content or path (extends existing deep linking) [🟡 | Medium]

**Confidence:** 🟡 Medium — share sheet APIs are straightforward; WeChat card image generation (SHARE-3/4) is non-trivial.
**Total effort:** Medium-Large.
**Depends on:** HISTORY (for progress sharing in shared cards).
**Unlocks:** Organic acquisition, social proof, word-of-mouth growth.

---

### P1 — Tech Debt

#### IOSFIX — iOS Database Thread Safety
**Source:** TECH-audit-rm3 MAJOR-1 — latent crash risk.

The `Database` singleton uses a single `OpaquePointer` accessed from both the main thread (`replaceWith()` via MainActor) and a background `queryQueue`. A sync during path detail viewing (which fires 3+ concurrent queries via `withTaskGroup`) opens a crash timing window.

- [ ] IOSFIX-1: Convert `Database` to a Swift actor, or route all operations (including `replaceWith`) through the serial `queryQueue` with locking discipline [🟡 | Medium]
- [ ] IOSFIX-2: Verify fix with concurrent sync + path detail load test scenario [🟢 | Small]

**Confidence:** 🟢 High — well-understood concurrency pattern.
**Total effort:** Small (1-2 days).
**Priority:** High — latent crash. Fix before shipping any new features.

#### NPLUS1 — N+1 Query Elimination
**Source:** TECH-audit-rm3 MAJOR-3 — path detail fires 1 + N queries (N = number of steps).

- [ ] NPLUS1-1: Add `pathAllStepContents(pathId:)` method to iOS Database — single JOIN query across `path_steps`, `path_step_contents`, `contents`; group by step ID in Swift [🟢 | Small]
- [ ] NPLUS1-2: Same for Android Database [🟢 | Small]
- [ ] NPLUS1-3: Same for Go admin `pathDetailHandler` [🟢 | Small]

**Confidence:** 🟢 High — standard SQL optimization.
**Total effort:** Small (1 day per platform).

#### SCHEMA — Schema Drift Cleanup
**Source:** TECH-audit-rm3 MAJOR-4 + MINOR-4/5 — `admin/schema.sql` and `data/schema.sql` still labeled v2, missing v3 tables.

- [ ] SCHEMA-1: Update both `admin/schema.sql` and `data/schema.sql` to match full v3 schema from `build.py`; update version comment [🟢 | Small]
- [ ] SCHEMA-2: Add `UNIQUE` constraints on `source_url`, `(path_id, step_order)`, `(step_id, content_id)` [🟢 | Small]

**Confidence:** 🟢 High — documentation and constraint additions.
**Total effort:** Small (half day).

#### DROID — Android Maintenance Pass
**Source:** TECH-audit-rm3 MAJOR-5 + MINOR-2/6/9 — synchronous UserState saves, raw difficulty strings, BuildConfig reflection, outdated dependencies.

- [ ] DROID-1: Debounce UserState saves — post save to `Dispatchers.IO` coroutine with 300ms debounce delay (match iOS pattern) [🟢 | Small]
- [ ] DROID-2: Replace BuildConfig reflection in `SyncConfig` with direct `BuildConfig` fields or hardcoded constants [🟢 | Small]
- [ ] DROID-3: Bump compileSdk/targetSdk to 35 (Play Store requirement), Compose BOM, Kotlin, Coil 3.x, Navigation 2.8.x [🟡 | Medium]
- [ ] DROID-4: Enable R8 minification for release builds [🟢 | Small]

**Confidence:** 🟡 Medium — dependency bump (DROID-3) may require migration effort proportional to the 18-month version gap.
**Total effort:** Medium (3-5 days).

#### PIPELINE — Build Pipeline Optimization
**Source:** TECH-audit-rm3 MAJOR-2 + MINOR-8 — `build_content_slug_map()` re-reads all files and issues N queries; no tests.

- [ ] PIPELINE-1: Eliminate `build_content_slug_map()` by collecting `{slug: cur.lastrowid}` during `build_contents()` insert loop; remove dead code [🟢 | Small]
- [ ] PIPELINE-2: Add smoke test for `build.py` — run build, open `bmc.db`, verify row counts and schema version [🟡 | Medium]
- [ ] PIPELINE-3: Add validation edge-case tests for `validate.py` [🟢 | Small]

**Confidence:** 🟢 High — straightforward refactoring and test authoring.
**Total effort:** Small-Medium (2-3 days).

---

### P2 — Nice-to-Haves

#### ONBOARD — First-Run Experience
Deferred from RM-3. More impactful now that the library is larger and learning paths provide a clear starting action.

- [ ] ONBOARD-1: Single-screen onboarding: value proposition + "Start a Learning Path" CTA [🟢 | Small]
- [ ] ONBOARD-2: Optional skill level selection (beginner/intermediate/advanced) → suggest a matching learning path [🟡 | Medium]

**Confidence:** 🟡 Medium — design decisions needed for ONBOARD-2.
**Total effort:** Small-Medium.
**Depends on:** GROW (onboarding is more compelling with 120+ items and 5+ paths).

#### STREAK — Practice Streaks & Daily Goals
The single most effective retention mechanic in learning apps (Duolingo, Yousician, Nike Training Club). Tying streaks to learning path progress creates a daily habit.

- [ ] STREAK-1: Define streak rules: what counts as a "practice day" (open a tutorial? complete a path step? either?) [🟢 | Small — design decision]
- [ ] STREAK-2: Extend UserState JSON with streak data (current streak, longest streak, last active date) [🟢 | Small]
- [ ] STREAK-3: iOS — streak display on home screen (flame icon + day count) [🟡 | Medium]
- [ ] STREAK-4: Android — same as STREAK-3 [🟡 | Medium]
- [ ] STREAK-5: "Weekly Summary" — end-of-week review showing tutorials watched, path steps completed, streak status [🟡 | Medium]
- [ ] STREAK-6: Shareable streak card ("I've practiced badminton for 14 days straight") [🟢 | Small]

**Confidence:** 🟡 Medium — straightforward state management but touches multiple views; design decisions on streak rules.
**Total effort:** Medium.
**Depends on:** HISTORY (streaks require knowing when user last engaged).

#### WEBSRCH — Web Learning Path Search
RM-3 user testing found that web search does not include learning paths (mobile does).

- [ ] WEBSRCH-1: Include learning paths in web search results alongside content items [🟢 | Small]

**Confidence:** 🟢 High — mobile implementation exists as reference.
**Total effort:** Small.

#### WEBPOL — Web Client Polish
Structural improvements for the admin panel to scale with 120+ items.

- [ ] WEBPOL-1: Extract shared header/nav/CSS into Go base template (`{{ template "header" }}`) [🟡 | Medium]
- [ ] WEBPOL-2: Add pagination to content list (`?page=1&per_page=20`) [🟢 | Small]

**Confidence:** 🟢 High — standard web patterns.
**Total effort:** Small-Medium.

---

## Dependency Graph

```
META ──────────────────────────────────────────> (backfill existing content first)
BUGFIX ────────────────────────────────────────> (ship alongside META, independent)
IOSFIX ────────────────────────────────────────> (fix before any new features)

GROW ──> depends on META (existing items clean before adding more)
     ──> runs continuously in parallel with all engineering work

HISTORY ──> depends on STATE (RM-3, shipped)
        ──> unlocks: FRESH, STREAK, SHARE

FRESH ──> depends on GROW (freshness needs fresh content)
      ──> depends on HISTORY ("continue learning" integration)

SHARE ──> depends on HISTORY (progress sharing)
      ──> enhanced by GROW (more content = more shareable moments)

STREAK ──> depends on HISTORY (streak needs engagement tracking)
       ──> feeds into SHARE (shareable streak cards)

Tech debt (NPLUS1, SCHEMA, DROID, PIPELINE) ──> independent of product EPICs
                                              ──> can ship in parallel

ONBOARD ──> depends on GROW (compelling with larger library)
WEBSRCH / WEBPOL ──> independent, ship anytime
```

## Ship Order

| Phase | EPICs | Goal | Rough Timing |
|-------|-------|------|--------------|
| **Phase 1** | META + BUGFIX + IOSFIX + SCHEMA | Clean foundation: metadata complete, RM-3 bugs fixed, iOS crash risk eliminated, schema docs current | Week 1-2 |
| **Phase 2** | GROW (start) + NPLUS1 + DROID + PIPELINE | Content growth begins; tech debt cleared across all platforms | Week 3-5 |
| **Phase 3** | HISTORY + GROW (ongoing) | Watch history on all surfaces; library approaching 80+ items | Week 6-8 |
| **Phase 4** | FRESH + SHARE | Freshness signals live; sharing flow enables organic growth | Week 9-11 |
| **Stretch** | ONBOARD + STREAK + WEBSRCH + WEBPOL | First-run flow, retention mechanics, web polish | If time permits |

**Critical path:** META → GROW → FRESH (freshness needs content)
**Engineering critical path:** IOSFIX → HISTORY → SHARE (engagement stack builds sequentially)

---

## Decisions Needed

1. **GROW: New top-level categories** — Add Tactics (战术) and Gear (装备) as proposed? Or keep existing 6 categories and add subcategories only? New categories require schema changes and UI updates on 3 surfaces.

2. **HISTORY: Storage model** — Extend the existing UserState JSON with watch history, or create a separate history file? At 120+ items viewed multiple times, the history array could grow large. May need a cap (last 100 entries) or rotation strategy.

3. **FRESH: `created_at` source** — Derive from git commit date during build (automatic but couples to git) or add as an explicit field in content JSON (manual but decoupled)? Git-derived is recommended for accuracy.

4. **SHARE: WeChat card images** — Generate card images natively on-device (SwiftUI/Compose render + screenshot) or server-side? On-device is simpler but produces variable quality. Server-side needs infrastructure BMC does not have.

5. **DROID: Dependency bump scope** — Bump all dependencies at once (DROID-3 is a large changeset) or incrementally? Incremental is safer but slower. Recommend a single dedicated branch with thorough testing.

6. **STREAK: Streak definition** — What counts as a "practice day"? Opening any tutorial? Completing a learning path step? Either? This affects how aggressive the streak feels — too easy and it is meaningless, too hard and users break it quickly.

7. **ONBOARD: Scope** — Ship a minimal single-screen onboarding (ONBOARD-1 only) as a quick win, or wait for the fuller skill-level-selection flow (ONBOARD-2)? Recommend shipping ONBOARD-1 early and iterating.

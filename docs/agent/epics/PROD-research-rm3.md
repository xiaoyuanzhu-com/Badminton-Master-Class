# Product Research — RM-3 EPIC Proposals

**Date:** 2026-04-14
**Scope:** Strategic analysis of what BMC needs next, based on RM-2 testing feedback and competitive positioning
**Inputs:** RM-2 user testing report, RM-1 product research, RM-2 roadmap (deferred items), current codebase state

---

## Current State Summary

After RM-2, BMC has a solid technical foundation:
- **27 content items** across 6 technique categories, **4 people**, compiled from plain files to SQLite
- Build pipeline (`make build`), validation, ETag-based sync, deep linking
- Three browsing surfaces: iOS, Android, web client
- An ingestion workflow exists in the roadmap but no `ingest.py` script was found in the repo

What BMC does *not* have: any personal state (favorites, history), any learning structure (paths, progressions), any content freshness signals, and no thumbnails (every item shows a gray placeholder).

**The honest product assessment:** BMC is a well-engineered link directory. A badminton player would browse it once, maybe bookmark a few videos in their browser, and have no reason to return. The RM-1 research nailed this: "They probably wouldn't open BMC instead of Bilibili directly."

---

## Strategic Questions Answered

### 1. What is the single most impactful thing to build next?

**Learning paths, not favorites.**

The RM-1 research identified "personal state" (favorites/history) as the retention unlock. That was correct for a traditional content app. But BMC's content-as-code architecture opens a different door: **learning paths as files**.

Here is the reasoning:

- **Favorites solve a problem users don't have.** With 27 items across 6 categories, users can see everything in one scroll. There is nothing to "save for later" — the library is small enough to browse entirely. Favorites become valuable at 200+ items; building the feature now adds complexity without payoff.
- **Learning paths solve a problem users *do* have.** "I want to get better at badminton, but I don't know what to practice in what order." No Chinese-language product answers this. Badminton Famly charges $5.99/mo for their Training Plan Builder. BMC can offer it for free using curated community content.
- **Learning paths give the content-as-code architecture its killer feature.** A path is just another file in the repo — a JSON array of content slugs with editorial notes. The same build pipeline compiles it. The same ingestion workflow could help populate it. This is where the architecture pays off.
- **Learning paths create the word-of-mouth moment.** "I found this app with a 30-day beginner plan that uses free Bilibili videos" is something a player would actually tell their practice partner. "I found an app that lists 27 badminton videos" is not.

Favorites and history should still be built, but as supporting infrastructure *after* the content is worth favoriting — and after learning paths give users something to track progress against.

### 2. Is the two-DB architecture still the right approach?

**Yes, but scope it down.** The RM-2 roadmap designed a full two-database architecture (ARCH-1 through ARCH-5) with separate SQLite connections, concurrency protection, and updated sync logic. This is correct in principle — the content DB is a replaceable artifact, and user state must survive sync.

However, the initial implementation can be simpler:

- **Phase 1: Single user-state file, not a second database.** Favorites and path progress are small data sets. A JSON file in the app's documents directory (or UserDefaults/SharedPreferences for very small state) avoids the complexity of managing two SQLite connections. Sync already replaces only the content DB file; the user-state file is never touched.
- **Phase 2: Upgrade to SQLite user DB when the data model grows.** Once watch history, notes, or cross-device sync are on the table, migrate the JSON to a proper database.

This phased approach ships personal state in days instead of weeks, and avoids premature architecture for data that starts as a handful of content IDs.

### 3. With content-as-code, what new possibilities open up?

The content-as-code architecture is BMC's most underappreciated asset. It enables:

1. **Learning paths as files.** A `data/content/paths/` directory with JSON files that reference content slugs in order. The compiler builds a `learning_paths` table. Zero new infrastructure.

2. **Collections / playlists as files.** "Best videos for doubles positioning" or "Weekend drill session" — curated groups that cross category boundaries. Same file-based pattern.

3. **Content tags as a flat file.** A tag taxonomy (e.g., `footwork`, `wrist-technique`, `deception`, `power`) that cross-cuts the category tree. One JSON file mapping tags to content slugs. Enables discovery by theme, not just by technique hierarchy.

4. **Contributor-friendly content growth.** The file format is simple enough that a knowledgeable badminton player (not a developer) could submit content via a GitHub issue template or a simple form that generates the JSON. The validation pipeline catches errors.

5. **Changelog / "what's new" from git.** Since content changes are commits, `git log --since="7 days ago" -- data/content/` generates a "new this week" feed automatically. The compiler could write a `changelog` table with recently added/updated items.

### 4. What is the next step to make content growth easy?

The ingestion workflow is the bottleneck. Currently: manually create a JSON file, ensure the person exists, run `make build`, verify. This is too many steps.

**The next step is a working `ingest.py` CLI** that does:
1. Accept a URL
2. Auto-detect platform from domain
3. Fetch page metadata (title, thumbnail URL, author name)
4. Match or create the person file
5. Write the content JSON to the correct category (prompted or auto-suggested)
6. Run validation
7. Output a summary for human review

This was INGEST in the RM-2 roadmap but was not built. It should be the first engineering task of RM-3, because every other content initiative (learning paths, collections, growing to 100+ items) depends on efficient content addition.

**Second step: thumbnail pipeline.** Every content item in the app shows a gray placeholder. This is the single biggest visual weakness. The ingestion script should save `thumbnail_url` from platform metadata, and the build pipeline should either embed URLs in the DB or upload images to CDN during `make upload`.

### 5. What would make a badminton player recommend this app to a friend?

Three scenarios, in order of likelihood:

1. **"Follow this 30-day plan"** — A structured learning path that turns scattered YouTube/Bilibili videos into a curriculum. The recommendation is the path, not the app. ("My coach shared this beginner plan with me, it's all free videos but organized really well.")

2. **"This app has the best X tutorial I've found"** — Strong editorial voice with editor's notes that explain *why* each video was selected. The user trusts the curation. ("I spent an hour searching for smash tutorials, then someone told me about BMC and the top pick was exactly what I needed.")

3. **"Track your practice with me"** — Social progress sharing. ("I've completed 8/12 lessons in the footwork path, how far are you?") This requires learning paths + progress tracking + share functionality.

Scenario 1 is achievable in RM-3. Scenario 2 requires content scale (100+ items with quality editor's notes). Scenario 3 is RM-4+.

---

## Proposed EPICs for RM-3

### P0 — Bug Fixes & Debt (carry from RM-2 testing)

#### EPIC: FIX — RM-2 Testing Fixes

Ship the P0 items from the RM-2 user testing report. These are prerequisites — without them, new features land on a shaky foundation.

| # | Task | Effort |
|---|------|--------|
| FIX-1 | Android search debounce — add `delay(300)` in `HomeScreen.kt` `LaunchedEffect` | S |
| FIX-2 | Android DB queries off main thread — `withContext(Dispatchers.IO)` wrappers | S |
| FIX-3 | URL-encode category name in Android nav route (or pass ID only) | S |
| FIX-4 | Handle b23.tv Bilibili short links in deep linking | S |
| FIX-5 | Make `person` field optional in content schema (schema says required, compiler handles null) | S |
| FIX-6 | Handle Xiaohongshu `/discovery/` and `xhslink.com` short URLs in deep linking | S |
| FIX-7 | Add category sort_order to `_technique.json` so curators control display order | S |

**Total effort: Small. Ship first, before any new features.**

---

### P0 — Content Pipeline (unlocks everything else)

#### EPIC: INGEST — Content Ingestion Script

The ingestion script was designed in RM-2 but not built. It is the single biggest bottleneck to content growth. Without it, adding content is manual and error-prone, and BMC stays at 27 items.

| # | Task | Effort |
|---|------|--------|
| INGEST-1 | CLI that accepts a URL + optional category path | S |
| INGEST-2 | Auto-detect platform from URL domain | S |
| INGEST-3 | Fetch page metadata: title, author, thumbnail URL, duration (Bilibili API, HTML scraping for others) | M |
| INGEST-4 | Match existing person by name or create new person file | S |
| INGEST-5 | Generate slug from title (pinyin transliteration) | S |
| INGEST-6 | Write content JSON + validate with existing `validate.py` | S |
| INGEST-7 | Save thumbnail URL in content file (for compiler to include in DB) | S |

**Depends on:** Nothing (file schema and build pipeline already exist).
**Unlocks:** Rapid content growth, learning path population, thumbnail pipeline.

#### EPIC: THUMB — Thumbnail Pipeline

Gray placeholders are the single biggest visual weakness. With the ingestion script writing `thumbnail_url` from platform metadata, the pipeline needs to surface them.

| # | Task | Effort |
|---|------|--------|
| THUMB-1 | Update `build.py` to write `thumbnail_url` from content JSON into the compiled DB (currently hardcoded empty) | S |
| THUMB-2 | During `make build`, download thumbnails and upload to OSS/CDN; rewrite URLs in DB to CDN paths | M |
| THUMB-3 | Fallback: if no CDN, use source platform thumbnail URLs directly in the DB (works for Bilibili/YouTube; may break for Douyin/Xiaohongshu) | S |

**Ship THUMB-3 first** as a quick win — Bilibili and YouTube thumbnail URLs are stable and publicly accessible. CDN upload (THUMB-2) can follow.

---

### P1 — The Differentiator

#### EPIC: PATH — Learning Paths

This is the single highest-value feature BMC can build. No Chinese-language product offers free, structured badminton learning paths built from community content.

**File format** — a new `data/content/paths/` directory:

```json
// data/content/paths/beginner-30-day.json
{
  "title": "羽毛球入门30天计划",
  "summary": "从零基础到能打一场完整比赛",
  "difficulty": "beginner",
  "steps": [
    {
      "day": 1,
      "title": "握拍与站姿",
      "note": "先别急着打球，花20分钟把握拍方式练对",
      "content": ["correct-grip", "ready-stance"]
    },
    {
      "day": 2,
      "title": "正手发高远球",
      "note": "发球是每个回合的开始，练好发球等于赢在起跑线",
      "content": ["forehand-serve-high"]
    }
  ]
}
```

| # | Task | Effort |
|---|------|--------|
| PATH-1 | Design path file schema (JSON Schema in `data/content/schemas/path.schema.json`) | S |
| PATH-2 | Create 2-3 starter paths: "Beginner 30-Day Plan", "Smash Mastery", "Doubles Positioning" | M (content work) |
| PATH-3 | Update `validate.py` to validate path files (content slugs exist, no duplicates) | S |
| PATH-4 | Update `build.py` to compile `learning_paths` and `path_steps` tables into `bmc.db` | M |
| PATH-5 | Web client — learning paths list page and path detail page with step-by-step view | M |
| PATH-6 | iOS — learning paths tab or section on home screen; path detail with step list | M |
| PATH-7 | Android — learning paths tab or section on home screen; path detail with step list | M |
| PATH-8 | Progress tracking — local state (JSON file or UserDefaults/SharedPreferences) to mark steps as completed | M |
| PATH-9 | Visual progress indicator on path cards (e.g., "4/12 完成") | S |

**Depends on:** INGEST (to populate content referenced by paths), THUMB (paths look bad with gray placeholders).
**Unlocks:** Word-of-mouth growth (shareable paths), retention (progress tracking), future social features.

---

### P1 — Personal State (lightweight)

#### EPIC: STATE — User State (Lightweight)

Instead of the full two-database architecture from RM-2, ship a minimal personal state system that stores favorites and path progress. Upgrade to SQLite user DB later if needed.

| # | Task | Effort |
|---|------|--------|
| STATE-1 | iOS — `UserState` class that reads/writes a JSON file in documents directory (favorites list, path progress) | S |
| STATE-2 | Android — `UserState` class with the same JSON-file approach | S |
| STATE-3 | iOS — heart icon on content rows, tap to toggle favorite, persists to UserState | M |
| STATE-4 | Android — heart icon on content rows, tap to toggle favorite, persists to UserState | M |
| STATE-5 | iOS — "My Favorites" section accessible from home screen | S |
| STATE-6 | Android — "My Favorites" section accessible from home screen | S |
| STATE-7 | Verify sync (`Database.replaceWith`) does not touch the user-state file (it shouldn't — different file path) | S |

**Why JSON file instead of second SQLite DB:**
- Favorites and path progress are tiny data (arrays of content IDs + timestamps)
- No query complexity — just read/write the whole file
- Avoids two-database connection management, concurrency issues, and migration headaches
- Can upgrade to SQLite later when watch history or cross-device sync requires it

---

### P2 — Content Quality & Discovery

#### EPIC: GROW — Content Library Growth

BMC needs to grow from 27 to 100+ items to be credible. This is a content operations effort, not engineering, but it depends on INGEST being built.

| # | Task | Effort |
|---|------|--------|
| GROW-1 | Define target: 20+ items per top-level category (120+ total) | Content planning |
| GROW-2 | Use ingestion script to add content — batch sessions of 10-20 items | Content work |
| GROW-3 | Write genuine editor's notes for every item (not just video descriptions — *why* this was selected) | Content work |
| GROW-4 | Add 2-3 new people (creators) with full profiles | Content work |
| GROW-5 | Ensure every difficulty level is represented in every category | Content work |

#### EPIC: SRCH — Enhanced Search

Carry forward from RM-2, now more important with a larger library.

| # | Task | Effort |
|---|------|--------|
| SRCH-1 | Include category names in mobile search (searching "步法" surfaces the footwork category) | M |
| SRCH-2 | Show category name badge in mobile search results (join `categories` table) | S |
| SRCH-3 | Include learning paths in search results | S |
| SRCH-4 | Web client: add direct source link in search results (N6 from RM-2 testing) | S |

#### EPIC: FRESH — Content Freshness Signals

Give users a reason to re-open BMC after initial browse.

| # | Task | Effort |
|---|------|--------|
| FRESH-1 | Compiler: add `created_at` field to content files (or derive from git history) and include in DB | S |
| FRESH-2 | "New" badge on content items added within last 14 days | S |
| FRESH-3 | "New This Week" section on home screen (both mobile and web) | M |
| FRESH-4 | App badge / indicator when new content exists since last open (compare DB version or content count) | M |

---

### P3 — Polish & Foundation for RM-4

#### EPIC: ONBOARD — First-Run Experience

| # | Task | Effort |
|---|------|--------|
| ONBOARD-1 | Single-screen onboarding: value proposition + category overview + "Start Learning" CTA | S |
| ONBOARD-2 | Optional skill level selection (beginner/intermediate/advanced) → filters default view | M |

#### EPIC: SHARE — Share & Invite

| # | Task | Effort |
|---|------|--------|
| SHARE-1 | Share a content item (deep link to BMC or fallback to source URL) | M |
| SHARE-2 | Share a learning path as a card image (for WeChat Moments) | M |
| SHARE-3 | "Invite a practice partner" flow | S |

#### EPIC: WEBPOL — Web Client Polish

| # | Task | Effort |
|---|------|--------|
| WEBPOL-1 | Extract shared header/nav/CSS into base template (`{{ template "header" }}`) | M |
| WEBPOL-2 | Add pagination to content list (`?page=1&per_page=20`) | S |
| WEBPOL-3 | Learning paths section on web home page | M |

---

## Dependency Graph

```
FIX ──────────────────────────────────────────> (ship first, independent)

INGEST ──> THUMB ──> GROW (content operations)
       ──> PATH ──> STATE (path progress depends on user state)
                ──> SRCH (paths searchable)
                ──> FRESH (new path steps show as fresh)

STATE ──> PATH-8, PATH-9 (progress tracking)
      ──> SHARE (share what you've favorited/completed)

GROW ──> ONBOARD (onboarding is more useful with 100+ items)
     ──> SRCH (search matters more with a larger library)
```

## Recommended Ship Order

| Phase | EPICs | Goal |
|-------|-------|------|
| **Week 1-2** | FIX + INGEST | Stable base + content pipeline working |
| **Week 3-4** | THUMB + PATH (schema, compiler, starter paths) + GROW (start batch ingestion) | Content visible with thumbnails; first learning paths exist |
| **Week 5-6** | PATH (mobile + web UI) + STATE | Users can browse paths, track progress, save favorites |
| **Week 7-8** | SRCH + FRESH + polish | Discovery and re-engagement; library at 100+ items |
| **Stretch** | ONBOARD + SHARE + WEBPOL | Growth features if time permits |

---

## Key Strategic Takeaways

1. **Learning paths are BMC's "Badminton Famly for free" pitch.** This is the feature that makes BMC worth recommending. It transforms the app from "list of links" to "structured learning companion." Build it before anything else that is new.

2. **The ingestion script is the content bottleneck.** It was designed in RM-2 but not built. Every content initiative — paths, growing the library, thumbnails — is blocked until adding content is fast and reliable. Ship it first.

3. **Thumbnails have outsized visual impact.** Every item showing a gray placeholder makes BMC look unfinished. Pulling thumbnail URLs from platform metadata and writing them into the DB is a quick win that transforms perceived quality.

4. **Skip the full two-DB architecture for now.** A JSON file for user state is sufficient for favorites and path progress. The engineering effort saved can go toward learning paths, which have higher user impact.

5. **Content scale is as important as features.** 27 items across 6 categories means some categories have 2-3 items. That is not enough to demonstrate editorial value. Getting to 100+ items with genuine editor's notes is a content operations priority that runs in parallel with engineering.

6. **The content-as-code architecture is the moat.** Learning paths as files, collections as files, changelogs from git history — these are all possible *because* content is plain files in a repo. Every new content type follows the same pattern: define a schema, write files, extend the compiler. This is the strategic advantage of the RM-2 pivot.

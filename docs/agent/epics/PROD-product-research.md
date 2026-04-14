# Product Research & RM-2 EPIC Proposals — Badminton Master Class

**Date:** 2026-04-14
**Scope:** Product strategy, competitive analysis, roadmap proposals for RM-2
**Inputs:** UX user testing report, design doc, competitive landscape research

---

## 1. Competitive Landscape

### Direct competitors

| Product | Model | Strengths | Weaknesses |
|---------|-------|-----------|------------|
| **爱羽客 (Aiyuke)** | Free + ads, self-produced content | Large content library, live event coverage, community forum, club finder | Heavy monetization (ads everywhere), dated UI, content is mostly self-produced so it has a specific editorial voice. Tries to do everything (news, events, gear, community) which dilutes the teaching focus. |
| **Badminton Famly+** | Subscription ($5.99/mo) | Structured curriculum, Training Plan Builder, offline viewing, 2200+ videos, content for all levels from kids to advanced | English-only, subscription paywall, no Chinese content creators, web-first experience |
| **中羽在线 (BadmintonCN)** | Community forum | Massive Chinese-speaking community, decades of accumulated knowledge in forum threads | Forum format makes tutorials hard to find, no structured learning paths, UX is a traditional BBS |
| **友练羽毛球** | Smart hardware + app | Sensor-driven data analysis, personalized training plans, technique-specific drills | Requires hardware purchase, niche audience, limited free content |

### Indirect competitors (where users actually learn today)

| Platform | Why users go there | Why they leave |
|----------|-------------------|----------------|
| **Bilibili** | Richest badminton tutorial library in Chinese; creators like 杨晨大神, 惠程俊, 李宇轩 produce high-quality content | Algorithm-driven feed buries older gems; no technique-based organization; hard to build a systematic study plan; autoplay leads to unrelated content |
| **抖音/Douyin** | Short, punchy technique tips; easy to consume | Extremely short attention span format; content disappears from feed; no way to organize or revisit systematically |
| **小红书** | Good for gear reviews, beginner tips, visual guides | Search-driven discovery only; no curriculum structure; mixed with lifestyle content |
| **YouTube** | International coaching content (Badminton Insight, Tobias Wadenka) | Requires VPN in China; English-language barrier for target users |

### The gap BMC fills

No product in the Chinese market does **cross-platform content curation organized by technique taxonomy**. The closest analog is a "Goodreads for badminton tutorials" — the value is not in hosting content, but in answering: *"What is the single best video to learn X technique?"*

---

## 2. Strategic Analysis

### Why would someone open BMC instead of Bilibili directly?

Today, the honest answer is: **they probably wouldn't**. BMC currently functions as a link directory — 20 items organized into categories. That is useful exactly once (to discover the links) and has no reason for repeat visits.

To become a daily-open app, BMC needs to offer at least one of:

1. **Curation authority** — "BMC picked this as the best 杀球 tutorial" carries editorial weight that a Bilibili search does not. This requires visible editorial quality signals (ratings, editor's notes, "why this video") and a growing reputation.

2. **Personal progress** — "I've studied 12 of 20 基本功 tutorials" gives users a sense of progression. Bilibili has no concept of a personal badminton learning journey.

3. **Fresh content cadence** — "3 new tutorials added this week" gives users a reason to check back. A static library has no pull.

4. **Learning structure** — "Follow this 30-day beginner plan" is something no Chinese platform offers. Badminton Famly charges $5.99/mo for exactly this.

### What is the content moat?

The moat is **editorial taste + organizational structure**, not content volume. BMC will never have more content than Bilibili. The value is that someone with domain expertise has watched hundreds of videos and selected the best 3-5 per technique. This is analogous to how Wirecutter beats Amazon search — fewer options, but each one is vetted.

To strengthen this moat:
- Every content item should have an **editor's note** explaining *why* it was selected (the `summary` field does this partially, but it reads more like a video description than an editorial pick)
- Add **difficulty level** so users self-select appropriately
- Add **"what you'll learn"** tags so the curation feels intentional, not random

### How could BMC grow beyond a link directory?

Three strategic directions, ordered by feasibility:

1. **Learning companion** (near-term) — Favorites, watch history, learning paths, progress tracking. Makes BMC the user's personal badminton study tool.

2. **Community curation** (medium-term) — Let advanced users submit and vote on tutorials. Scale curation beyond one editor. Think "Product Hunt for badminton videos."

3. **Original micro-content** (long-term) — Short technique GIFs, annotated slow-motion breakdowns, text guides that complement the video links. Content that only exists in BMC.

---

## 3. Proposed EPICs for RM-2

### P0 — Critical fixes (must ship before any growth effort)

These come directly from the user testing report. Without fixing these, new users will churn immediately.

#### EPIC: Platform Stability

**What:** Fix the four critical issues from user testing:
1. Android back navigation in CategoryScreen
2. Android thumbnail loading (integrate Coil)
3. Conditional sync (ETag/If-Modified-Since to avoid re-downloading unchanged database)
4. iOS main-thread database queries (move to async/await)

**Why:** These are table-stakes quality issues. No back button on Android is a blocker. Identical gray thumbnails make content indistinguishable. Full DB download on every launch wastes bandwidth and feels sluggish. Main-thread queries will cause jank as the library grows.

**Expected impact:** Eliminates the top 4 reasons a new user would uninstall after first session.

---

#### EPIC: Content Presentation Polish

**What:** Fix important UX issues that degrade first impression:
1. Add external link indicator on content rows ("在B站观看" or external-link icon)
2. Populate seed data with real thumbnail URLs
3. Show content count on category rows
4. Add loading state for first launch

**Why:** First impression drives retention. Right now, every row looks like a gray box, users don't know tapping opens a browser, and empty categories look the same as full ones. These are quick wins with outsized impact on perceived quality.

**Expected impact:** Users immediately understand the app's value proposition and navigation model. Time-to-value drops from "confused" to "oh, curated tutorials, let me browse."

---

### P1 — High-value differentiators (what makes BMC worth installing)

#### EPIC: Personal Learning State

**What:** Implement favorites (bookmarks) and watch history:
- Local `favorites` table — tap a heart icon to save tutorials for later
- Local `watch_history` table — automatically track when a user opens a tutorial link
- "My Library" tab showing favorited and recently watched content
- Visual indicator (checkmark or opacity change) on already-watched items

**Why:** This is the single highest-leverage feature for retention. Without personal state, BMC is a read-only directory with zero switching cost. With favorites and history, BMC becomes "my badminton study notebook" — something a user has invested in and would lose by switching away. Every tutorial app user expects this (Badminton Famly has it, YouTube has it, even 爱羽客 has it). Its absence is conspicuous.

**Expected impact:** Transforms BMC from a one-time discovery tool into a repeated-use study companion. Enables future features (learning paths, progress tracking, recommendations) that depend on knowing what the user has engaged with.

---

#### EPIC: Enhanced Search & Discovery

**What:** Improve search to match user mental models:
1. Include category names in search index (searching "步法" surfaces the footwork category)
2. Show category context in search results (each result shows its parent category)
3. Add search debounce on iOS (300ms)
4. Add "Editor's Pick" or featured content on the home screen

**Why:** Search is how intermediate users navigate — they know what technique they want to learn. Today's search misses categories entirely and returns results without context. The editor's pick feature gives first-time users immediate value without requiring them to know what to search for.

**Expected impact:** Reduces navigation friction for returning users. Editor's picks create a "front page" reason to open the app and see what's new.

---

#### EPIC: Content Detail Preview

**What:** Before opening the external link, show a half-sheet / bottom sheet with:
- Full title and summary
- Platform and author
- Difficulty level (once added)
- Editor's note explaining why this tutorial was selected
- "Watch on [Platform]" button that makes the external navigation explicit

**Why:** Two problems solved at once: (1) users are surprised when tapping opens a browser, and (2) the summary field is truncated in the list view so users can't evaluate content before committing. The preview sheet is a decision point — it sets expectations and adds editorial value. Badminton Famly has detailed lesson pages before video playback; this is BMC's equivalent.

**Expected impact:** Reduces "accidental" browser opens and bounce-backs. Surfaces the editorial curation (the "why this video" note) which is BMC's core differentiator.

---

### P2 — Engagement & retention features (keep users coming back)

#### EPIC: Difficulty Levels & Filtering

**What:** Add a `difficulty` enum (beginner / intermediate / advanced) to the `contents` table:
- Display as a colored badge on content rows
- Add filter chips on category screens (show All / Beginner / Intermediate / Advanced)
- Admin panel supports setting difficulty when adding content

**Why:** A beginner searching for 杀球 tutorials needs fundamentally different content than an advanced player. Without difficulty levels, users must evaluate each video themselves, which is exactly the friction BMC should eliminate. This also enables future learning paths ("Beginner's Track").

**Expected impact:** Users find relevant content faster. Reduces frustration from watching a tutorial that's too basic or too advanced. Creates natural content segmentation that aids future personalization.

---

#### EPIC: Curated Learning Paths

**What:** Introduce a `learning_paths` table — ordered sequences of content with editorial guidance:
- "Beginner's First 10 Lessons" — a guided tour through fundamentals
- "杀球 from Zero to Hero" — progressive skill building within one technique
- Each step has a short intro note from the editor
- Users can mark steps as completed (ties into Personal Learning State)

**Why:** This is BMC's strongest possible differentiator. No Chinese-language product offers structured badminton learning paths. Badminton Famly charges $5.99/mo for their "Training Plan Builder." BMC can offer a free, curated equivalent using existing community content. Learning paths transform BMC from "browse random tutorials" to "follow a curriculum" — a fundamentally more valuable proposition.

**Expected impact:** Dramatically increases session depth (users follow a path rather than bounce after one video). Creates a strong reason to return ("I'm on step 4 of 8"). Positions BMC as the "Duolingo of badminton" in users' mental model.

---

#### EPIC: Fresh Content Notifications

**What:** Lightweight "what's new" system:
- Badge on app icon when new content has been added since last open
- "New" tag on recently added content items (within last 7 days)
- Optional: a "New This Week" section on the home screen

**Why:** A static content library gives users no reason to return after initial browse. Even small signals of freshness ("3 new tutorials this week") create a pull to re-open the app. This also makes the editorial curation effort visible — users see that someone is actively maintaining and growing the library.

**Expected impact:** Increases weekly active opens. Low implementation cost, high signal value.

---

### P3 — Growth features (expand the user base)

#### EPIC: Onboarding & First-Run Experience

**What:** Single-screen onboarding that appears on first launch:
- Explains BMC's value proposition in one sentence: "精选羽毛球教学，从入门到高手"
- Shows the category structure visually
- Optionally asks user's skill level (beginner/intermediate/advanced) to personalize initial view
- "Start Learning" CTA that drops into the home screen

**Why:** Right now, a new user sees 6 emoji categories with no context. They don't know what the app does, why it exists, or what happens when they tap something. The design doc's positioning ("让好内容沉淀下来，长期可找") is compelling but invisible to the user. Even a single screen that communicates this will reduce first-session drop-off.

**Expected impact:** Reduces first-session confusion. If skill level is captured, enables personalized sorting/filtering from day one.

---

#### EPIC: Share & Invite Flow

**What:** Enable sharing from within BMC:
- Share a specific tutorial with a deep link (opens BMC if installed, falls back to source URL)
- Share a learning path as a card image (for WeChat Moments, Xiaohongshu)
- "Invite a practice partner" flow that shares the app with a personal message

**Why:** Badminton is inherently social — people play with partners and in groups. The most natural growth channel is one player sharing a useful tutorial with their practice partner. Without a share flow, users who discover a great tutorial in BMC will copy the source URL and share that instead, bypassing BMC entirely.

**Expected impact:** Organic word-of-mouth growth. Each share is a free acquisition touchpoint.

---

#### EPIC: Admin Panel Improvements

**What:** Quality-of-life improvements for the content curator:
1. Hierarchical category dropdown (indent subcategories under parents)
2. Server-side form validation with clear error messages
3. Bulk import from CSV/spreadsheet (for efficient content loading)
4. Thumbnail URL auto-fetch from source platform (Bilibili, YouTube have og:image)

**Why:** The speed at which the content library grows is directly tied to how efficient the admin workflow is. Right now, adding content is manual and error-prone (flat dropdown, silent validation failures). Every minute saved in admin workflow compounds into more content, which is BMC's core product.

**Expected impact:** Faster content addition rate. Fewer data entry errors. Enables scaling the library from 20 to 200+ items efficiently.

---

## 4. Recommended RM-2 Scope

Given the current state (20 seed tutorials, MVP quality, 4 critical bugs), the RM-2 goal should be: **make BMC good enough that a real badminton player would keep it installed after the first week.**

### Suggested RM-2 bundle

| Priority | EPIC | Effort estimate | Ship order |
|----------|------|----------------|------------|
| P0 | Platform Stability | S-M | 1st |
| P0 | Content Presentation Polish | S | 2nd |
| P1 | Personal Learning State (favorites + history) | M | 3rd |
| P1 | Enhanced Search & Discovery | S-M | 4th |
| P1 | Content Detail Preview | M | 5th |
| P2 | Difficulty Levels & Filtering | S | 6th |

The remaining P2 and P3 EPICs (Learning Paths, Fresh Content Notifications, Onboarding, Share Flow, Admin Improvements) are strong candidates for RM-3.

### Content investment (parallel to engineering)

Engineering alone will not make BMC compelling. The library needs to grow from 20 to at least 100+ tutorials across all categories, with real thumbnail URLs, difficulty tags, and genuine editor's notes. This is a content operations effort that should run in parallel with engineering work.

---

## 5. Key Strategic Takeaways

1. **BMC's moat is editorial judgment, not technology.** The app is simple; the value is that someone watched 50 杀球 videos and picked the best 3. Every product decision should amplify that editorial voice.

2. **Personal state is the retention unlock.** Without favorites and history, BMC is a brochure. With them, it becomes a tool. This is the single most important feature to build next.

3. **Learning paths are the long-term differentiator.** No Chinese-language product offers free, structured badminton curricula built from community content. This is the "Badminton Famly experience for free" pitch.

4. **Content freshness drives re-opens.** A content app that never changes is a dead app. Even adding 2-3 tutorials per week with a "New" badge creates a reason to come back.

5. **Fix the basics first.** None of the strategic features matter if Android has no back button and every thumbnail is a gray box. P0 items are prerequisites for everything else.

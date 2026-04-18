# Product Research — RM-4 EPIC Proposals

**Date:** 2026-04-15
**Scope:** Strategic analysis and competitive research for RM-4, building on RM-3 (learning paths, favorites, progress, enhanced search)
**Inputs:** RM-3 roadmap execution results, RM-1/RM-2 product research, competitive landscape analysis, content-as-code architecture review

---

## 1. Where BMC Stands After RM-3

### What shipped
- **Learning paths** — 3 curated paths (beginner 30-day, smash mastery, doubles positioning) with step-by-step progression
- **User state** — Favorites and path progress via local JSON, surviving content sync
- **Enhanced search** — Category badges, learning path search, source links
- **Content presentation** — Thumbnails from platform URLs, difficulty badges, editor's notes, content counts, platform indicators
- **Bug fixes** — Android search debounce, off-main-thread DB queries, URL encoding, deep linking for short URLs

### Current product profile
- **20 content items** across 6 top-level categories (26 subcategories), **4 creators**, **3 learning paths**
- Three surfaces: iOS, Android, web
- Content-as-code architecture: JSON files → validated → compiled to SQLite → synced via ETag
- No watch history, no social features, no notifications, no user accounts, no monetization

### The honest assessment
BMC has evolved from "link directory" to "structured learning companion." The learning paths are a genuine differentiator — no Chinese-language competitor offers free, curated badminton curricula from community content. But the library is still small (20 items), the user base is zero (personal project stage), and there is no mechanism for content discovery beyond opening the app and browsing.

**The RM-4 question is: what turns BMC from a well-built tool into something that grows?**

---

## 2. What's Working — BMC's Strongest Differentiators

### 2.1 Editorial curation over algorithmic recommendation
BMC's core value proposition is human judgment. Someone watched dozens of videos and selected the best ones per technique. This is the Wirecutter model applied to badminton tutorials. In a landscape where Bilibili's algorithm optimizes for watch time (not learning efficiency) and Douyin's format prevents systematic study, BMC's editorial voice is genuinely valuable.

### 2.2 Learning paths as content-as-code
The file-based architecture means learning paths are versioned, validated, and reproducible. This is a technical moat that enables rapid iteration on content structure without touching application code. No competitor has this — Badminton Famly's training plans are locked in their CMS; 爱羽客's content is tangled with ads and navigation.

### 2.3 Cross-platform source aggregation
BMC is the only product that aggregates content across Bilibili, YouTube, Xiaohongshu, and Douyin into a unified taxonomy. Each platform has excellent content trapped in its own silo. BMC breaks those silos.

### 2.4 Clean, focused design
The Nike-inspired monochromatic design makes BMC feel premium compared to 爱羽客 (ad-heavy, cluttered) and 中羽在线 (BBS-era UI). The design communicates editorial seriousness.

---

## 3. What's Missing — What Would Make a Player Recommend BMC

### 3.1 Content scale
20 items across 26 subcategories means many subcategories are empty. Categories like `attack/slice-drop`, `defense/net-block`, `defense/drive-block`, `net-play/pounce`, `fitness/flexibility`, `doubles/serve-receive`, and `doubles/rotation` have zero content. A user browsing these categories sees nothing — which destroys credibility. **Getting to 100+ items is the single highest-leverage investment for RM-4.**

### 3.2 Watch history and engagement tracking
Users have no way to see what they've already watched. Favorites exist, but passive tracking ("you watched this 3 days ago") is the feature that makes an app feel like it knows you. Every learning platform (Duolingo, Coursera, YouTube) tracks engagement passively.

### 3.3 Content freshness signals
There is no "what's new" mechanism. A user who checked BMC last week has no reason to check again unless they remember to. The app looks identical on every visit. Fresh content signals (badges, "New This Week" section, push notifications) create a pull to re-engage.

### 3.4 Social proof and sharing
Badminton is inherently social — players have partners, clubs, WeChat groups. BMC has no sharing mechanism. A player who discovers a great tutorial or learning path has no in-app way to share it. They would copy the Bilibili URL directly, bypassing BMC entirely.

### 3.5 Offline access / in-context viewing
Every content interaction requires leaving BMC to open a browser. This is the weakest point of the link-aggregator model. Users lose context, get distracted by Bilibili's recommendations, and may not return to BMC. While embedding video is legally and technically complex, there are intermediate solutions worth exploring.

---

## 4. Competitive Landscape — Beyond Badminton

### 4.1 Sports skill-learning apps

| Product | Sport | Model | What they do well | What BMC can learn |
|---------|-------|-------|-------------------|-------------------|
| **Badminton Famly+** | Badminton | Subscription ($5.99/mo) | Structured curriculum, Training Plan Builder, 2200+ videos, offline viewing | The "training plan" concept is validated — users will pay for structured learning. BMC offers this free with community content. |
| **TopCourt** | Tennis | Subscription ($14.99/mo) | Celebrity instructors (Serena Williams, etc.), cinematic production, drills with on-court filming guidance | Premium positioning through instructor branding. BMC's creators (杨晨大神, etc.) have strong recognition in the Chinese badminton community — their names carry weight. |
| **HomeCourt** | Basketball | Free + premium | AI-powered shot tracking via phone camera, real-time form feedback | Sensor/AI features require massive engineering investment. Not relevant for BMC's stage. But the idea of "practice tracking" (manual logging) is lightweight and achievable. |
| **STEEZY** | Dance | Subscription ($9.99/mo) | Step-by-step video breakdowns with speed control, mirror mode, loop sections, class-style progression | Video playback controls (speed, loop, mirror) are killer features for technique learning. BMC can't control playback on external platforms, but could suggest timestamp-based viewing ("watch 2:30-3:15 for the key technique"). |
| **Nike Training Club** | Fitness | Free | Workout programs with day-by-day scheduling, streak tracking, difficulty progression, social sharing of completions | The "program" model maps directly to BMC's learning paths. NTC's streak tracking and social sharing of completions are lightweight features with high engagement impact. |
| **Yousician** | Music | Freemium | Real-time feedback via microphone, structured lessons, daily practice goals, streak system | The daily practice goal concept is interesting — "practice 15 minutes of badminton drills today" combined with learning path steps could create a habit loop. |

### 4.2 General skill-learning platforms

| Product | What BMC can learn |
|---------|-------------------|
| **Duolingo** | Streak system creates daily habit. Bite-sized lessons (5 min) reduce friction. Social leaderboards create friendly competition. Notifications are aggressive but effective. The "daily goal" concept is the single most copied engagement mechanic in learning apps. |
| **Skillshare** | Class-based organization with "related classes" discovery. User reviews and project galleries create social proof. The "class" maps to BMC's learning path; "projects" could map to "practice session logs." |
| **MasterClass** | Premium instructor branding. Each class has a trailer, chapter list, and supplementary materials (workbooks, PDFs). BMC's learning paths could include supplementary notes, drill descriptions, or practice checklists alongside the video links. |
| **Notion / Obsidian** | Knowledge management as a personal tool. Users build their own structure. BMC could eventually allow users to create personal playlists or custom learning paths from the library — turning the app into "my badminton knowledge base." |

### 4.3 Key competitive insights

1. **Structured learning is a proven paid feature.** Badminton Famly, TopCourt, STEEZY, and Yousician all charge for curriculum structure. BMC offers this free — a strong acquisition angle.

2. **Streaks and daily goals are the most effective retention mechanic** across Duolingo, Nike Training Club, and Yousician. They work because they create a psychological cost of breaking the chain.

3. **Social sharing of progress is low-effort, high-impact.** Nike Training Club's "completed workout" share cards, Duolingo's streak shares, and Strava's activity posts all drive organic acquisition. The content is the user's achievement, not the app's marketing.

4. **No competitor in the Chinese badminton space has a mobile-first, clean, curated experience.** 爱羽客 is cluttered with ads and news. 中羽在线 is a BBS. The platforms (Bilibili, Douyin) are general-purpose. BMC's design quality is a real differentiator in this specific niche.

5. **Supplementary content (text guides, drill descriptions, practice checklists) adds value that video alone cannot.** STEEZY has written step breakdowns alongside video; MasterClass has workbooks. BMC's editor's notes are a start, but richer supplementary content per learning path step would deepen the experience.

---

## 5. Content Strategy — Growing from 20 to 100+

### 5.1 Current coverage gaps

| Category | Subcategories | Content items | Gap assessment |
|----------|--------------|---------------|----------------|
| **Basics** (基本功) | grip, serve, clear, footwork | 8 | Decent foundation. Missing: stance/ready position, basic rally patterns. |
| **Attack** (进攻) | smash, drop, slice-drop | 3 | Thin. Missing: push attack, flat drive, deceptive shots. `slice-drop` is empty. |
| **Defense** (防守) | smash-return, net-block, drive-block | 2 | Very thin. `net-block` and `drive-block` are empty. Missing: defensive footwork, lifting. |
| **Net Play** (网前) | spin, push, hook, pounce | 3 | Thin. `pounce` is empty. Missing: net kill, tumble net shot. |
| **Doubles** (双打) | positioning, rotation, serve-receive | 2 | Very thin. `rotation` and `serve-receive` are empty. Missing: doubles attack patterns, mixed doubles tactics. |
| **Fitness** (体能) | speed-agility, strength, flexibility | 2 | Minimal. `flexibility` is empty. Missing: injury prevention, warm-up routines, endurance. |

### 5.2 Highest-value content to add next

**Priority 1 — Fill empty subcategories (credibility)**
Every empty subcategory that a user navigates to is a broken promise. The category structure implies content exists; finding nothing erodes trust. Fill all 7 empty subcategories with at least 1-2 items each (14 items minimum).

**Priority 2 — Deepen the most-used categories (retention)**
Basics and attack are where beginners spend the most time. Growing these from 3-8 items to 10-15 each creates genuine browse depth.

**Priority 3 — Add high-demand content types not yet represented**
- **Match analysis / game IQ** — "Watch this rally and understand the tactical thinking." Very popular on Bilibili (赛事解析). Not represented in BMC's taxonomy at all.
- **Common mistakes / troubleshooting** — "Why your smash keeps going into the net." Extremely searchable, high-intent content. Only `common-grip-mistakes` exists.
- **Equipment / gear** — Racket selection, string tension, shoe recommendations. Xiaohongshu is full of this content. Useful for beginners who do not know what gear to buy.
- **Warm-up and injury prevention** — Highly practical. Players search for this after they get hurt. Preventive content has long-term SEO and shareability value.

### 5.3 New top-level category candidates

| Proposed category | Rationale |
|-------------------|-----------|
| **Tactics** (战术) | Match analysis, rally patterns, game IQ. Distinct from technique categories which focus on individual shots. Popular on Bilibili. |
| **Gear** (装备) | Racket, string, shoe, grip recommendations. Xiaohongshu content. High search intent. |
| **Recovery** (恢复) | Warm-up, cool-down, injury prevention, stretching. Could absorb `fitness/flexibility` and expand. |

---

## 6. Monetization Potential

Not urgent, but worth establishing a framework for future decisions.

### 6.1 Viable models for BMC's stage

| Model | Feasibility | Notes |
|-------|-------------|-------|
| **Free forever (passion project)** | High | Current model. Sustainable if BMC remains a personal/hobby project. No server costs beyond static file hosting. |
| **Donations / tip jar** | Medium | WeChat/Alipay tip jar with a "support the curator" message. Low friction, low revenue, but validates willingness to pay. |
| **Premium learning paths** | Medium | Free library + 2-3 free paths; additional advanced paths behind a one-time purchase or low subscription. The "Badminton Famly for free" positioning makes this tricky — charging contradicts the core pitch. |
| **Affiliate / gear referral** | Medium | Link to JD/Taobao for recommended gear with affiliate codes. Natural fit if a Gear category is added. Does not compromise the content curation mission. |
| **Sponsorship from badminton brands** | Low (for now) | Yonex, Li-Ning, Victor might sponsor a popular badminton learning platform. Requires significant user base first. |

### 6.2 Recommendation
Do not monetize in RM-4. Focus on growing the library and user base. If a revenue experiment is needed, **affiliate gear links** are the most natural fit — they add value (gear recommendations) rather than extracting it (paywalls).

---

## 7. Social / Community — Is There a Play?

### 7.1 The opportunity
Badminton is social by nature — you need at least one partner to play. Players organize through WeChat groups, club networks, and local courts. There is a natural community around skill improvement.

### 7.2 What's worth exploring (lightweight)

| Feature | Effort | Impact | RM-4? |
|---------|--------|--------|-------|
| **Share a learning path** (deep link or card image for WeChat) | Medium | High — organic acquisition | Yes |
| **Share progress** ("I completed 8/12 steps in the beginner plan") | Small | Medium — social proof + FOMO | Yes |
| **Practice log** (manual: "I practiced smash for 30 min today") | Medium | Medium — personal accountability | Maybe |
| **Leaderboard** (most paths completed, longest streak) | Medium | Medium — competitive motivation | No (needs user base) |
| **User-submitted content** (suggest a tutorial URL) | Medium | Medium — scales curation | No (needs moderation) |
| **Club/group features** | Large | High — but premature | No |
| **Comments / ratings on content** | Medium | Medium — social proof | No (needs user base) |

### 7.3 Recommendation
Ship **share flow** (learning paths and progress) in RM-4. This is the lowest-effort social feature with the highest acquisition impact. Defer community features (comments, leaderboards, clubs) until there is a meaningful user base to serve.

---

## 8. Proposed EPICs for RM-4

Based on the research above, RM-4 should focus on two themes: **content scale** (the library must grow to be credible) and **engagement loops** (give users reasons to come back). Social sharing bridges both — it drives acquisition and creates accountability.

### EPIC 1: HISTORY — Watch History & Engagement Tracking

**What:** Automatically track when a user opens a content link. Show a "Recently Watched" section on the home screen. Mark watched items with a visual indicator in lists.

**Why:** This is the most-requested missing feature from the RM-1 research that was intentionally deferred. With 100+ items and learning paths, users need passive tracking to remember where they left off. Every learning platform has this. Its absence makes BMC feel like a static brochure rather than a personal tool.

**Scope:**
| # | Task | Effort |
|---|------|--------|
| HISTORY-1 | Extend UserState JSON to include `watch_history` array (content slug + timestamp) | Small |
| HISTORY-2 | iOS — record watch event when opening external link; "Recently Watched" section on home | Medium |
| HISTORY-3 | Android — same as HISTORY-2 | Medium |
| HISTORY-4 | Visual indicator (checkmark or reduced opacity) on already-watched items in all list views | Small |
| HISTORY-5 | "Continue Learning" — on path detail screen, highlight the next unwatched step | Small |
| HISTORY-6 | Web client — watch history section on home page | Small |

**Depends on:** STATE (RM-3, already shipped).
**Unlocks:** Streak tracking (STREAK), personalized recommendations (future).

---

### EPIC 2: FRESH — Content Freshness & Re-engagement

**What:** Add "New" badges on recently added content, a "New This Week" section on the home screen, and an app badge count for new items since last open.

**Why:** A content app that looks identical every time you open it is a dead app. Even adding 3-5 items per week with visible freshness signals creates a pull to re-engage. This was identified as P2 in RM-3 research but was deferred. With content growth planned for RM-4, freshness signals become critical — users need to see that the library is actively curated.

**Scope:**
| # | Task | Effort |
|---|------|--------|
| FRESH-1 | Add `created_at` field to content schema (or derive from git commit date during build); write to compiled DB | Small |
| FRESH-2 | "New" badge on content items added within last 14 days (iOS, Android, web) | Small |
| FRESH-3 | "New This Week" section on home screen, showing items added in the last 7 days | Medium |
| FRESH-4 | Track last-open timestamp in UserState; show app badge count of new items since last open (iOS, Android) | Medium |
| FRESH-5 | Compiler: generate `content_changelog` table from git history (added/updated items per week) | Medium |

**Depends on:** Content growth (GROW-4). Freshness signals without fresh content are meaningless.
**Unlocks:** Push notifications (future), email digest (future).

---

### EPIC 3: GROW — Content Library Scale-Up (20 → 120+)

**What:** Batch ingestion campaign to fill all empty subcategories and deepen existing ones. Add 2-3 new top-level categories (Tactics, Gear, Recovery). Grow creators from 4 to 10+. Write genuine editor's notes for every item.

**Why:** 20 items across 26 subcategories (7 of which are empty) does not demonstrate editorial authority. The taxonomy promises breadth that the content does not deliver. Users who browse into an empty subcategory lose trust. Growing to 120+ items with full coverage makes BMC credible as "the curated library" rather than "a small side project."

**Scope:**
| # | Task | Effort |
|---|------|--------|
| GROW-1 | Define per-category content targets (minimum 3 items per subcategory, 8+ for popular ones) | Small (planning) |
| GROW-2 | Fill 7 empty subcategories: `slice-drop`, `net-block`, `drive-block`, `pounce`, `rotation`, `serve-receive`, `flexibility` (14+ items) | Medium (content) |
| GROW-3 | Deepen basics (grip, serve, clear, footwork) and attack (smash, drop) to 8+ items each | Medium (content) |
| GROW-4 | Add new top-level category: Tactics (战术) with subcategories for match analysis, rally patterns, game IQ | Medium (content + schema) |
| GROW-5 | Add new top-level category: Gear (装备) with subcategories for rackets, strings, shoes | Small (content + schema) |
| GROW-6 | Add 6+ new creators with full person profiles | Small (content) |
| GROW-7 | Write genuine editor's notes for all items — "why this video" not "what this video is about" | Medium (content) |
| GROW-8 | Add 1-2 new learning paths using newly ingested content (e.g., "Defensive Fundamentals", "Gear Guide for Beginners") | Medium (content) |

**Depends on:** Ingestion pipeline (shipped in RM-3).
**Unlocks:** FRESH (freshness signals need fresh content), credibility for sharing/growth.

---

### EPIC 4: SHARE — Social Sharing & Organic Growth

**What:** Enable sharing content items and learning paths via native share sheet. Generate shareable card images for WeChat Moments. Include progress in shared cards ("I completed 8/12 steps").

**Why:** Badminton players organize in WeChat groups and share tips constantly. Without a share flow, BMC's best content gets shared as raw Bilibili URLs, bypassing BMC entirely. Sharing is the primary organic acquisition channel for a niche app. Nike Training Club, Strava, and Duolingo all drive growth through share-your-achievement flows.

**Scope:**
| # | Task | Effort |
|---|------|--------|
| SHARE-1 | iOS — share button on content detail: native share sheet with BMC deep link + fallback source URL | Medium |
| SHARE-2 | Android — same as SHARE-1 | Medium |
| SHARE-3 | iOS — share learning path as card image (title, step count, progress if applicable, QR code to BMC) | Medium |
| SHARE-4 | Android — same as SHARE-3 | Medium |
| SHARE-5 | Web — share URL with Open Graph meta tags so links preview nicely in WeChat/iMessage | Small |
| SHARE-6 | Deep link handler: receiving a shared BMC link opens the correct content or path (extends existing deep linking) | Medium |

**Depends on:** PATH + STATE (RM-3, already shipped). HISTORY (for "continue where I left off" in shared paths).
**Unlocks:** Organic acquisition, social proof, word-of-mouth growth.

---

### EPIC 5: STREAK — Practice Streaks & Daily Goals

**What:** Introduce a lightweight streak system: users set a daily goal (e.g., "watch 1 tutorial" or "complete 1 learning path step"), and the app tracks consecutive days of engagement. Display streak count on home screen. Optional: weekly summary.

**Why:** Streaks are the single most effective retention mechanic in learning apps (Duolingo, Yousician, Nike Training Club). They work because breaking a streak has a psychological cost. For BMC, a streak tied to learning path progress creates a daily habit: "I need to watch today's tutorial to keep my streak." This transforms BMC from an occasional reference tool into a daily practice companion.

**Scope:**
| # | Task | Effort |
|---|------|--------|
| STREAK-1 | Define streak rules: what counts as a "practice day" (opened a tutorial? completed a path step? either?) | Small (design decision) |
| STREAK-2 | Extend UserState JSON with streak data (current streak, longest streak, last active date) | Small |
| STREAK-3 | iOS — streak display on home screen (flame icon + day count), streak broken/maintained notification | Medium |
| STREAK-4 | Android — same as STREAK-3 | Medium |
| STREAK-5 | "Weekly Summary" — end-of-week review showing tutorials watched, path steps completed, streak status | Medium |
| STREAK-6 | Shareable streak card ("I've practiced badminton for 14 days straight with BMC") | Small |

**Depends on:** HISTORY (streak requires knowing when user last engaged with content).
**Unlocks:** Daily active usage, retention, shareable achievement moments.

---

## 9. Recommended RM-4 Scope & Ship Order

### Guiding principle
RM-3 built the structural differentiator (learning paths). RM-4 must fill it with content and create reasons to come back. **Content scale + engagement loops + social sharing.**

### Suggested RM-4 bundle

| Priority | EPIC | Theme | Effort | Ship order |
|----------|------|-------|--------|------------|
| **P0** | GROW (content scale-up) | Content | Large (content ops, not engineering) | Continuous, start first |
| **P0** | HISTORY (watch history) | Engagement | Medium | 1st engineering sprint |
| **P1** | FRESH (freshness signals) | Engagement | Medium | 2nd sprint (needs GROW in progress) |
| **P1** | SHARE (social sharing) | Growth | Medium-Large | 3rd sprint |
| **P2** | STREAK (practice streaks) | Retention | Medium | 4th sprint (needs HISTORY) |

### Dependency graph

```
GROW ──────────────────────────────> (content ops, runs continuously in parallel)
                                     |
HISTORY ──> FRESH (freshness needs created_at; history enables "continue learning")
        ──> STREAK (streaks need engagement tracking from history)
        ──> SHARE (share progress requires knowing what user watched)

SHARE ──> depends on HISTORY (for progress sharing)
      ──> enhanced by GROW (more content = more shareable moments)

STREAK ──> depends on HISTORY
       ──> feeds into SHARE (shareable streak cards)
```

### What to defer to RM-5+

| Feature | Why defer |
|---------|-----------|
| **Cross-device sync / user accounts** | Engineering-heavy. No user base to justify cloud infrastructure. Local-first is fine for now. |
| **Content tags / themes** | Useful for discovery at 200+ items. At 120 items, the category tree is sufficient. |
| **FTS5 full-text search** | Current LIKE-based search is adequate for the library size. Revisit when search performance degrades. |
| **Push notifications** | Requires server infrastructure and notification service. Defer until there are users to notify. |
| **User-submitted content / community curation** | Requires moderation tooling and trust framework. Premature without a user base. |
| **In-app video playback** | Legally complex (content licensing), technically heavy (player integration per platform). The link-out model works. |
| **Monetization** | No users yet. Focus on growth. If experimenting, affiliate gear links are the lowest-friction option. |
| **Onboarding flow** | Deferred from RM-3. More impactful once the library is larger and there are learning paths to recommend based on skill level. Could ship as a quick win alongside GROW. |

---

## 10. Key Strategic Takeaways

1. **Content scale is the bottleneck, not features.** BMC has learning paths, favorites, progress tracking, search, and a beautiful design. What it lacks is enough content to justify its existence. 7 empty subcategories out of 26 means 27% of the app is broken promises. Growing from 20 to 120+ items is the single highest-leverage investment.

2. **Watch history completes the personal learning loop.** Favorites (active intent) + watch history (passive tracking) + path progress (structured progression) + streaks (habit formation) create a complete engagement stack. History is the missing piece that enables the rest.

3. **Sharing is the growth engine for a niche app.** BMC will never have a marketing budget. Growth comes from one player sharing a learning path with their practice partner in a WeChat group. The share flow must be frictionless and produce artifacts (card images, deep links) that look good in Chinese social platforms.

4. **Streaks are the retention mechanic, not a gimmick.** Duolingo's entire business model rests on streak psychology. For BMC, tying streaks to learning path progress ("Day 15 of your 30-day beginner plan") creates a daily habit that transforms occasional browsing into systematic practice.

5. **BMC's positioning crystallizes in RM-4.** After RM-4, the pitch becomes: "Free, curated badminton curriculum from the best Chinese-language creators. Follow a 30-day plan, track your progress, share with your practice partner." That is a product worth recommending. The RM-3 pitch ("curated badminton tutorials with learning paths") was close but lacked the engagement and social layers to stick.

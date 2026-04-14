# UX User Testing Report — Badminton Master Class

**Date:** 2026-04-14
**Scope:** iOS app, Android app, Admin panel
**Persona:** Badminton enthusiast, first-time user, intermediate player looking to improve technique

---

## Critical Issues (Blocks Usability)

### 1. Android: No back navigation in CategoryScreen
The Android `CategoryScreen` has a `TopAppBar` but no back/up button. Once a user taps into a category, the only way back is the system back gesture or hardware back button. Many users (especially those coming from iOS or using gesture navigation on newer Android) may not discover this. The iOS version handles this automatically via `NavigationStack`.

**Files:** `android/.../ui/CategoryScreen.kt` (line 64-73 — TopAppBar has no `navigationIcon`)

### 2. Android: Thumbnails never load actual images
The Android `ContentThumbnail()` composable is a hardcoded placeholder (gray box with play icon). It completely ignores the `thumbnailUrl` from the data model. The comment on line 220 says "remote image loading requires Coil (to be added later)." This means every content row looks identical on Android, making it much harder for users to visually scan and identify content.

**Files:** `android/.../ui/CategoryScreen.kt` (lines 222-239)

### 3. Sync always downloads the full database on every launch
Both iOS (`DataSync.syncIfNeeded()` called in `BMCApp.onAppear`) and Android (`DataSync.syncIfNeeded(context)` in `BMCApp LaunchedEffect`) download the entire `bmc.db` from OSS on every app launch. There is no conditional logic (e.g., ETag, If-Modified-Since, or local timestamp check). This means:
- Users on slow/metered connections waste bandwidth every launch
- The app may feel sluggish on startup, especially with a growing database
- The name `syncIfNeeded` is misleading — it always syncs unconditionally

**Files:** `ios/BadmintonMasterClass/DataSync.swift` (line 81), `android/.../data/DataSync.kt` (line 66)

### 4. iOS: Database queries run on the main thread
All `Database.shared.categories()`, `Database.shared.contents()`, and `Database.shared.searchContents()` calls happen synchronously on the main thread (called directly in SwiftUI `onAppear` and `onChange`). With a growing content library, this will cause UI jank and potential ANR-like freezes on iOS.

**Files:** `ios/BadmintonMasterClass/Database.swift`, `ios/BadmintonMasterClass/HomeView.swift` (lines 31-33, 39, 68)

---

## Important Issues (Degrades Experience)

### 5. No onboarding or first-launch explanation
When a user opens the app for the first time, they see a list of 6 categories with emoji icons and Chinese names. There is no explanation of what the app is, how it works, or what to expect when tapping a content item (that it opens a browser/Safari). A first-time user may be confused about the app's purpose.

### 6. No indication that content opens externally
When a user taps a content item, it opens in SFSafariViewController (iOS) or Chrome Custom Tabs (Android). There is no visual cue (e.g., an external link icon, a "Watch on Bilibili" label, or a tooltip) telling the user they will leave the app context. The platform badge helps but is subtle.

### 7. Search only queries content, not categories
The search bar searches `title`, `summary`, and `author_name` in the `contents` table. A user searching for "步法" (footwork) will find individual videos, but not the "步法" category itself. Users may expect to find categories by name.

**Files:** `ios/BadmintonMasterClass/Database.swift` (line 120), `android/.../data/Database.kt` (line 120)

### 8. Search results lack category context
When search returns results, each `ContentRow` shows title, summary, platform badge, and author — but not which category the content belongs to. A user searching "教学" would get many results with no way to understand the organizational context.

### 9. iOS: Sync status bar auto-dismisses too quickly
The "已同步" (synced) message disappears after 2 seconds and "同步失败" (sync failed) after 3 seconds. If the user is not looking at the bottom of the screen, they miss it entirely. The failed state is especially problematic — there is no retry button, no persistent indicator, and no way for the user to know their data might be stale.

**Files:** `ios/BadmintonMasterClass/SyncManager.swift` (lines 25-42)

### 10. Android: Pull-to-refresh calls `syncIfNeeded` which doesn't force sync
On Android, pull-to-refresh calls `DataSync.syncIfNeeded(context)`. Since this function name implies conditional sync but actually always downloads, the naming is misleading. More importantly, if the sync fails, the `isRefreshing` flag is set to `false` immediately after the coroutine completes, but the user gets no feedback about the failure other than the briefly-visible sync status bar.

### 11. Admin: Category dropdown in "Add Content" is flat, not hierarchical
When adding content in the admin panel, the category dropdown shows all categories (top-level and subcategories) in a flat list ordered by `sort_order`. This means "握拍" (a subcategory of "基本功") appears alongside "基本功" with no visual distinction. Admins could accidentally assign content to a top-level category instead of a subcategory.

**Files:** `admin/templates/contents.html` (lines 81-84)

### 12. Admin: No validation feedback on form errors
The admin forms use basic HTML `required` attributes but have no server-side validation feedback. If a user enters an invalid `sort_order` (e.g., a non-number), `strconv.Atoi` silently returns 0 instead of showing an error.

**Files:** `admin/handlers.go` (lines 79, 186)

### 13. Android: Category name in navigation route can break with special characters
The Android navigation passes `category.name` directly in the URL route: `"category/${category.id}/${category.name}"`. If a category name contains `/`, `?`, `#`, or other URL-special characters, navigation will break. Should encode the name or pass only the ID and look up the name.

**Files:** `android/.../ui/BMCApp.kt` (lines 43, 59)

---

## Nice-to-Haves (Would Improve but Not Critical)

### 14. No content count shown on categories
The home screen category list shows only icon + name. Users have no idea if a category has 1 item or 100. Showing a count (e.g., "杀球 (3)") would help users decide where to explore.

### 15. No loading state while data loads on first launch
On first launch, categories are loaded in `onAppear`/`LaunchedEffect`. If the bundled DB is empty or the initial sync is slow, the user sees the empty state ("暂无内容 — 下拉刷新获取最新数据") briefly before content appears. There is no loading spinner for the initial data load.

### 16. Seed data has no thumbnail URLs
All 20 seed content entries have empty `thumbnail_url` fields. This means even on iOS (which supports `AsyncImage`), every content row shows a gray placeholder. The admin form supports entering thumbnail URLs, but none of the seed data has them.

**Files:** `data/seed.sql` (line 60 — INSERT has no `thumbnail_url` column)

### 17. iOS: Search executes on every keystroke
The search in `HomeView` fires `Database.shared.searchContents(keyword:)` on every character change via `.onChange(of: searchText)`. There is no debounce. For fast typers, this creates many rapid SQLite queries.

**Files:** `ios/BadmintonMasterClass/HomeView.swift` (lines 31-33)

### 18. Admin: Export triggers OSS upload synchronously in the request handler
The `exportHandler` calls `tryUploadToOSS(dbPath)` synchronously before streaming the file to the browser. If the OSS upload is slow, the admin user waits for it to complete before the download starts. The comment says "background (best-effort)" but the code is blocking.

**Files:** `admin/handlers.go` (lines 500-521)

### 19. No dark mode-specific styling on iOS
The iOS app uses default SwiftUI styling, which does adapt to dark mode, but there is no custom dark mode color scheme. The Android version uses Material 3 dynamic colors which adapt well. The iOS version could benefit from matching intentionality.

### 20. Admin panel has no mobile responsiveness
The admin panel uses a simple `max-width: 800px` CSS layout with tables. On a phone, the content table with 8 columns will overflow horizontally. Admin work is typically done on desktop, but mobile admin access would be difficult.

---

## Strengths (What Works Well)

### Clean information architecture
The two-level category hierarchy (6 top-level categories, ~3-4 subcategories each) is intuitive for the domain. A badminton player can quickly navigate: 基本功 > 步法 > specific video. The taxonomy is well-thought-out.

### Consistent cross-platform experience
iOS and Android share the same data model, navigation pattern (Home > Category > Content), and visual structure. A user switching between platforms would feel at home.

### Platform badges are useful and well-designed
The colored platform badges (B站, 小红书, 抖音, YouTube) immediately tell users what app/site the content links to. The color coding matches each platform's brand identity.

### Smart content row design
The `ContentRow` layout (thumbnail + title + summary + platform/author) follows a familiar pattern from YouTube, Bilibili, and similar apps. Users instantly understand what they're looking at.

### Graceful degradation on sync failure
Both platforms handle sync failures gracefully — the app continues working with local/bundled data. Users are never blocked from browsing even without network.

### Robust admin CRUD with safety checks
The admin panel prevents deleting categories that have children or associated content (`categoryDeleteHandler` checks for both). This prevents accidental data loss.

### Schema migration system
The admin server has a proper migration system (`migrate.go`) with version tracking, which will make future schema changes safe and reliable.

### Pull-to-refresh for manual sync
Both platforms support pull-to-refresh to manually trigger a data update, which is the expected interaction pattern for content apps.

---

## Recommended Product Features for Next Roadmap

### P0 — Fix before launch
1. **Add back navigation on Android CategoryScreen** — Add `navigationIcon` with back arrow to `TopAppBar`
2. **Implement Android thumbnail loading** — Integrate Coil for `AsyncImage` on Android
3. **Add conditional sync** — Use HTTP ETag or If-Modified-Since headers to skip download when data hasn't changed
4. **Move iOS database queries off the main thread** — Use async/await or background actor

### P1 — High-value features
5. **Favorites/bookmarks** — Let users save tutorials they want to revisit. Store locally in a `favorites` table. This is the #1 feature users of tutorial browsers expect.
6. **Search improvements** — Include category names in search results; add debounce on iOS; show which category each result belongs to
7. **Content detail preview** — Before opening the external link, show a half-sheet with title, full summary, platform, author, and a "Watch" button. This sets expectations and gives the user a chance to decide.
8. **External link indicator** — Add a subtle external-link icon or "在B站观看" text on content rows

### P2 — Engagement features
9. **Watch history** — Track which tutorials the user has opened. Show a "Recently watched" section or a checkmark on viewed items.
10. **Difficulty levels** — Add a `difficulty` field (beginner/intermediate/advanced) to content. Show as a badge. Let users filter by level.
11. **Curated learning paths** — Create ordered playlists like "Beginner's 10-day program" that guide users through content in a recommended sequence.
12. **Content count badges** — Show item counts on category rows so users know where the most content lives.
13. **Offline indicator** — Show a persistent indicator when the app has never successfully synced (e.g., first launch without network).

### P3 — Polish
14. **Onboarding screen** — A single-screen explanation: "Curated badminton tutorials from top creators. Browse by technique, search, and watch."
15. **Seed thumbnail URLs** — Add real thumbnail images to seed data for a better first impression
16. **Admin UX improvements** — Hierarchical category dropdown, form validation, mobile-responsive layout
17. **Search result ranking** — Rank by relevance (title match > summary match > author match) instead of `sort_order`

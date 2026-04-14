# Design System — Badminton Master Class (羽球大师课)

Nike-inspired design adapted for a native mobile app (iOS + Android) that curates badminton instructional videos from Chinese social platforms.

## 1. Visual Theme & Atmosphere

Badminton Master Class is a training ground distilled into an app — a curated library of badminton technique videos that channels the explosive energy of sport into a focused learning experience. The design operates on the same principle as Nike's retail cathedral: radical monochromatic simplicity so that content — video thumbnails, instructor photos, platform badges — provides all vibrancy without competition from the UI itself.

The interface disappears into Ink Black (`#111111`) text and white surfaces, allowing technique thumbnails and colorful platform badges (Bilibili pink, Xiaohongshu red, Douyin black, WeChat green, YouTube red) to carry the visual energy. These platform colors are the app's equivalent of Nike's product photography — the only source of chromatic life in an otherwise greyscale world. When color does appear in the UI layer, it's purely functional: red for errors, green for sync success.

Typography uses platform-native system fonts — SF Pro on iOS, Roboto on Android — both excellent for CJK text rendering. The hierarchy mirrors Nike's tension between expressive display type (large, bold, tight line-height for screen titles) and clinical body type (regular weight, generous line-height for comfortable content browsing). Weight Medium (500/.semibold) dominates interactive elements, giving every label a quiet confidence.

**Key Characteristics:**
- Monochromatic UI (black/white/grey) that lets thumbnails and platform badges be the only color source
- Bold screen titles with tight line-height that anchor each view
- Thumbnails with subtle corner radius (6dp) — the app's visual heartbeat
- Capsule-shaped badges and pill-shaped buttons as primary interactive accents
- 8dp spacing grid with athletic discipline — every measurement snaps to the system
- Category-driven technique hierarchy with emoji icons as wayfinding
- Shadow-free, border-minimal elevation — surface differentiation through grey shifts only

## 2. Color Palette & Roles

### Primary

- **Ink Black** (`#111111`): The foundation — primary text, navigation bar text, button backgrounds. Deliberately not pure black (#000000), creating a fractionally softer reading experience against white
- **Canvas White** (`#FFFFFF`): Primary screen background, button text on dark, card surfaces

### Surface & Background

- **Snow** (`#FAFAFA`): Lightest surface, near-white subtle differentiation for grouped backgrounds
- **Light Gray** (`#F5F5F5`): Secondary background, search input fill, thumbnail placeholder, loading skeleton
- **Press Gray** (`#E5E5E5`): Pressed/highlighted state background, disabled button fill
- **Dark Surface** (`#28282A`): Primary background for dark mode or inverted sections
- **Deep Charcoal** (`#1F1F21`): Darkest non-black surface, dark mode canvas
- **Dark Press** (`#39393B`): Pressed state on dark backgrounds

### Neutrals & Text

- **Primary Text** (`#111111`): Main body text, headings, navigation titles
- **Secondary Text** (`#707072`): Descriptive copy, metadata, author names, timestamps, summaries
- **Disabled Text** (`#9E9EA0`): Inactive elements, unavailable options
- **Disabled Inverse** (`#4B4B4D`): Disabled text on dark backgrounds
- **Divider** (`#CACACB`): List separators, input borders, subtle divider lines
- **Border Active** (`#111111`): Focused input border, active selection indicator

### Platform Badge Colors

These are the app's "product colors" — the only non-greyscale hues permitted in the UI, used exclusively for source platform identification:

- **Bilibili Pink** (`#FB7299`): Bilibili content badge and tinted background at 15% opacity
- **Xiaohongshu Red** (`#FF2442`): Xiaohongshu content badge
- **Douyin Black** (`#161823`): Douyin content badge (near-black, uses lighter text)
- **WeChat Green** (`#07C160`): WeChat content badge
- **YouTube Red** (`#FF0000`): YouTube content badge

Badge rendering: platform color as text, same color at 15% opacity as background fill, capsule shape.

### Semantic

- **Error Red** (`#D30005`): Sync failures, validation errors
- **Success Green** (`#007D48`): Sync complete confirmation, availability indicators
- **Success Inverse** (`#1EAA52`): Success on dark backgrounds
- **Warning Yellow** (`#FEDF35`): Attention states (rarely used)

### Gradient System

No UI gradients. The design system is flat-color only. If gradients appear, they exist only within thumbnail imagery from source platforms.

## 3. Typography Rules

### Font Family

**iOS:** SF Pro — the platform default, with excellent CJK (Simplified Chinese) support via SF Pro SC / PingFang SC automatic fallback.

**Android:** Roboto — the Material Design default, with Noto Sans SC as the automatic CJK fallback.

No custom fonts. System fonts ensure optimal rendering, consistent CJK character display, and zero load time.

### Hierarchy

| Role | iOS (SwiftUI) | Android (Compose) | Weight | Usage |
|------|---------------|-------------------|--------|-------|
| Screen Title | .largeTitle (34pt) | headlineLarge (32sp) | Bold/.bold | NavigationStack inline title, screen headers |
| Section Title | .title2 (22pt) | titleLarge (22sp) | Bold/.bold | Section headers within a screen |
| Category Name | .body (17pt) | bodyLarge (16sp) | Regular/.regular | Category list item text |
| Content Title | .headline (17pt) | titleMedium (16sp) | Semibold/.semibold | Content row title — the primary scannable element |
| Summary | .subheadline (15pt) | bodyMedium (14sp) | Regular/.regular | Content row description, 2-line max, secondary color |
| Caption | .caption (12pt) | labelSmall (11sp) | Regular/.regular | Author name, duration, metadata |
| Badge | .caption2 (11pt) | labelSmall (11sp) | Medium/.medium | Platform badge text (Bilibili, Douyin, etc.) |
| Category Icon | .title2 (22pt) | titleLarge (22sp) | — | Emoji icons in category list |
| Sync Status | .caption2 (11pt) | labelSmall (11sp) | Regular/.regular | Bottom bar sync indicator text |

### Principles

The typography serves two masters: Chinese readability and athletic energy. Screen titles use bold weight with the default tight line-height to punch through each view — they feel like section headers in a training manual. Below the titles, body text relaxes into regular weight with generous line-height for comfortable vertical scanning through technique lists. Weight Medium/Semibold dominates interactive elements (content titles, badges, buttons), giving every tappable element a subtle visual assertiveness without the heaviness of bold.

CJK-specific: Chinese characters are visually denser than Latin text. The system fonts handle this natively, but designers should note that Chinese text at the same point size reads "heavier" than English — resist the urge to compensate by increasing size.

## 4. Component Stylings

### Content Row (the core component)

The content row is the most-seen element in the app — it appears in every category screen and search result.

**Layout:** Horizontal — thumbnail left, text stack right
- Row padding: 16dp horizontal, 12dp vertical
- Internal gap: 12dp between thumbnail and text

**Thumbnail:**
- Size: 60dp wide x 45dp tall (4:3 aspect ratio)
- Corner radius: 6dp
- Content scale: fill/crop
- Placeholder: Light Gray (`#F5F5F5`) background with a centered play icon (20dp, secondary color)
- Loading: async with fade-in transition

**Text Stack (vertical):**
- Title: Content Title style (headline/titleMedium), Ink Black, 2-line max, ellipsis truncation
- Summary: Summary style (subheadline/bodyMedium), Secondary Text color, 2-line max
- Metadata row (horizontal, 4dp gap): platform badge + author name in Caption style

**Platform Badge:**
- Shape: capsule (fully rounded, RoundedCornerShape(50%) / .capsule)
- Padding: 8dp horizontal, 2dp vertical
- Text: Badge style, platform color as text color
- Background: platform color at 15% opacity
- No border, no shadow

### Category Row

**Layout:** Horizontal — emoji icon left, category name right
- Row padding: 16dp horizontal, 14dp vertical
- Internal gap: 12dp between icon and name
- Separator: 1px divider in Divider color (`#CACACB`) between rows

**Icon:** Category Icon style (title2/titleLarge)
**Name:** Category Name style (body/bodyLarge), Ink Black

### Search Bar

- Background: Light Gray (`#F5F5F5`)
- Corner radius: 12dp
- Padding: 16dp horizontal, 8dp vertical (outer); content padding per platform default
- Leading icon: magnifying glass, Secondary Text color
- Trailing icon: clear button (X), visible only when text is present
- Placeholder text: "搜索教程", Secondary Text color
- Input text: Ink Black, body/bodyLarge style
- Debounce: 300ms before triggering search

### Buttons

**Primary (pill)**
- Background: Ink Black (`#111111`)
- Text: Canvas White, body/bodyLarge weight Medium
- Corner radius: fully rounded pill (30dp)
- Padding: 12dp vertical, 24dp horizontal
- Pressed: background shifts to `#707072`

**Secondary (outlined pill)**
- Background: transparent
- Text: Ink Black (`#111111`)
- Border: 1.5dp solid Divider (`#CACACB`)
- Corner radius: 30dp
- Pressed: border darkens to `#707072`, background to Press Gray

**Disabled**
- Background: Press Gray (`#E5E5E5`)
- Text: Disabled Text (`#9E9EA0`)

### Navigation

**iOS (NavigationStack):**
- Large title mode for Home screen ("羽球大师课")
- Inline title mode for category detail screens
- System back button with category name
- Background: system default (translucent white, blurs on scroll)

**Android (TopAppBar):**
- Background: primaryContainer (Material 3 dynamic color, or Light Gray in custom theme)
- Title text: onPrimaryContainer color
- Navigation icon: back arrow on detail screens
- No elevation shadow — flat

### Sync Status Bar

- Position: bottom of screen, above safe area
- Background: Light Gray (`#F5F5F5`)
- Layout: horizontal — progress indicator + status text
- Progress: 14dp circular indicator, 2dp stroke
- Text: Sync Status style (caption2/labelSmall), Secondary Text color
- States: "正在同步..." (syncing), "已同步" (synced), "同步失败" (failed, Error Red text)
- Animation: fade in/out with 200ms duration

### Empty State

- Layout: centered vertically and horizontally
- Icon: 48dp, Secondary Text color
- Title: bodyLarge/body style, Secondary Text color
- Description: bodyMedium/subheadline, Secondary Text at 70% opacity
- Generous vertical spacing (16dp between elements)

### Pull-to-Refresh

- Standard platform refresh control
- Triggers sync check with ETag-based conditional download
- No custom styling — use system default

## 5. Layout Principles

### Spacing System

Base unit: 4dp (primary grid is 8dp multiples)

| Token | Value | Use |
|-------|-------|-----|
| space-1 | 2dp | Badge internal vertical padding |
| space-2 | 4dp | Tight inline gaps, metadata row spacing |
| space-3 | 8dp | Badge horizontal padding, search bar vertical padding, icon gaps |
| space-4 | 12dp | Content row internal gap, category row gap, section internal padding |
| space-5 | 14dp | Category row vertical padding |
| space-6 | 16dp | Standard screen horizontal padding, content row horizontal padding |
| space-7 | 24dp | Section breaks within a screen |
| space-8 | 32dp | Major section separation |

### Screen Structure

All screens follow a single-column vertical scroll (LazyColumn on Android, List on iOS):

- **Home:** Search bar (pinned/scrollable) → category list → sync status bar
- **Category:** Back navigation → subcategory section (if any) → content list
- **Search Results:** Search bar (active) → filtered content list → empty state

No grids. No multi-column layouts. The app is phone-first, single-column, vertically scrolling.

### Whitespace Philosophy

The app borrows Nike's compressed density for content lists — tight vertical gaps (12-14dp between rows) create a sense of abundant technique coverage. But within each row, spacing is generous (12dp between thumbnail and text) for comfortable scanning. The overall effect: a well-organized training library that feels comprehensive without being cluttered.

### Corner Radius Scale

| Value | Context |
|-------|---------|
| 0dp | No usage — the app avoids sharp edges on visible elements |
| 6dp | Thumbnails — subtle rounding, still feels editorial |
| 12dp | Search bar — medium rounding, friendly input feel |
| 30dp | Buttons — full pill shape |
| 50% | Capsule badges, circular icon buttons |

## 6. Depth & Elevation

| Level | Treatment | Use |
|-------|-----------|-----|
| Flat | No shadow, no border | Default for all surfaces — cards, rows, backgrounds |
| Divider | 1px line in Divider color (`#CACACB`) | Between category rows, between sections |
| System | Platform navigation bar blur | iOS NavigationStack translucent bar on scroll |

The elevation philosophy is radically flat, inherited from Nike. No card shadows, no floating action buttons, no raised surfaces. Depth is communicated exclusively through color shifts — Light Gray sections recede, white sections advance, divider lines separate list items. This flatness reinforces the athletic, no-nonsense personality: no visual frills, just direct access to technique content.

### Dark Mode

Follow platform conventions:
- **iOS:** Automatic with system appearance. Ink Black → White, Canvas White → system background
- **Android:** Material 3 dynamic color or darkColorScheme(). Surface colors invert automatically
- Platform badge colors remain unchanged in dark mode — they are the constant visual anchors

## 7. Do's and Don'ts

### Do

- Use Ink Black (#111111) for all primary text — never pure #000000
- Let thumbnails and platform badges be the only color in the UI
- Use platform-native system fonts exclusively — SF Pro on iOS, Roboto on Android
- Keep buttons pill-shaped (30dp radius) and limited to primary/secondary variants
- Use capsule-shaped badges for platform identification with 15% opacity tinted backgrounds
- Maintain weight Medium/Semibold for all interactive text (content titles, badges, buttons)
- Use the 8dp grid — every padding, margin, and gap should snap to a multiple of 4dp
- Reserve color exclusively for semantic meaning (red=error, green=success) plus platform badges
- Use Grey-100 (#F5F5F5) for search input backgrounds and thumbnail placeholders
- Respect CJK typography: Chinese characters are visually denser, don't over-size them
- Support both light and dark mode using platform conventions

### Don't

- Don't add shadows to any surface — the elevation model is entirely flat
- Don't introduce brand colors or accent colors beyond the greyscale + platform badges
- Don't use custom fonts — system fonts handle CJK perfectly and load instantly
- Don't add decorative elements (gradient overlays, background patterns, ornamental dividers)
- Don't use more than two levels of text hierarchy per content row (title + summary)
- Don't soften the contrast — the design deliberately pushes Ink Black on white to maximum
- Don't use regular weight (400) for buttons or content titles — always use Medium/Semibold
- Don't add floating action buttons or bottom tab bars — the app is a drill-down hierarchy
- Don't round thumbnails beyond 6dp — they should feel editorial, not bubbly
- Don't animate list items on appearance — content should be immediately present, not sliding in

## 8. Device Adaptation

### Size Classes

| Class | Width | Platforms | Key Behavior |
|-------|-------|-----------|--------------|
| Compact | < 600dp | iPhone, most Android phones | Single column, standard padding (16dp) |
| Medium | 600–840dp | iPad portrait, large Android phones, foldables | Single column, increased padding (24dp) |
| Expanded | > 840dp | iPad landscape, Android tablets | Consider split view: categories left, content right |

### Touch Targets

- Minimum touch target: 44x44dp (Apple HIG) / 48x48dp (Material guidelines)
- Category rows: full width tappable, minimum 48dp height
- Content rows: full width tappable
- Back button: platform default (44pt iOS / 48dp Android)
- Search clear button: 44x44dp touch area minimum
- Platform badges: non-interactive (display only), no touch target needed

### Platform-Specific Behavior

**iOS:**
- NavigationStack with automatic large/inline title transitions
- System pull-to-refresh control
- SafariViewController for external links (when deep link fails)
- Haptic feedback on pull-to-refresh completion

**Android:**
- Material 3 TopAppBar with scroll behavior
- SwipeRefresh for pull-to-refresh
- Chrome Custom Tabs for external links (when deep link fails)
- Edge-to-edge rendering with system bar insets

### Deep Linking

Content opens in the source platform's native app when installed:
- Bilibili: `bilibili://` scheme
- Douyin: `snssdk1128://` scheme
- Xiaohongshu: `xhsdiscover://` scheme
- WeChat: `weixin://` scheme (articles only)
- YouTube: `youtube://` or `vnd.youtube:` scheme

Fallback: in-app browser (SafariViewController / Chrome Custom Tabs) with source URL.

## 9. Agent Prompt Guide

### Quick Color Reference

- Primary text / buttons: Ink Black (`#111111`)
- Screen background: Canvas White (`#FFFFFF`)
- Secondary surface / inputs: Light Gray (`#F5F5F5`)
- Secondary text / metadata: `#707072`
- Dividers / borders: `#CACACB`
- Error / sync failed: `#D30005`
- Success / sync complete: `#007D48`
- Bilibili badge: `#FB7299` text, 15% opacity background
- Xiaohongshu badge: `#FF2442` text, 15% opacity background
- Douyin badge: `#161823` text, 15% opacity background
- WeChat badge: `#07C160` text, 15% opacity background
- YouTube badge: `#FF0000` text, 15% opacity background

### Example Component Prompts

- "Build a SwiftUI category list with emoji icons (.title2) and names (.body) in rows with 16pt horizontal and 14pt vertical padding, 12pt gap between icon and name, separated by system Divider, in a NavigationStack with large title '羽球大师课'"
- "Create a Compose ContentRow with a 60x45dp thumbnail (6dp corners, Coil async loading, #F5F5F5 placeholder), title in titleMedium/semibold #111111, summary in bodyMedium #707072 (2 lines max), and a capsule platform badge with platform-colored text on 15% opacity background"
- "Design a search bar with #F5F5F5 background, 12dp corner radius, leading search icon, trailing clear button, placeholder '搜索教程' in #707072, with 300ms debounce on text changes"
- "Build a sync status bar at the bottom of the screen: horizontal layout with a 14dp circular progress indicator (2dp stroke) and labelSmall status text in #707072, fade in/out animation, three states: syncing/synced/failed"
- "Create a Compose empty state: centered Column with a 48dp icon, bodyLarge title, and bodyMedium description in onSurfaceVariant at 70% alpha, 16dp spacing between elements"

### Iteration Guide

When refining existing screens:
1. Focus on ONE component at a time
2. Reference specific color names and hex codes from this document
3. Remember: thumbnails and platform badges are the color — UI stays monochromatic greyscale
4. Use the grey scale for state changes: #F5F5F5 → #E5E5E5 → #CACACB → #707072
5. If something feels too colorful in the UI chrome, it probably is — strip it back to greyscale
6. Content titles should ALWAYS be Medium/Semibold weight — they're the primary scannable element
7. Test with Chinese text — 4-character technique names behave differently than English at the same size
8. Verify touch targets meet platform minimums (44pt iOS / 48dp Android)

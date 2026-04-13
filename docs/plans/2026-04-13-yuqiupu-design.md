# 羽球谱 (YuQiuPu) — Design Doc

## One-liner

精选羽毛球教学内容，按技术分类整理，让用户快速找到最好的学习资源。

## Positioning

像 Badminton Famly 的结构化体验，但内容来自社区 — Bilibili、小红书、抖音、微信等平台上散落的优质教学内容，被人工精选、按技术分类整理，沉淀成一本可以反复研习的"谱"。

短视频平台的内容生命周期很短，好的技术教学被算法淹没。羽球谱让这些内容沉淀下来，长期可找。

## Target User

想系统学习羽毛球技术的中文用户。打开 app，选一个技术动作，立刻看到最好的教学视频。

## MVP Scope

### Data Model

两张表：

**categories（技术分类）**
- id
- name（如：正手高远球、反手吊球、杀球、步法、发球、网前）
- icon
- sort_order
- parent_id（支持嵌套，如 正手 → 正手高远球、正手吊球）

**contents（内容）**
- id
- title
- summary（编辑笔记）
- thumbnail_url
- source_url（原始链接）
- source_platform（bilibili / xiaohongshu / douyin / wechat / youtube / other）
- author_name
- category_id
- sort_order（编辑排序）
- created_at
- updated_at

### Architecture

**Backend（Aliyun）**
- REST API（Go or Node.js）
- 两个接口：`GET /categories`，`GET /contents`
- MySQL or PostgreSQL
- Admin web panel 用于管理内容

**Native Apps**
- iOS（Swift）+ Android（Kotlin），独立代码库
- App 内嵌默认 JSON 数据（categories + contents），首次打开即可用
- 启动时从 API 拉取最新数据，成功则覆盖本地数据；失败则用本地数据

### App Screens（MVP）

1. **首页** — 技术分类列表/网格
2. **分类详情** — 该分类下的精选内容列表（缩略图 + 标题 + 摘要）
3. **点击内容** — 在 app 内浏览器打开原始链接（iOS: SFSafariViewController / Android: Chrome Custom Tabs）

### Not in MVP

- 用户账号 / 登录
- 搜索
- 作者页面
- 评论 / 收藏
- 内容推荐算法

## Competitive Landscape

- **爱羽客** — 最接近的竞品，但内容自产为主，UI 陈旧，商业化重
- **Bilibili / 小红书 / 抖音** — 内容丰富但无结构，算法驱动，内容生命周期短
- **Badminton Famly** — 国际市场，付费订阅，结构化课程，仅 web
- **Gap** — 没有产品在做"跨平台精选社区内容 + 按技术分类整理"

## Tech Decisions

| Decision | Choice | Reason |
|----------|--------|--------|
| 平台 | Native iOS + Android | 公开 app，体验优先 |
| iOS 语言 | Swift | 原生最佳体验 |
| Android 语言 | Kotlin | 原生最佳体验 |
| 后端 | Aliyun | 面向中国用户，速度优先 |
| 数据同步 | 内嵌默认数据 + API 刷新 | 首次体验好，离线可用 |
| 语言 | 中文 | 目标用户明确 |
| 内容来源 | 人工精选 | 质量可控 |

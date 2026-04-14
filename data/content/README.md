# Content-as-Code File Structure

All curated content lives under `data/content/` as plain JSON files.
No database seed files — the app compiler reads these files to build the SQLite DB.

## Directory Layout

```
data/content/
├── schemas/                          # JSON Schema definitions
│   ├── technique.schema.json
│   ├── person.schema.json
│   └── content.schema.json
├── people/                           # Author/creator profiles
│   ├── yang-chen-da-shen.json
│   └── yang-chen-da-shen.png        # Optional avatar
├── techniques/                       # Technique hierarchy (folder nesting = tree)
│   ├── basics/
│   │   ├── _technique.json           # { "name": "基本功", "icon": "🏸" }
│   │   ├── grip/
│   │   │   ├── _technique.json       # { "name": "握拍", "icon": "✋" }
│   │   │   ├── forehand-backhand-grip.json   # Content item
│   │   │   └── common-grip-mistakes.json     # Content item
│   │   └── footwork/
│   │       ├── _technique.json
│   │       └── ...
│   └── ...
├── validate.py                       # Validation script
└── README.md                         # This file
```

## Three Data Models

### Techniques

**File:** `techniques/{path}/_technique.json`

Hierarchy is expressed through folder nesting — no `parent_id`, no `sort_order`.
Every technique folder must contain a `_technique.json` file.

| Field | Type     | Required | Description          |
|-------|----------|----------|----------------------|
| name  | string   | yes      | Display name         |
| icon  | string   | yes      | Emoji icon           |

### People

**File:** `people/{slug}.json`
**Optional avatar:** `people/{slug}.png`

| Field     | Type   | Required | Description                |
|-----------|--------|----------|----------------------------|
| name      | string | yes      | Display name               |
| bio       | string | no       | Short biography            |
| platforms | object | no       | Platform URLs keyed by name |

### Content

**File:** `techniques/{path}/{slug}.json` (alongside its technique's `_technique.json`)
**Optional thumbnail:** `techniques/{path}/{slug}.png`

| Field           | Type   | Required | Description                          |
|-----------------|--------|----------|--------------------------------------|
| title           | string | yes      | Content title                        |
| summary         | string | no       | Brief description                    |
| source_url      | string | yes      | URL to original content              |
| source_platform | enum   | yes      | bilibili, xiaohongshu, douyin, wechat, youtube, other |
| person          | string | yes      | Slug reference to `people/{slug}.json` |
| difficulty      | enum   | no       | beginner, intermediate, advanced     |
| duration        | string | no       | Duration as `M:SS` or `MM:SS`        |
| editor_notes    | string | no       | Why this content is included         |

## Conventions

- **Slugs** use lowercase ASCII with hyphens: `yang-chen-da-shen`, `forehand-clear`
- **No sort_order** — display ordering is app-level logic
- **Hierarchy from folders** — a technique's parent is its containing folder
- **Person references** — content's `person` field must match a filename (without `.json`) in `people/`

## Validation

Run the validator to check all files:

```bash
python3 data/content/validate.py
```

The validator checks:
- All JSON files parse correctly and match their schema
- Every technique folder has a `_technique.json`
- Every content `person` reference points to an existing people file
- No duplicate `source_url` values across content files

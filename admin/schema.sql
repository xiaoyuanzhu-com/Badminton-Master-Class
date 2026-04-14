-- admin/schema.sql — v2
-- Used by the admin panel for fresh database creation.
-- Matches data/schema.sql; kept in sync via migrations.

CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    icon TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    parent_id INTEGER REFERENCES categories(id)
);

CREATE TABLE people (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    bio TEXT NOT NULL DEFAULT '',
    platforms_json TEXT NOT NULL DEFAULT '{}'
);

CREATE TABLE contents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    thumbnail_url TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL,
    source_platform TEXT NOT NULL CHECK(source_platform IN ('bilibili', 'xiaohongshu', 'douyin', 'wechat', 'youtube', 'other')),
    author_name TEXT NOT NULL DEFAULT '',
    person_id INTEGER REFERENCES people(id),
    difficulty TEXT NOT NULL DEFAULT '',
    duration TEXT NOT NULL DEFAULT '',
    editor_notes TEXT NOT NULL DEFAULT '',
    category_id INTEGER NOT NULL REFERENCES categories(id),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_contents_category ON contents(category_id);
CREATE INDEX idx_contents_person ON contents(person_id);

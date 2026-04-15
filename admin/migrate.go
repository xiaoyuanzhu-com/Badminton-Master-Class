package main

import (
	"database/sql"
	"fmt"
	"log"
)

// migration represents a numbered schema migration.
type migration struct {
	Version     int
	Description string
	SQL         string
}

// migrations is the ordered list of all schema migrations.
// The initial migration (v1) matches the original schema.sql exactly,
// so existing databases are recognized as already at v1.
var migrations = []migration{
	{
		Version:     1,
		Description: "initial schema: categories, contents, idx_contents_category",
		SQL: `
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    icon TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    parent_id INTEGER REFERENCES categories(id)
);

CREATE TABLE IF NOT EXISTS contents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    thumbnail_url TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL,
    source_platform TEXT NOT NULL CHECK(source_platform IN ('bilibili', 'xiaohongshu', 'douyin', 'wechat', 'youtube', 'other')),
    author_name TEXT NOT NULL DEFAULT '',
    category_id INTEGER NOT NULL REFERENCES categories(id),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_contents_category ON contents(category_id);
`,
	},
	{
		Version:     2,
		Description: "add people table, person_id/difficulty/duration/editor_notes to contents",
		SQL: `
CREATE TABLE IF NOT EXISTS people (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    bio TEXT NOT NULL DEFAULT '',
    platforms_json TEXT NOT NULL DEFAULT '{}'
);

ALTER TABLE contents ADD COLUMN person_id INTEGER REFERENCES people(id);
ALTER TABLE contents ADD COLUMN difficulty TEXT NOT NULL DEFAULT '';
ALTER TABLE contents ADD COLUMN duration TEXT NOT NULL DEFAULT '';
ALTER TABLE contents ADD COLUMN editor_notes TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_contents_person ON contents(person_id);
`,
	},
	{
		Version:     3,
		Description: "add learning_paths, path_steps, path_step_contents tables",
		SQL: `
CREATE TABLE IF NOT EXISTS learning_paths (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    difficulty TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS path_steps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path_id INTEGER NOT NULL REFERENCES learning_paths(id),
    step_order INTEGER NOT NULL DEFAULT 0,
    day INTEGER,
    title TEXT NOT NULL,
    note TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS path_step_contents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    step_id INTEGER NOT NULL REFERENCES path_steps(id),
    content_id INTEGER NOT NULL REFERENCES contents(id),
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_path_steps_path ON path_steps(path_id);
CREATE INDEX IF NOT EXISTS idx_path_step_contents_step ON path_step_contents(step_id);
`,
	},
}

// ensureSchemaVersionTable creates the schema_version table if it does not exist.
func ensureSchemaVersionTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER NOT NULL
		)
	`)
	return err
}

// getSchemaVersion returns the current schema version, or 0 if no version is recorded.
func getSchemaVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// detectExistingDB checks whether the database already has the v1 schema
// (categories and contents tables) but no schema_version table.
// This allows recognizing databases created before the migration system.
func detectExistingDB(db *sql.DB) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='table' AND name IN ('categories', 'contents')
	`).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 2, nil
}

// migrateDB applies all pending migrations to the database.
// It recognizes existing v1 databases and marks them accordingly.
func migrateDB(db *sql.DB) error {
	if err := ensureSchemaVersionTable(db); err != nil {
		return fmt.Errorf("ensure schema_version table: %w", err)
	}

	currentVersion, err := getSchemaVersion(db)
	if err != nil {
		return fmt.Errorf("get schema version: %w", err)
	}

	// If version is 0, check whether this is a pre-migration database.
	if currentVersion == 0 {
		existing, err := detectExistingDB(db)
		if err != nil {
			return fmt.Errorf("detect existing db: %w", err)
		}
		if existing {
			// Mark as v1 without re-running the migration SQL.
			if _, err := db.Exec("INSERT INTO schema_version (version) VALUES (?)", 1); err != nil {
				return fmt.Errorf("mark existing db as v1: %w", err)
			}
			currentVersion = 1
			log.Printf("Detected existing database, marked as schema version %d", currentVersion)
		}
	}

	for _, m := range migrations {
		if m.Version <= currentVersion {
			continue
		}
		log.Printf("Applying migration v%d: %s", m.Version, m.Description)
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for v%d: %w", m.Version, err)
		}
		if _, err := tx.Exec(m.SQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration v%d: %w", m.Version, err)
		}
		if _, err := tx.Exec("INSERT INTO schema_version (version) VALUES (?)", m.Version); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration v%d: %w", m.Version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration v%d: %w", m.Version, err)
		}
	}

	finalVersion, err := getSchemaVersion(db)
	if err != nil {
		return fmt.Errorf("get final schema version: %w", err)
	}
	log.Printf("Database schema at version %d", finalVersion)
	return nil
}

package main

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// v1SchemaSQL is the original v1 schema used to simulate pre-migration databases.
const v1SchemaSQL = `
CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    icon TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    parent_id INTEGER REFERENCES categories(id)
);

CREATE TABLE contents (
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

CREATE INDEX idx_contents_category ON contents(category_id);
`

func openTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	dbPath := "test_migrate_" + t.Name() + ".db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return db, func() {
		db.Close()
		os.Remove(dbPath)
	}
}

func TestMigrateDB_FreshDB(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	if err := migrateDB(db); err != nil {
		t.Fatalf("migrateDB: %v", err)
	}

	// Verify schema_version table exists and has the right version.
	version, err := getSchemaVersion(db)
	if err != nil {
		t.Fatalf("getSchemaVersion: %v", err)
	}
	if version != len(migrations) {
		t.Errorf("expected version %d, got %d", len(migrations), version)
	}

	// Verify application tables exist.
	for _, table := range []string{"categories", "contents", "people", "learning_paths", "path_steps", "path_step_contents", "schema_version"} {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %s not found: %v", table, err)
		}
	}

	// Verify indexes exist.
	for _, idx := range []string{"idx_contents_category", "idx_contents_person", "idx_path_steps_path", "idx_path_step_contents_step"} {
		var idxName string
		err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name=?", idx).Scan(&idxName)
		if err != nil {
			t.Errorf("index %s not found: %v", idx, err)
		}
	}

	// Verify v2 columns exist on contents table.
	for _, col := range []string{"person_id", "difficulty", "duration", "editor_notes"} {
		rows, err := db.Query("SELECT " + col + " FROM contents LIMIT 0")
		if err != nil {
			t.Errorf("column contents.%s not found: %v", col, err)
		} else {
			rows.Close()
		}
	}
}

func TestMigrateDB_Idempotent(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	// Run migrations twice.
	if err := migrateDB(db); err != nil {
		t.Fatalf("first migrateDB: %v", err)
	}
	if err := migrateDB(db); err != nil {
		t.Fatalf("second migrateDB: %v", err)
	}

	// Version should still be the latest (not duplicated).
	version, err := getSchemaVersion(db)
	if err != nil {
		t.Fatalf("getSchemaVersion: %v", err)
	}
	if version != len(migrations) {
		t.Errorf("expected version %d, got %d", len(migrations), version)
	}

	// Should have exactly one row per applied migration in schema_version.
	// After two runs on a fresh DB: migrations are applied once on first run, skipped on second.
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_version").Scan(&count)
	if err != nil {
		t.Fatalf("count schema_version rows: %v", err)
	}
	if count != len(migrations) {
		t.Errorf("expected %d schema_version rows, got %d", len(migrations), count)
	}
}

func TestMigrateDB_ExistingV1Database(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	// Simulate a pre-migration v1 database by creating the original v1 schema directly.
	_, err := db.Exec(v1SchemaSQL)
	if err != nil {
		t.Fatalf("create original schema: %v", err)
	}

	// Insert some test data to ensure it survives migration.
	_, err = db.Exec("INSERT INTO categories (name, icon, sort_order) VALUES ('Test', 'T', 1)")
	if err != nil {
		t.Fatalf("insert test data: %v", err)
	}

	// Run migrations — should detect existing DB as v1, then apply v2.
	if err := migrateDB(db); err != nil {
		t.Fatalf("migrateDB on existing db: %v", err)
	}

	// Version should be the latest (v2).
	version, err := getSchemaVersion(db)
	if err != nil {
		t.Fatalf("getSchemaVersion: %v", err)
	}
	if version != len(migrations) {
		t.Errorf("expected version %d, got %d", len(migrations), version)
	}

	// Test data should still be there.
	var name string
	err = db.QueryRow("SELECT name FROM categories WHERE name = 'Test'").Scan(&name)
	if err != nil {
		t.Fatalf("test data lost: %v", err)
	}
	if name != "Test" {
		t.Errorf("expected 'Test', got '%s'", name)
	}

	// v2 tables and columns should exist.
	var peopleName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='people'").Scan(&peopleName)
	if err != nil {
		t.Errorf("people table not found after v1->v2 migration: %v", err)
	}

	// Verify v2 columns on contents.
	for _, col := range []string{"person_id", "difficulty", "duration", "editor_notes"} {
		rows, err := db.Query("SELECT " + col + " FROM contents LIMIT 0")
		if err != nil {
			t.Errorf("column contents.%s not found after v1->v2 migration: %v", col, err)
		} else {
			rows.Close()
		}
	}
}

package main

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

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
	for _, table := range []string{"categories", "contents", "schema_version"} {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %s not found: %v", table, err)
		}
	}

	// Verify index exists.
	var idxName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name='idx_contents_category'").Scan(&idxName)
	if err != nil {
		t.Errorf("index idx_contents_category not found: %v", err)
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
	// After two runs on a fresh DB: v1 is applied once on first run, skipped on second.
	// But detectExistingDB won't re-insert because version is already >= 1.
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_version").Scan(&count)
	if err != nil {
		t.Fatalf("count schema_version rows: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 schema_version row, got %d", count)
	}
}

func TestMigrateDB_ExistingV1Database(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	// Simulate a pre-migration database by creating the original schema directly.
	_, err := db.Exec(schemaSQL)
	if err != nil {
		t.Fatalf("create original schema: %v", err)
	}

	// Insert some test data to ensure it survives migration.
	_, err = db.Exec("INSERT INTO categories (name, icon, sort_order) VALUES ('Test', 'T', 1)")
	if err != nil {
		t.Fatalf("insert test data: %v", err)
	}

	// Run migrations — should detect existing DB and mark as v1.
	if err := migrateDB(db); err != nil {
		t.Fatalf("migrateDB on existing db: %v", err)
	}

	// Version should be 1.
	version, err := getSchemaVersion(db)
	if err != nil {
		t.Fatalf("getSchemaVersion: %v", err)
	}
	if version != 1 {
		t.Errorf("expected version 1, got %d", version)
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
}

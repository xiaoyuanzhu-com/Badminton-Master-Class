package main

import (
	"os"
	"testing"
)

func TestInitDB(t *testing.T) {
	dbPath := "test.db"
	defer os.Remove(dbPath)

	db, err := initDB(dbPath)
	if err != nil {
		t.Fatalf("initDB failed: %v", err)
	}
	defer db.Close()

	// Verify tables exist
	tables := []string{"categories", "contents"}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %s not found: %v", table, err)
		}
	}
}

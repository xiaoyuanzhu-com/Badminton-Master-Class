package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	dbPath := "test_" + t.Name() + ".db"
	db, err := initDB(dbPath)
	if err != nil {
		t.Fatalf("initDB: %v", err)
	}
	return db, func() {
		db.Close()
		os.Remove(dbPath)
	}
}

func TestListCategories(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := db.Exec("INSERT INTO categories (name, icon, sort_order) VALUES (?, ?, ?)", "Basics", "🏸", 1)
	if err != nil {
		t.Fatalf("insert category: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	w := httptest.NewRecorder()
	categoriesHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Basics") {
		t.Errorf("response body should contain 'Basics', got: %s", w.Body.String())
	}
}

func TestCreateCategory(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Set("name", "Serve")
	form.Set("icon", "🎾")
	form.Set("sort_order", "2")

	req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	categoriesHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM categories WHERE name = ?", "Serve").Scan(&count)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

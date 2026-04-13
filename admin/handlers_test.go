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

func TestListContents(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := db.Exec("INSERT INTO categories (id, name, icon, sort_order) VALUES (1, 'Basics', '🏸', 1)")
	if err != nil {
		t.Fatalf("insert category: %v", err)
	}
	_, err = db.Exec(`INSERT INTO contents (title, summary, source_url, source_platform, author_name, category_id, sort_order)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, "Grip Tutorial", "How to grip", "https://example.com/1", "bilibili", "Coach Li", 1, 1)
	if err != nil {
		t.Fatalf("insert content: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/contents", nil)
	w := httptest.NewRecorder()
	contentsHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Grip Tutorial") {
		t.Errorf("response body should contain 'Grip Tutorial', got: %s", w.Body.String())
	}
}

func TestCreateContent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := db.Exec("INSERT INTO categories (id, name, icon, sort_order) VALUES (1, 'Basics', '🏸', 1)")
	if err != nil {
		t.Fatalf("insert category: %v", err)
	}

	form := url.Values{}
	form.Set("title", "Smash Guide")
	form.Set("summary", "How to smash")
	form.Set("thumbnail_url", "https://example.com/thumb.jpg")
	form.Set("source_url", "https://example.com/2")
	form.Set("source_platform", "bilibili")
	form.Set("author_name", "Coach Wang")
	form.Set("category_id", "1")
	form.Set("sort_order", "1")

	req := httptest.NewRequest(http.MethodPost, "/contents", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	contentsHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM contents WHERE title = ?", "Smash Guide").Scan(&count)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

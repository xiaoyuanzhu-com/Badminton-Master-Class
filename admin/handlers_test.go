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

func TestEditCategoryGET(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := db.Exec("INSERT INTO categories (id, name, icon, sort_order) VALUES (1, 'Basics', '🏸', 1)")
	if err != nil {
		t.Fatalf("insert category: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/categories/1/edit", nil)
	w := httptest.NewRecorder()
	categoryActionHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "Basics") {
		t.Errorf("response should contain 'Basics', got: %s", body)
	}
}

func TestEditCategoryGET_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/categories/999/edit", nil)
	w := httptest.NewRecorder()
	categoryActionHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestEditCategoryPOST(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := db.Exec("INSERT INTO categories (id, name, icon, sort_order) VALUES (1, 'Basics', '🏸', 1)")
	if err != nil {
		t.Fatalf("insert category: %v", err)
	}

	form := url.Values{}
	form.Set("name", "Advanced")
	form.Set("icon", "🎯")
	form.Set("sort_order", "5")

	req := httptest.NewRequest(http.MethodPost, "/categories/1/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	categoryActionHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d: %s", w.Code, w.Body.String())
	}

	var name, icon string
	var sortOrder int
	err = db.QueryRow("SELECT name, icon, sort_order FROM categories WHERE id = 1").Scan(&name, &icon, &sortOrder)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if name != "Advanced" {
		t.Errorf("expected name 'Advanced', got '%s'", name)
	}
	if icon != "🎯" {
		t.Errorf("expected icon '🎯', got '%s'", icon)
	}
	if sortOrder != 5 {
		t.Errorf("expected sort_order 5, got %d", sortOrder)
	}
}

func TestEditCategoryPOST_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Set("name", "Ghost")
	form.Set("icon", "👻")
	form.Set("sort_order", "1")

	req := httptest.NewRequest(http.MethodPost, "/categories/999/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	categoryActionHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDeleteCategory(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := db.Exec("INSERT INTO categories (id, name, icon, sort_order) VALUES (1, 'Basics', '🏸', 1)")
	if err != nil {
		t.Fatalf("insert category: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/categories/1/delete", nil)
	w := httptest.NewRecorder()
	categoryActionHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d: %s", w.Code, w.Body.String())
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM categories WHERE id = 1").Scan(&count)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows, got %d", count)
	}
}

func TestDeleteCategory_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/categories/999/delete", nil)
	w := httptest.NewRecorder()
	categoryActionHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDeleteCategory_WithChildren(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := db.Exec("INSERT INTO categories (id, name, icon, sort_order) VALUES (1, 'Parent', '📁', 1)")
	if err != nil {
		t.Fatalf("insert parent: %v", err)
	}
	_, err = db.Exec("INSERT INTO categories (id, name, icon, sort_order, parent_id) VALUES (2, 'Child', '📄', 2, 1)")
	if err != nil {
		t.Fatalf("insert child: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/categories/1/delete", nil)
	w := httptest.NewRecorder()
	categoryActionHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "child categories") {
		t.Errorf("expected error about child categories, got: %s", w.Body.String())
	}
}

func TestDeleteCategory_WithContent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := db.Exec("INSERT INTO categories (id, name, icon, sort_order) VALUES (1, 'Basics', '🏸', 1)")
	if err != nil {
		t.Fatalf("insert category: %v", err)
	}
	_, err = db.Exec(`INSERT INTO contents (title, summary, source_url, source_platform, author_name, category_id, sort_order)
		VALUES ('Video', 'A video', 'https://example.com/1', 'bilibili', 'Coach', 1, 1)`)
	if err != nil {
		t.Fatalf("insert content: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/categories/1/delete", nil)
	w := httptest.NewRecorder()
	categoryActionHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "associated content") {
		t.Errorf("expected error about associated content, got: %s", w.Body.String())
	}
}

func TestDeleteCategory_MethodNotAllowed(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/categories/1/delete", nil)
	w := httptest.NewRecorder()
	categoryActionHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestCategoryActionHandler_InvalidPath(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/categories/notanumber/edit", nil)
	w := httptest.NewRecorder()
	categoryActionHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestEditContentGET(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := db.Exec("INSERT INTO categories (id, name, icon, sort_order) VALUES (1, 'Basics', '🏸', 1)")
	if err != nil {
		t.Fatalf("insert category: %v", err)
	}
	_, err = db.Exec(`INSERT INTO contents (id, title, summary, source_url, source_platform, author_name, category_id, sort_order)
		VALUES (1, 'Grip Tutorial', 'How to grip', 'https://example.com/1', 'bilibili', 'Coach Li', 1, 1)`)
	if err != nil {
		t.Fatalf("insert content: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/contents/1/edit", nil)
	w := httptest.NewRecorder()
	contentActionHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "Grip Tutorial") {
		t.Errorf("response should contain 'Grip Tutorial', got: %s", body)
	}
}

func TestEditContentGET_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/contents/999/edit", nil)
	w := httptest.NewRecorder()
	contentActionHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestEditContentPOST(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := db.Exec("INSERT INTO categories (id, name, icon, sort_order) VALUES (1, 'Basics', '🏸', 1)")
	if err != nil {
		t.Fatalf("insert category: %v", err)
	}
	_, err = db.Exec("INSERT INTO categories (id, name, icon, sort_order) VALUES (2, 'Advanced', '🎯', 2)")
	if err != nil {
		t.Fatalf("insert category 2: %v", err)
	}
	_, err = db.Exec(`INSERT INTO contents (id, title, summary, source_url, source_platform, author_name, category_id, sort_order)
		VALUES (1, 'Grip Tutorial', 'How to grip', 'https://example.com/1', 'bilibili', 'Coach Li', 1, 1)`)
	if err != nil {
		t.Fatalf("insert content: %v", err)
	}

	form := url.Values{}
	form.Set("title", "Updated Grip Tutorial")
	form.Set("summary", "Updated summary")
	form.Set("thumbnail_url", "https://example.com/thumb2.jpg")
	form.Set("source_url", "https://example.com/updated")
	form.Set("source_platform", "youtube")
	form.Set("author_name", "Coach Wang")
	form.Set("category_id", "2")
	form.Set("sort_order", "5")

	req := httptest.NewRequest(http.MethodPost, "/contents/1/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	contentActionHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d: %s", w.Code, w.Body.String())
	}

	var title, summary, sourceURL, sourcePlatform, authorName string
	var categoryID, sortOrder int
	err = db.QueryRow(`SELECT title, summary, source_url, source_platform, author_name, category_id, sort_order
		FROM contents WHERE id = 1`).Scan(&title, &summary, &sourceURL, &sourcePlatform, &authorName, &categoryID, &sortOrder)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if title != "Updated Grip Tutorial" {
		t.Errorf("expected title 'Updated Grip Tutorial', got '%s'", title)
	}
	if sourcePlatform != "youtube" {
		t.Errorf("expected platform 'youtube', got '%s'", sourcePlatform)
	}
	if categoryID != 2 {
		t.Errorf("expected category_id 2, got %d", categoryID)
	}
	if sortOrder != 5 {
		t.Errorf("expected sort_order 5, got %d", sortOrder)
	}
}

func TestEditContentPOST_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Set("title", "Ghost")
	form.Set("source_url", "https://example.com/ghost")
	form.Set("source_platform", "bilibili")
	form.Set("category_id", "1")
	form.Set("sort_order", "1")

	req := httptest.NewRequest(http.MethodPost, "/contents/999/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	contentActionHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDeleteContent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := db.Exec("INSERT INTO categories (id, name, icon, sort_order) VALUES (1, 'Basics', '🏸', 1)")
	if err != nil {
		t.Fatalf("insert category: %v", err)
	}
	_, err = db.Exec(`INSERT INTO contents (id, title, summary, source_url, source_platform, author_name, category_id, sort_order)
		VALUES (1, 'Grip Tutorial', 'How to grip', 'https://example.com/1', 'bilibili', 'Coach Li', 1, 1)`)
	if err != nil {
		t.Fatalf("insert content: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/contents/1/delete", nil)
	w := httptest.NewRecorder()
	contentActionHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d: %s", w.Code, w.Body.String())
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM contents WHERE id = 1").Scan(&count)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows, got %d", count)
	}
}

func TestDeleteContent_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/contents/999/delete", nil)
	w := httptest.NewRecorder()
	contentActionHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDeleteContent_MethodNotAllowed(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/contents/1/delete", nil)
	w := httptest.NewRecorder()
	contentActionHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestContentActionHandler_InvalidPath(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/contents/notanumber/edit", nil)
	w := httptest.NewRecorder()
	contentActionHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestBasicAuth_NoCredentials(t *testing.T) {
	db, cleanup := setupTestDB(t)
	dbPath := "test_" + t.Name() + ".db"
	defer cleanup()

	mux := setupRoutes(db, dbPath)
	handler := basicAuth(mux, "admin", "admin")

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if w.Header().Get("WWW-Authenticate") == "" {
		t.Error("expected WWW-Authenticate header")
	}
}

func TestBasicAuth_WrongCredentials(t *testing.T) {
	db, cleanup := setupTestDB(t)
	dbPath := "test_" + t.Name() + ".db"
	defer cleanup()

	mux := setupRoutes(db, dbPath)
	handler := basicAuth(mux, "admin", "admin")

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	req.SetBasicAuth("admin", "wrongpass")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestBasicAuth_ValidCredentials(t *testing.T) {
	db, cleanup := setupTestDB(t)
	dbPath := "test_" + t.Name() + ".db"
	defer cleanup()

	mux := setupRoutes(db, dbPath)
	handler := basicAuth(mux, "admin", "admin")

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	req.SetBasicAuth("admin", "admin")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestExportDB(t *testing.T) {
	db, cleanup := setupTestDB(t)
	dbPath := "test_" + t.Name() + ".db"
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/export", nil)
	w := httptest.NewRecorder()
	exportHandler(db, dbPath).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if ct != "application/x-sqlite3" {
		t.Errorf("expected content-type application/x-sqlite3, got %s", ct)
	}
	if w.Body.Len() == 0 {
		t.Error("expected non-empty body")
	}
}

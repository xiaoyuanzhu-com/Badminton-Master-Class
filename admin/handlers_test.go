package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
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

// insertTestData populates the DB with categories, people, and contents for tests.
func insertTestData(t *testing.T, db *sql.DB) {
	t.Helper()
	stmts := []string{
		"INSERT INTO categories (id, name, icon, sort_order) VALUES (1, 'Basics', '🏸', 1)",
		"INSERT INTO categories (id, name, icon, sort_order, parent_id) VALUES (2, 'Grip', '🤚', 2, 1)",
		"INSERT INTO people (id, slug, name, bio) VALUES (1, 'coach-li', 'Coach Li', 'Professional coach')",
		`INSERT INTO contents (id, title, summary, source_url, source_platform, author_name,
			person_id, difficulty, duration, editor_notes, category_id, sort_order)
			VALUES (1, 'Grip Tutorial', 'How to grip a racket', 'https://example.com/1', 'bilibili', 'Coach Li',
			1, 'beginner', '10:30', 'Great intro video', 1, 1)`,
		`INSERT INTO contents (id, title, summary, source_url, source_platform, author_name,
			category_id, sort_order)
			VALUES (2, 'Smash Guide', 'How to smash', 'https://example.com/2', 'youtube', 'Coach Wang',
			1, 2)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("insert: %v\nSQL: %s", err, s)
		}
	}
}

// ── Home ────────────────────────────────────────────────────────────

func TestHome(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	insertTestData(t, db)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	homeHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Basics") {
		t.Errorf("should contain 'Basics', got: %s", body)
	}
}

func TestHome_NotFoundForOtherPaths(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	homeHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ── Categories ──────────────────────────────────────────────────────

func TestListCategories(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	insertTestData(t, db)

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	w := httptest.NewRecorder()
	categoriesHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Basics") {
		t.Errorf("should contain 'Basics', got: %s", body)
	}
	if !strings.Contains(body, "Grip") {
		t.Errorf("should contain 'Grip', got: %s", body)
	}
}

func TestCategories_MethodNotAllowed(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/categories", nil)
	w := httptest.NewRecorder()
	categoriesHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// ── Contents list ───────────────────────────────────────────────────

func TestListContents(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	insertTestData(t, db)

	req := httptest.NewRequest(http.MethodGet, "/contents", nil)
	w := httptest.NewRecorder()
	contentsHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Grip Tutorial") {
		t.Errorf("should contain 'Grip Tutorial', got: %s", body)
	}
	if !strings.Contains(body, "Smash Guide") {
		t.Errorf("should contain 'Smash Guide', got: %s", body)
	}
}

func TestListContents_FilterByCategory(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	insertTestData(t, db)

	// Add a second category with different content
	db.Exec("INSERT INTO categories (id, name, icon, sort_order) VALUES (3, 'Advanced', '🎯', 3)")
	db.Exec(`INSERT INTO contents (id, title, summary, source_url, source_platform, author_name, category_id, sort_order)
		VALUES (3, 'Advanced Smash', 'Pro smash', 'https://example.com/3', 'youtube', 'Coach X', 3, 1)`)

	req := httptest.NewRequest(http.MethodGet, "/contents?category_id=3", nil)
	w := httptest.NewRecorder()
	contentsHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Advanced Smash") {
		t.Errorf("should contain 'Advanced Smash', got: %s", body)
	}
	if strings.Contains(body, "Grip Tutorial") {
		t.Errorf("should not contain 'Grip Tutorial' when filtering by category 3")
	}
}

func TestContents_MethodNotAllowed(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/contents", nil)
	w := httptest.NewRecorder()
	contentsHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// ── Content detail ──────────────────────────────────────────────────

func TestContentDetail(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	insertTestData(t, db)

	req := httptest.NewRequest(http.MethodGet, "/contents/1", nil)
	w := httptest.NewRecorder()
	contentDetailHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Grip Tutorial") {
		t.Errorf("should contain 'Grip Tutorial', got: %s", body)
	}
	if !strings.Contains(body, "Coach Li") {
		t.Errorf("should contain 'Coach Li', got: %s", body)
	}
}

func TestContentDetail_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/contents/999", nil)
	w := httptest.NewRecorder()
	contentDetailHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestContentDetail_InvalidID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/contents/abc", nil)
	w := httptest.NewRecorder()
	contentDetailHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ── People list ─────────────────────────────────────────────────────

func TestPeopleList(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	insertTestData(t, db)

	req := httptest.NewRequest(http.MethodGet, "/people", nil)
	w := httptest.NewRecorder()
	peopleListHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Coach Li") {
		t.Errorf("should contain 'Coach Li', got: %s", body)
	}
}

func TestPeople_MethodNotAllowed(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/people", nil)
	w := httptest.NewRecorder()
	peopleListHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// ── Person detail ───────────────────────────────────────────────────

func TestPersonDetail(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	insertTestData(t, db)

	req := httptest.NewRequest(http.MethodGet, "/people/1", nil)
	w := httptest.NewRecorder()
	personDetailHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Coach Li") {
		t.Errorf("should contain 'Coach Li', got: %s", body)
	}
	if !strings.Contains(body, "Grip Tutorial") {
		t.Errorf("should contain associated content 'Grip Tutorial', got: %s", body)
	}
}

func TestPersonDetail_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/people/999", nil)
	w := httptest.NewRecorder()
	personDetailHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestPersonDetail_InvalidID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/people/abc", nil)
	w := httptest.NewRecorder()
	personDetailHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ── Search ──────────────────────────────────────────────────────────

func TestSearch_Contents(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	insertTestData(t, db)

	req := httptest.NewRequest(http.MethodGet, "/search?q=Grip", nil)
	w := httptest.NewRecorder()
	searchHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Grip Tutorial") {
		t.Errorf("should find 'Grip Tutorial', got: %s", body)
	}
}

func TestSearch_People(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	insertTestData(t, db)

	req := httptest.NewRequest(http.MethodGet, "/search?q=Coach", nil)
	w := httptest.NewRecorder()
	searchHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Coach Li") {
		t.Errorf("should find 'Coach Li' in people results, got: %s", body)
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/search?q=", nil)
	w := httptest.NewRecorder()
	searchHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestSearch_NoResults(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	insertTestData(t, db)

	req := httptest.NewRequest(http.MethodGet, "/search?q=nonexistent999", nil)
	w := httptest.NewRecorder()
	searchHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// ── Auth ────────────────────────────────────────────────────────────

func TestAuthDisabledByDefault(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mux := setupRoutes(db)

	// Without auth wrapper, should serve directly
	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 without auth, got %d", w.Code)
	}
}

func TestBasicAuth_NoCredentials(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mux := setupRoutes(db)
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
	defer cleanup()

	mux := setupRoutes(db)
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
	defer cleanup()

	mux := setupRoutes(db)
	handler := basicAuth(mux, "admin", "admin")

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	req.SetBasicAuth("admin", "admin")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

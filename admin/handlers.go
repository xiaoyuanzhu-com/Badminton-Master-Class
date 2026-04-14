package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Category struct {
	ID        int
	Name      string
	Icon      string
	SortOrder int
	ParentID  sql.NullInt64
}

type Content struct {
	ID             int
	Title          string
	Summary        string
	ThumbnailURL   string
	SourceURL      string
	SourcePlatform string
	AuthorName     string
	CategoryID     int
	CategoryName   string
	SortOrder      int
}

func categoriesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			rows, err := db.Query("SELECT id, name, icon, sort_order, parent_id FROM categories ORDER BY sort_order")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var cats []Category
			for rows.Next() {
				var c Category
				if err := rows.Scan(&c.ID, &c.Name, &c.Icon, &c.SortOrder, &c.ParentID); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				cats = append(cats, c)
			}
			if err := rows.Err(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			tmpl, err := template.ParseFiles("templates/categories.html")
			if err != nil {
				// Fallback: plain text output for testing
				w.Header().Set("Content-Type", "text/plain")
				for _, c := range cats {
					fmt.Fprintf(w, "Category: %s (icon=%s, sort=%d)\n", c.Name, c.Icon, c.SortOrder)
				}
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			tmpl.Execute(w, cats)

		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			name := r.FormValue("name")
			icon := r.FormValue("icon")
			sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))
			parentIDStr := r.FormValue("parent_id")

			var parentID sql.NullInt64
			if parentIDStr != "" {
				pid, err := strconv.ParseInt(parentIDStr, 10, 64)
				if err == nil {
					parentID = sql.NullInt64{Int64: pid, Valid: true}
				}
			}

			_, err := db.Exec(
				"INSERT INTO categories (name, icon, sort_order, parent_id) VALUES (?, ?, ?, ?)",
				name, icon, sortOrder, parentID,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/categories", http.StatusSeeOther)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func contentsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			rows, err := db.Query(`
				SELECT c.id, c.title, c.summary, c.thumbnail_url, c.source_url,
				       c.source_platform, c.author_name, c.category_id,
				       COALESCE(cat.name, ''), c.sort_order
				FROM contents c
				LEFT JOIN categories cat ON c.category_id = cat.id
				ORDER BY c.sort_order`)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var contents []Content
			for rows.Next() {
				var ct Content
				if err := rows.Scan(&ct.ID, &ct.Title, &ct.Summary, &ct.ThumbnailURL,
					&ct.SourceURL, &ct.SourcePlatform, &ct.AuthorName,
					&ct.CategoryID, &ct.CategoryName, &ct.SortOrder); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				contents = append(contents, ct)
			}
			if err := rows.Err(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Also load categories for the dropdown
			catRows, err := db.Query("SELECT id, name FROM categories ORDER BY sort_order")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer catRows.Close()

			var cats []Category
			for catRows.Next() {
				var c Category
				if err := catRows.Scan(&c.ID, &c.Name); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				cats = append(cats, c)
			}

			data := struct {
				Contents   []Content
				Categories []Category
			}{contents, cats}

			tmpl, err := template.ParseFiles("templates/contents.html")
			if err != nil {
				// Fallback: plain text output for testing
				w.Header().Set("Content-Type", "text/plain")
				for _, ct := range contents {
					fmt.Fprintf(w, "Content: %s (category=%s, platform=%s)\n", ct.Title, ct.CategoryName, ct.SourcePlatform)
				}
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			tmpl.Execute(w, data)

		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			title := r.FormValue("title")
			summary := r.FormValue("summary")
			thumbnailURL := r.FormValue("thumbnail_url")
			sourceURL := r.FormValue("source_url")
			sourcePlatform := r.FormValue("source_platform")
			authorName := r.FormValue("author_name")
			categoryID, _ := strconv.Atoi(r.FormValue("category_id"))
			sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))

			_, err := db.Exec(
				`INSERT INTO contents (title, summary, thumbnail_url, source_url, source_platform, author_name, category_id, sort_order)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				title, summary, thumbnailURL, sourceURL, sourcePlatform, authorName, categoryID, sortOrder,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/contents", http.StatusSeeOther)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// contentActionHandler routes /contents/{id}/edit and /contents/{id}/delete.
func contentActionHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/contents/")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}

		id, err := strconv.Atoi(parts[0])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		switch parts[1] {
		case "edit":
			contentEditHandler(db, id, w, r)
		case "delete":
			contentDeleteHandler(db, id, w, r)
		default:
			http.NotFound(w, r)
		}
	}
}

func contentEditHandler(db *sql.DB, id int, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var ct Content
		err := db.QueryRow(`SELECT id, title, summary, thumbnail_url, source_url,
			source_platform, author_name, category_id, sort_order
			FROM contents WHERE id = ?`, id).
			Scan(&ct.ID, &ct.Title, &ct.Summary, &ct.ThumbnailURL,
				&ct.SourceURL, &ct.SourcePlatform, &ct.AuthorName,
				&ct.CategoryID, &ct.SortOrder)
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Load categories for dropdown
		rows, err := db.Query("SELECT id, name FROM categories ORDER BY sort_order")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var cats []Category
		for rows.Next() {
			var c Category
			if err := rows.Scan(&c.ID, &c.Name); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			cats = append(cats, c)
		}

		data := struct {
			Content    Content
			Categories []Category
		}{ct, cats}

		tmpl, err := template.ParseFiles("templates/content_edit.html")
		if err != nil {
			// Fallback: plain text for testing
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "Edit Content: %s (id=%d, platform=%s, sort=%d)", ct.Title, ct.ID, ct.SourcePlatform, ct.SortOrder)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, data)

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		title := r.FormValue("title")
		summary := r.FormValue("summary")
		thumbnailURL := r.FormValue("thumbnail_url")
		sourceURL := r.FormValue("source_url")
		sourcePlatform := r.FormValue("source_platform")
		authorName := r.FormValue("author_name")
		categoryID, _ := strconv.Atoi(r.FormValue("category_id"))
		sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))

		result, err := db.Exec(
			`UPDATE contents SET title = ?, summary = ?, thumbnail_url = ?, source_url = ?,
			 source_platform = ?, author_name = ?, category_id = ?, sort_order = ?,
			 updated_at = datetime('now') WHERE id = ?`,
			title, summary, thumbnailURL, sourceURL, sourcePlatform, authorName, categoryID, sortOrder, id,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/contents", http.StatusSeeOther)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func contentDeleteHandler(db *sql.DB, id int, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := db.Exec("DELETE FROM contents WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/contents", http.StatusSeeOther)
}

// categoryActionHandler routes /categories/{id}/edit and /categories/{id}/delete.
func categoryActionHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse path: /categories/{id}/edit or /categories/{id}/delete
		path := strings.TrimPrefix(r.URL.Path, "/categories/")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}

		id, err := strconv.Atoi(parts[0])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		switch parts[1] {
		case "edit":
			categoryEditHandler(db, id, w, r)
		case "delete":
			categoryDeleteHandler(db, id, w, r)
		default:
			http.NotFound(w, r)
		}
	}
}

func categoryEditHandler(db *sql.DB, id int, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var c Category
		err := db.QueryRow("SELECT id, name, icon, sort_order, parent_id FROM categories WHERE id = ?", id).
			Scan(&c.ID, &c.Name, &c.Icon, &c.SortOrder, &c.ParentID)
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Load all categories for parent dropdown (exclude self)
		rows, err := db.Query("SELECT id, name FROM categories WHERE id != ? ORDER BY sort_order", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var cats []Category
		for rows.Next() {
			var cat Category
			if err := rows.Scan(&cat.ID, &cat.Name); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			cats = append(cats, cat)
		}

		parentIDValue := 0
		if c.ParentID.Valid {
			parentIDValue = int(c.ParentID.Int64)
		}

		data := struct {
			Category      Category
			Categories    []Category
			ParentIDValue int
		}{c, cats, parentIDValue}

		tmpl, err := template.ParseFiles("templates/category_edit.html")
		if err != nil {
			// Fallback: plain text for testing
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "Edit Category: %s (id=%d, icon=%s, sort=%d)", c.Name, c.ID, c.Icon, c.SortOrder)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, data)

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		name := r.FormValue("name")
		icon := r.FormValue("icon")
		sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))
		parentIDStr := r.FormValue("parent_id")

		var parentID sql.NullInt64
		if parentIDStr != "" {
			pid, err := strconv.ParseInt(parentIDStr, 10, 64)
			if err == nil {
				parentID = sql.NullInt64{Int64: pid, Valid: true}
			}
		}

		result, err := db.Exec(
			"UPDATE categories SET name = ?, icon = ?, sort_order = ?, parent_id = ? WHERE id = ?",
			name, icon, sortOrder, parentID, id,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/categories", http.StatusSeeOther)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func categoryDeleteHandler(db *sql.DB, id int, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check for child categories
	var childCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM categories WHERE parent_id = ?", id).Scan(&childCount); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if childCount > 0 {
		http.Error(w, "Cannot delete category: it has child categories. Remove or reassign them first.", http.StatusConflict)
		return
	}

	// Check for associated contents
	var contentCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM contents WHERE category_id = ?", id).Scan(&contentCount); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if contentCount > 0 {
		http.Error(w, "Cannot delete category: it has associated content. Remove or reassign the content first.", http.StatusConflict)
		return
	}

	result, err := db.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

func exportHandler(db *sql.DB, dbPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Upload to Aliyun OSS in the background (best-effort).
		tryUploadToOSS(dbPath)

		f, err := os.Open(dbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		w.Header().Set("Content-Type", "application/x-sqlite3")
		w.Header().Set("Content-Disposition", "attachment; filename=badminton-master-class.db")
		io.Copy(w, f)
	}
}

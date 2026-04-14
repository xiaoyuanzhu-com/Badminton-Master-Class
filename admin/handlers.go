package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

// ── Models ──────────────────────────────────────────────────────────

type Category struct {
	ID           int
	Name         string
	Icon         string
	SortOrder    int
	ParentID     sql.NullInt64
	ParentName   string
	ContentCount int
	Children     []Category
}

type Person struct {
	ID            int
	Slug          string
	Name          string
	Bio           string
	PlatformsJSON string
	ContentCount  int
}

type Content struct {
	ID             int
	Title          string
	Summary        string
	ThumbnailURL   string
	SourceURL      string
	SourcePlatform string
	AuthorName     string
	PersonID       sql.NullInt64
	PersonName     string
	Difficulty     string
	Duration       string
	EditorNotes    string
	CategoryID     int
	CategoryName   string
	SortOrder      int
}

// ── Template helpers ────────────────────────────────────────────────

var funcMap = template.FuncMap{
	"platformLabel": func(p string) string {
		switch p {
		case "bilibili":
			return "Bilibili"
		case "xiaohongshu":
			return "小红书"
		case "douyin":
			return "抖音"
		case "wechat":
			return "微信"
		case "youtube":
			return "YouTube"
		default:
			return "其他"
		}
	},
	"difficultyLabel": func(d string) string {
		switch d {
		case "beginner":
			return "入门"
		case "intermediate":
			return "进阶"
		case "advanced":
			return "高级"
		default:
			return d
		}
	},
}

func parseTemplate(name string) (*template.Template, error) {
	return template.New(name).Funcs(funcMap).ParseFiles("templates/" + name)
}

// ── Handlers ────────────────────────────────────────────────────────

// homeHandler shows the home page with category hierarchy.
func homeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		rows, err := db.Query(`
			SELECT c.id, c.name, c.icon, c.sort_order, c.parent_id,
			       (SELECT COUNT(*) FROM contents WHERE category_id = c.id) AS content_count
			FROM categories c
			ORDER BY c.sort_order`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var allCats []Category
		for rows.Next() {
			var c Category
			if err := rows.Scan(&c.ID, &c.Name, &c.Icon, &c.SortOrder, &c.ParentID, &c.ContentCount); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			allCats = append(allCats, c)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Build hierarchy: top-level categories with children
		childMap := make(map[int64][]Category)
		var topLevel []Category
		for _, c := range allCats {
			if c.ParentID.Valid {
				childMap[c.ParentID.Int64] = append(childMap[c.ParentID.Int64], c)
			} else {
				topLevel = append(topLevel, c)
			}
		}
		for i := range topLevel {
			topLevel[i].Children = childMap[int64(topLevel[i].ID)]
		}

		tmpl, err := parseTemplate("home.html")
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			for _, c := range topLevel {
				w.Write([]byte("Category: " + c.Name + "\n"))
			}
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, topLevel)
	}
}

// categoriesHandler shows a list of all categories.
func categoriesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rows, err := db.Query(`
			SELECT c.id, c.name, c.icon, c.sort_order, c.parent_id,
			       COALESCE(p.name, '') AS parent_name,
			       (SELECT COUNT(*) FROM contents WHERE category_id = c.id) AS content_count
			FROM categories c
			LEFT JOIN categories p ON c.parent_id = p.id
			ORDER BY c.sort_order`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var cats []Category
		for rows.Next() {
			var c Category
			if err := rows.Scan(&c.ID, &c.Name, &c.Icon, &c.SortOrder, &c.ParentID, &c.ParentName, &c.ContentCount); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			cats = append(cats, c)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl, err := parseTemplate("categories.html")
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			for _, c := range cats {
				w.Write([]byte("Category: " + c.Name + "\n"))
			}
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, cats)
	}
}

// contentsHandler shows contents, optionally filtered by category.
func contentsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		query := `
			SELECT c.id, c.title, c.summary, c.thumbnail_url, c.source_url,
			       c.source_platform, c.author_name, c.person_id,
			       COALESCE(p.name, c.author_name) AS person_name,
			       c.difficulty, c.duration, c.category_id,
			       COALESCE(cat.name, ''), c.sort_order
			FROM contents c
			LEFT JOIN categories cat ON c.category_id = cat.id
			LEFT JOIN people p ON c.person_id = p.id`

		var args []interface{}
		catIDStr := r.URL.Query().Get("category_id")
		var catName string
		if catIDStr != "" {
			catID, err := strconv.Atoi(catIDStr)
			if err == nil {
				query += " WHERE c.category_id = ?"
				args = append(args, catID)
				_ = db.QueryRow("SELECT name FROM categories WHERE id = ?", catID).Scan(&catName)
			}
		}
		query += " ORDER BY c.sort_order"

		rows, err := db.Query(query, args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var contents []Content
		for rows.Next() {
			var ct Content
			if err := rows.Scan(&ct.ID, &ct.Title, &ct.Summary, &ct.ThumbnailURL,
				&ct.SourceURL, &ct.SourcePlatform, &ct.AuthorName, &ct.PersonID,
				&ct.PersonName, &ct.Difficulty, &ct.Duration,
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

		data := struct {
			Contents     []Content
			CategoryName string
			CategoryID   string
		}{contents, catName, catIDStr}

		tmpl, err := parseTemplate("contents.html")
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			for _, ct := range contents {
				w.Write([]byte("Content: " + ct.Title + " (category=" + ct.CategoryName + ")\n"))
			}
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, data)
	}
}

// contentDetailHandler shows a single content item.
func contentDetailHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/contents/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		var ct Content
		err = db.QueryRow(`
			SELECT c.id, c.title, c.summary, c.thumbnail_url, c.source_url,
			       c.source_platform, c.author_name, c.person_id,
			       COALESCE(p.name, c.author_name) AS person_name,
			       c.difficulty, c.duration, c.editor_notes,
			       c.category_id, COALESCE(cat.name, ''), c.sort_order
			FROM contents c
			LEFT JOIN categories cat ON c.category_id = cat.id
			LEFT JOIN people p ON c.person_id = p.id
			WHERE c.id = ?`, id).
			Scan(&ct.ID, &ct.Title, &ct.Summary, &ct.ThumbnailURL,
				&ct.SourceURL, &ct.SourcePlatform, &ct.AuthorName, &ct.PersonID,
				&ct.PersonName, &ct.Difficulty, &ct.Duration, &ct.EditorNotes,
				&ct.CategoryID, &ct.CategoryName, &ct.SortOrder)
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl, err := parseTemplate("content_detail.html")
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("Content: " + ct.Title))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, ct)
	}
}

// peopleListHandler shows all people.
func peopleListHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rows, err := db.Query(`
			SELECT p.id, p.slug, p.name, p.bio, p.platforms_json,
			       (SELECT COUNT(*) FROM contents WHERE person_id = p.id) AS content_count
			FROM people p
			ORDER BY p.name`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var people []Person
		for rows.Next() {
			var p Person
			if err := rows.Scan(&p.ID, &p.Slug, &p.Name, &p.Bio, &p.PlatformsJSON, &p.ContentCount); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			people = append(people, p)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl, err := parseTemplate("people.html")
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			for _, p := range people {
				w.Write([]byte("Person: " + p.Name + "\n"))
			}
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, people)
	}
}

// personDetailHandler shows a single person with their contents.
func personDetailHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/people/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		var p Person
		err = db.QueryRow("SELECT id, slug, name, bio, platforms_json FROM people WHERE id = ?", id).
			Scan(&p.ID, &p.Slug, &p.Name, &p.Bio, &p.PlatformsJSON)
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Load their content
		rows, err := db.Query(`
			SELECT c.id, c.title, c.summary, c.thumbnail_url, c.source_url,
			       c.source_platform, c.author_name, c.difficulty, c.duration,
			       c.category_id, COALESCE(cat.name, ''), c.sort_order
			FROM contents c
			LEFT JOIN categories cat ON c.category_id = cat.id
			WHERE c.person_id = ?
			ORDER BY c.sort_order`, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var contents []Content
		for rows.Next() {
			var ct Content
			if err := rows.Scan(&ct.ID, &ct.Title, &ct.Summary, &ct.ThumbnailURL,
				&ct.SourceURL, &ct.SourcePlatform, &ct.AuthorName, &ct.Difficulty,
				&ct.Duration, &ct.CategoryID, &ct.CategoryName, &ct.SortOrder); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			contents = append(contents, ct)
		}

		data := struct {
			Person   Person
			Contents []Content
		}{p, contents}

		tmpl, err := parseTemplate("person_detail.html")
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("Person: " + p.Name + "\n"))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, data)
	}
}

// searchHandler searches across contents, categories, and people.
func searchHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		q := strings.TrimSpace(r.URL.Query().Get("q"))

		var contents []Content
		var people []Person

		if q != "" {
			like := "%" + q + "%"

			// Search contents by title, summary, author_name, or person name
			rows, err := db.Query(`
				SELECT c.id, c.title, c.summary, c.thumbnail_url, c.source_url,
				       c.source_platform, c.author_name, c.person_id,
				       COALESCE(p.name, c.author_name) AS person_name,
				       c.difficulty, c.duration, c.category_id,
				       COALESCE(cat.name, ''), c.sort_order
				FROM contents c
				LEFT JOIN categories cat ON c.category_id = cat.id
				LEFT JOIN people p ON c.person_id = p.id
				WHERE c.title LIKE ? OR c.summary LIKE ? OR c.author_name LIKE ? OR p.name LIKE ?
				ORDER BY c.sort_order`, like, like, like, like)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			for rows.Next() {
				var ct Content
				if err := rows.Scan(&ct.ID, &ct.Title, &ct.Summary, &ct.ThumbnailURL,
					&ct.SourceURL, &ct.SourcePlatform, &ct.AuthorName, &ct.PersonID,
					&ct.PersonName, &ct.Difficulty, &ct.Duration,
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

			// Search people by name or bio
			pRows, err := db.Query(`
				SELECT p.id, p.slug, p.name, p.bio, p.platforms_json,
				       (SELECT COUNT(*) FROM contents WHERE person_id = p.id) AS content_count
				FROM people p
				WHERE p.name LIKE ? OR p.bio LIKE ?
				ORDER BY p.name`, like, like)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer pRows.Close()

			for pRows.Next() {
				var p Person
				if err := pRows.Scan(&p.ID, &p.Slug, &p.Name, &p.Bio, &p.PlatformsJSON, &p.ContentCount); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				people = append(people, p)
			}
		}

		data := struct {
			Query    string
			Contents []Content
			People   []Person
		}{q, contents, people}

		tmpl, err := parseTemplate("search.html")
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("Search: " + q + "\n"))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, data)
	}
}

package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

type Category struct {
	ID        int
	Name      string
	Icon      string
	SortOrder int
	ParentID  sql.NullInt64
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

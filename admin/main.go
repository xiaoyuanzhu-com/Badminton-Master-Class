package main

import (
	"crypto/subtle"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

func initDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func basicAuth(next http.Handler, username, password string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(u), []byte(username)) != 1 ||
			subtle.ConstantTimeCompare([]byte(p), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="BMC Admin"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func setupRoutes(db *sql.DB, dbPath string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/categories", http.StatusSeeOther)
	})
	mux.HandleFunc("/categories", categoriesHandler(db))
	mux.HandleFunc("/categories/", categoryActionHandler(db))
	mux.HandleFunc("/contents", contentsHandler(db))
	mux.HandleFunc("/contents/", contentActionHandler(db))
	mux.HandleFunc("/export", exportHandler(db, dbPath))
	return mux
}

func main() {
	dbPath := "bmc.db"
	db, err := initDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	username := getEnv("BMC_ADMIN_USER", "admin")
	password := getEnv("BMC_ADMIN_PASSWORD", "admin")

	mux := setupRoutes(db, dbPath)
	handler := basicAuth(mux, username, password)

	fmt.Println("羽球大师课 Admin panel running on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

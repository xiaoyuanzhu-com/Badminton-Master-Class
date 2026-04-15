package main

import (
	"crypto/subtle"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

func initDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	if err := migrateDB(db); err != nil {
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
			w.Header().Set("WWW-Authenticate", `Basic realm="BMC"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func setupRoutes(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler(db))
	mux.HandleFunc("/categories", categoriesHandler(db))
	mux.HandleFunc("/contents", contentsHandler(db))
	mux.HandleFunc("/contents/", contentDetailHandler(db))
	mux.HandleFunc("/paths", pathsListHandler(db))
	mux.HandleFunc("/paths/", pathDetailHandler(db))
	mux.HandleFunc("/people", peopleListHandler(db))
	mux.HandleFunc("/people/", personDetailHandler(db))
	mux.HandleFunc("/search", searchHandler(db))
	return mux
}

func main() {
	dbPath := getEnv("BMC_DB_PATH", "bmc.db")
	db, err := initDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := setupRoutes(db)

	var handler http.Handler = mux
	authEnabled := strings.ToLower(getEnv("BMC_AUTH_ENABLED", "false"))
	if authEnabled == "true" || authEnabled == "1" {
		username := getEnv("BMC_ADMIN_USER", "admin")
		password := getEnv("BMC_ADMIN_PASSWORD", "admin")
		handler = basicAuth(mux, username, password)
	}

	addr := getEnv("BMC_ADDR", ":8080")
	fmt.Printf("羽球大师课 running on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"net/http"

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

func main() {
	dbPath := "bmc.db"
	db, err := initDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/categories", http.StatusSeeOther)
	})
	http.HandleFunc("/categories", categoriesHandler(db))
	http.HandleFunc("/categories/", categoryActionHandler(db))
	http.HandleFunc("/contents", contentsHandler(db))
	http.HandleFunc("/contents/", contentActionHandler(db))
	http.HandleFunc("/export", exportHandler(db, dbPath))

	fmt.Println("羽球大师课 Admin panel running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

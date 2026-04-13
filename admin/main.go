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
	db, err := initDB("yuqiupu.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Admin panel running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

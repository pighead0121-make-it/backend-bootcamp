package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var db *sql.DB

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello Backend")
	})

	http.HandleFunc("/health", healthHandler)

	http.HandleFunc("/info", infoHandler)

	http.HandleFunc("/books", booksHandler)

	http.HandleFunc("/books/", bookIDHandler)

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	name := os.Getenv("DB_NAME")
	password := os.Getenv("DB_PASSWORD")

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		user,
		password,
		host,
		port,
		name,
	)

	var err error
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		fmt.Println("failed to open database:", err)
		return
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Println("failed to connect to database:", err)
		return
	}

	fmt.Println("connected to database")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

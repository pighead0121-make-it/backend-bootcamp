package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pighead0121-make-it/backend-bootcamp/projects/book-api/config"
	"github.com/pighead0121-make-it/backend-bootcamp/projects/book-api/database"
)

func main() {

	configDB, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	db, err := database.Connect(configDB)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	log.Println("connected to database")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello Backend")
	})

	http.HandleFunc("/health", healthHandler)

	http.HandleFunc("/info", infoHandler)

	http.HandleFunc("/books", func(w http.ResponseWriter, r *http.Request) {
		booksHandler(db, w, r)
	})

	http.HandleFunc("/books/", func(w http.ResponseWriter, r *http.Request) {
		bookIDHandler(db, w, r)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

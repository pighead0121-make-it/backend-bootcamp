package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello Backend")
	})

	http.HandleFunc("/health", healthHandler)

	http.HandleFunc("/info", infoHandler)

	http.HandleFunc("/books", booksHandler)

	http.HandleFunc("/books/", bookIDHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	type Book struct {
		ID     int    `json:"id"`
		Title  string `json:"title"`
		Author string `json:"author"`
	}

	var books = []Book{
		{
			ID:     1,
			Title:  "The Go Programming Language",
			Author: "Alan Donovan",
		},
	}

	nextBookID := 2

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello Backend")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		fmt.Fprintln(w, "OK")
	})

	http.HandleFunc("/books", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(books)
			return
		} else if r.Method == http.MethodPost {
			var book Book
			err := json.NewDecoder(r.Body).Decode(&book)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			book.ID = nextBookID
			nextBookID++
			books = append(books, book)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(book)
			return
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
	})

	http.HandleFunc("/books/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/books/"))
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		if id <= 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		for _, book := range books {
			if id == book.ID {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(book)
				return
			}
		}
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

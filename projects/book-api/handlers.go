package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintln(w, "OK")
}

func infoHandler(w http.ResponseWriter, r *http.Request) {

	type info struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	appInfo := info{
		Name:    "compose-book-api",
		Version: "1.0.0",
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(appInfo)
}

func booksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Method == http.MethodGet {
		rows, err := db.Query("SELECT id, title, author FROM books")
		if err != nil {
			fmt.Println("SQL Error:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		var result []Book

		for rows.Next() {
			var book Book

			err := rows.Scan(
				&book.ID,
				&book.Title,
				&book.Author,
			)
			if err != nil {
				fmt.Println("SQL Error:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			result = append(result, book)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)

		return
	}

	if r.Method == http.MethodPost {
		var book Book

		err := json.NewDecoder(r.Body).Decode(&book)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		err = db.QueryRow(
			"INSERT INTO books (title, author) VALUES ($1, $2) RETURNING id",
			book.Title,
			book.Author,
		).Scan(&book.ID)

		if err != nil {
			fmt.Println("SQL Error:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(book)
		return
	}
}

func bookIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPut && r.Method != http.MethodDelete {
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

	if r.Method == http.MethodGet {
		var book Book

		err = db.QueryRow(
			"SELECT id, title, author FROM books WHERE id = $1",
			id,
		).Scan(
			&book.ID,
			&book.Title,
			&book.Author,
		)

		if err == sql.ErrNoRows {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		if err != nil {
			fmt.Println("SQL Error:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(book)
		return
	}

	if r.Method == http.MethodPut {
		var book Book

		err := json.NewDecoder(r.Body).Decode(&book)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		err = db.QueryRow(
			"UPDATE books SET title = $1, author = $2 WHERE id = $3 RETURNING id, title, author",
			book.Title,
			book.Author,
			id,
		).Scan(
			&book.ID,
			&book.Title,
			&book.Author,
		)

		if err == sql.ErrNoRows {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		if err != nil {
			fmt.Println("SQL error:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(book)
		return
	}

	if r.Method == http.MethodDelete {
		var deletedID int

		err := db.QueryRow(
			"DELETE FROM books WHERE id = $1 RETURNING id",
			id,
		).Scan(&deletedID)

		if err == sql.ErrNoRows {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		if err != nil {
			fmt.Println("SQL Error:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}
}

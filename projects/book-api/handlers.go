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
	switch r.Method {
	case http.MethodGet:
		author := r.URL.Query().Get("author")
		limitString := r.URL.Query().Get("limit")
		sort := r.URL.Query().Get("sort")

		var rows *sql.Rows
		var err error
		var limit int

		query := "SELECT id, title, author FROM books"
		args := []any{}

		if author != "" {
			args = append(args, author)
			query += fmt.Sprintf(" WHERE author = $%d", len(args))
		}

		switch sort {
		case "asc":
			query += " ORDER BY id ASC"
		case "desc":
			query += " ORDER BY id DESC"
		case "":

		default:
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		if limitString != "" {
			limit, err = strconv.Atoi(limitString)
			if err != nil || limit <= 0 {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			args = append(args, limit)
			query += fmt.Sprintf(" LIMIT $%d", len(args))
		}

		rows, err = db.Query(query, args...)

		if err != nil {
			fmt.Println("SQL Error:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		result := []Book{}

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
		if err := rows.Err(); err != nil {
			fmt.Println("SQL Error:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return

	case http.MethodPost:
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

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

func bookIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/books/"))

	if err != nil || id <= 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
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

	case http.MethodPut:
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

	case http.MethodDelete:
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

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

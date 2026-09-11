package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func writeJSONErr(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONErr(w, "Method Not Allowed", http.StatusMethodNotAllowed)
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
		writeJSONErr(w, "Method Not Allowed", http.StatusMethodNotAllowed)
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

		query := "SELECT b.id, b.title, a.name AS author FROM books b JOIN authors a ON b.author_id = a.id"
		args := []any{}

		if author != "" {
			args = append(args, author)
			query += fmt.Sprintf(" WHERE a.name = $%d", len(args))
		}

		switch sort {
		case "asc":
			query += " ORDER BY b.id ASC"
		case "desc":
			query += " ORDER BY b.id DESC"
		case "":

		default:
			writeJSONErr(w, "Bad Request", http.StatusBadRequest)
			return
		}

		if limitString != "" {
			limit, err = strconv.Atoi(limitString)
			if err != nil || limit <= 0 {
				writeJSONErr(w, "Bad Request", http.StatusBadRequest)
				return
			}
			args = append(args, limit)
			query += fmt.Sprintf(" LIMIT $%d", len(args))
		}

		rows, err = db.Query(query, args...)

		if err != nil {
			fmt.Println("SQL Error:", err)
			writeJSONErr(w, "Internal Server Error", http.StatusInternalServerError)
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
				writeJSONErr(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			result = append(result, book)
		}
		if err := rows.Err(); err != nil {
			fmt.Println("SQL Error:", err)
			writeJSONErr(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return

	case http.MethodPost:
		var book Book
		var authorID int

		err := json.NewDecoder(r.Body).Decode(&book)
		if err != nil {
			writeJSONErr(w, "Bad Request", http.StatusBadRequest)
			return
		}

		book.Title = strings.TrimSpace(book.Title)
		book.Author = strings.TrimSpace(book.Author)

		if book.Title == "" {
			writeJSONErr(w, "Title is required", http.StatusBadRequest)
			return
		}
		if book.Author == "" {
			writeJSONErr(w, "Author is required", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			fmt.Println("SQL Error:", err)
			writeJSONErr(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		err = tx.QueryRow(
			"SELECT id FROM authors WHERE name = $1",
			book.Author,
		).Scan(&authorID)

		if err == sql.ErrNoRows {
			err = tx.QueryRow(
				"INSERT INTO authors (name) VALUES ($1) RETURNING id",
				book.Author,
			).Scan(&authorID)

			if err != nil {
				fmt.Println("SQL Error:", err)
				writeJSONErr(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		if err != nil {
			fmt.Println("SQL Error:", err)
			writeJSONErr(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = tx.QueryRow(
			"INSERT INTO books (title, author_id) VALUES ($1, $2) RETURNING id",
			book.Title,
			authorID,
		).Scan(&book.ID)

		if err != nil {
			fmt.Println("SQL Error:", err)
			writeJSONErr(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = tx.Commit()
		if err != nil {
			fmt.Println("SQL Error:", err)
			writeJSONErr(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(book)
		return

	default:
		writeJSONErr(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

func bookIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/books/"))

	if err != nil || id <= 0 {
		writeJSONErr(w, "Bad Request", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		var book Book

		err = db.QueryRow(
			"SELECT b.id, b.title, a.name AS author FROM books b JOIN authors a ON b.author_id = a.id WHERE b.id = $1",
			id,
		).Scan(
			&book.ID,
			&book.Title,
			&book.Author,
		)

		if err == sql.ErrNoRows {
			writeJSONErr(w, "Not Found", http.StatusNotFound)
			return
		}

		if err != nil {
			fmt.Println("SQL Error:", err)
			writeJSONErr(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(book)
		return

	case http.MethodPut:
		var book Book
		var authorID int

		err = json.NewDecoder(r.Body).Decode(&book)
		if err != nil {
			writeJSONErr(w, "Bad Request", http.StatusBadRequest)
			return
		}

		book.Title = strings.TrimSpace(book.Title)
		book.Author = strings.TrimSpace(book.Author)

		if book.Title == "" {
			writeJSONErr(w, "Title is required", http.StatusBadRequest)
			return
		}
		if book.Author == "" {
			writeJSONErr(w, "Author is required", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			fmt.Println("SQL Error:", err)
			writeJSONErr(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		err = tx.QueryRow(
			"SELECT id FROM authors WHERE name = $1",
			book.Author,
		).Scan(&authorID)

		if err == sql.ErrNoRows {
			err = tx.QueryRow(
				"INSERT INTO authors (name) VALUES ($1) RETURNING id",
				book.Author,
			).Scan(&authorID)

			if err != nil {
				fmt.Println("SQL Error:", err)
				writeJSONErr(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		if err != nil {
			fmt.Println("SQL Error:", err)
			writeJSONErr(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = tx.QueryRow(
			"UPDATE books SET title = $1, author_id = $2 WHERE id = $3 RETURNING id",
			book.Title,
			authorID,
			id,
		).Scan(
			&book.ID,
		)

		if err == sql.ErrNoRows {
			writeJSONErr(w, "Not Found", http.StatusNotFound)
			return
		}

		if err != nil {
			fmt.Println("SQL Error:", err)
			writeJSONErr(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = tx.Commit()
		if err != nil {
			fmt.Println("SQL Error:", err)
			writeJSONErr(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(book)
		return

	case http.MethodDelete:
		var deletedID int

		err = db.QueryRow(
			"DELETE FROM books WHERE id = $1 RETURNING id",
			id,
		).Scan(&deletedID)

		if err == sql.ErrNoRows {
			writeJSONErr(w, "Not Found", http.StatusNotFound)
			return
		}

		if err != nil {
			fmt.Println("SQL Error:", err)
			writeJSONErr(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return

	default:
		writeJSONErr(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

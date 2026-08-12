package main

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

var nextBookID int = 2

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Book struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

var (
	books  = []Book{}
	nextID = 1
	mu     sync.Mutex
)

func main() {
	http.HandleFunc("/books", booksHandler)
	http.HandleFunc("/books/", bookHandler) // для get/put/delete по ID

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// GET /books, POST /books
func booksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(books)

	case http.MethodPost:
		var book Book
		if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
			http.Error(w, "Invalid body", http.StatusBadRequest)
			return
		}
		mu.Lock()
		book.ID = nextID
		nextID++
		books = append(books, book)
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(book)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET/PUT/DELETE /books/{id}
func bookHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/books/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	switch r.Method {
	case http.MethodGet:
		for _, b := range books {
			if b.ID == id {
				json.NewEncoder(w).Encode(b)
				return
			}
		}
		http.NotFound(w, r)

	case http.MethodPut:
		var updated Book
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			http.Error(w, "Invalid body", http.StatusBadRequest)
			return
		}
		for i, b := range books {
			if b.ID == id {
				books[i].Title = updated.Title
				json.NewEncoder(w).Encode(books[i])
				return
			}
		}
		http.NotFound(w, r)

	case http.MethodDelete:
		for i, b := range books {
			if b.ID == id {
				books = append(books[:i], books[i+1:]...)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		http.NotFound(w, r)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Person struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type PersonUpdate struct {
	Name *string `json:"name"`
	Age  *int    `json:"age"`
}

var (
	persons = []*Person{}
	nextId  = 1
	mu      sync.Mutex
)

func main() {

	http.HandleFunc("/persons", PersonsHandler)
	http.HandleFunc("/persons/", PersonHandler)

	fmt.Println("Server running on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))

}

func PersonsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(persons)

	case http.MethodPost:
		var person Person
		if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		person.Id = nextId
		nextId++
		persons = append(persons, &person)
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(person)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func PersonHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/persons/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	switch r.Method {
	case http.MethodGet:
		for _, person := range persons {
			if person.Id == id {
				json.NewEncoder(w).Encode(person)
				return
			}
		}

	case http.MethodPatch:
		var updated PersonUpdate
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for _, p := range persons {
			if p.Id == id {
				if updated.Name != nil {
					p.Name = *updated.Name
				}
				if updated.Age != nil {
					p.Age = *updated.Age
				}
				json.NewEncoder(w).Encode(p)
				return
			}
		}
		http.NotFound(w, r)

	case http.MethodDelete:
		for i, person := range persons {
			if person.Id == id {
				persons = append(persons[:i], persons[i+1:]...)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		http.NotFound(w, r)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

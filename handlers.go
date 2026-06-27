package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "UP"}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// middleware to log all incoming requests
func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Incoming Request: %s %s", r.Method, r.URL.Path)
		ctx := r.Context()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if len(tasks) == 0 {
		w.Write([]byte("No Tasks Found!"))
		return
	}

	json.NewEncoder(w).Encode(tasks)
}

func postTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newTask Task

	if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newTask.ID = nextID
	nextID++
	newTask.CreatedAt = time.Now()

	tasks = append(tasks, newTask)


	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTask)
}

package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
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

func (h *TaskServer) getTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.tasks) == 0 {
		w.Write([]byte("[]"))
		return
	}

	json.NewEncoder(w).Encode(h.tasks)
}

func (h *TaskServer) getTasksById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, `{"error": "Invalid task ID"}`, http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	for _, task := range h.tasks {
		if task.ID == id {
			json.NewEncoder(w).Encode(task)
			return
		}
	}

	http.Error(w, `"error":"Task not found"`, http.StatusNotFound)
}

func (h *TaskServer) postTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newTask Task

	if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
		if errors.Is(err, io.EOF) {
			http.Error(w, `{"error": "Empty request body"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error": "Malformed JSON syntax"}`, http.StatusBadRequest)
		return
	}

	if newTask.Title == "" {
		http.Error(w, `{"error": "Missing required field: title"}`, http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	newTask.ID = h.nextID
	h.nextID++
	newTask.CreatedAt = time.Now()

	h.tasks = append(h.tasks, newTask)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTask)
}

func (h *TaskServer) updateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, `{"error": "Invalid task ID"}`, http.StatusBadRequest)
		return
	}

	var updatedInput Task
	if err := json.NewDecoder(r.Body).Decode(&updatedInput); err != nil {
		http.Error(w, `{"error": "Malformed JSON syntax"}`, http.StatusBadRequest)
		return
	}

	if updatedInput.Title == "" {
		http.Error(w, `{"error": "Title field cannot be empty"}`, http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	for i, task := range h.tasks {
		if task.ID == id {
			// Update properties but keep the original ID and Creation date
			h.tasks[i].Title = updatedInput.Title
			h.tasks[i].Completed = updatedInput.Completed

			json.NewEncoder(w).Encode(h.tasks[i])
			return
		}
	}

	http.Error(w, `{"error": "Task not found"}`, http.StatusNotFound)
}

func (h *TaskServer) deleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, `{"error": "Invalid task ID"}`, http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	for i, task := range h.tasks {
		if task.ID == id {
			h.tasks = append(h.tasks[:i], h.tasks[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	http.Error(w, `{"error": "Task not found"}`, http.StatusNotFound)
}

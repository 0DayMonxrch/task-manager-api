package main

import (
	"database/sql"
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

	rows, err := h.db.Query("SELECT id, title, completed, created_at FROM tasks")
	if err != nil {
		http.Error(w, `{"error": "Failed to query tasks"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt); err != nil {
			http.Error(w, `{"error": "Failed to scan tasks"}`, http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, t)
	}

	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskServer) getTasksById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, `{"error": "Invalid task ID"}`, http.StatusBadRequest)
		return
	}

	var t Task
	err = h.db.QueryRow("SELECT id, title, completed, created_at FROM tasks WHERE id = ?", id).
		Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, `{"error": "Task not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(t)
}

func (h *TaskServer) postTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newTask Task
	err := json.NewDecoder(r.Body).Decode(&newTask)
	if err != nil {
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

	newTask.CreatedAt = time.Now()

	result, err := h.db.Exec("INSERT INTO tasks (title, completed, created_at) VALUES (?, ?, ?)",
		newTask.Title, newTask.Completed, newTask.CreatedAt)
	if err != nil {
		http.Error(w, `{"error": "Failed to save task"}`, http.StatusInternalServerError)
		return
	}

	lastID, err := result.LastInsertId()
	if err == nil {
		newTask.ID = int(lastID)
	}

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

	result, err := h.db.Exec("UPDATE tasks SET title = ?, completed = ? WHERE id = ?",
		updatedInput.Title, updatedInput.Completed, id)
	if err != nil {
		http.Error(w, `{"error": "Failed to update task"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, `{"error": "Task not found"}`, http.StatusNotFound)
		return
	}

	// Fetch the updated resource to echo back
	h.db.QueryRow("SELECT id, title, completed, created_at FROM tasks WHERE id = ?", id).
		Scan(&updatedInput.ID, &updatedInput.Title, &updatedInput.Completed, &updatedInput.CreatedAt)

	json.NewEncoder(w).Encode(updatedInput)
}

func (h *TaskServer) deleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, `{"error": "Invalid task ID"}`, http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		http.Error(w, `{"error": "Failed to delete task"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, `{"error": "Task not found"}`, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

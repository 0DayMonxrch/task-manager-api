package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/0DayMonxrch/task-manager-api/internal/repository"
)

type TaskHandler struct {
	repo *repository.TaskRepository
}

func NewTaskHandler(repo *repository.TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

func (h *TaskHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "UP"}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// middleware to log all incoming requests
func (h *TaskHandler) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Incoming Request: %s %s", r.Method, r.URL.Path)
		ctx := r.Context()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tasks, err := h.repo.GetAll()
	if err != nil {
		http.Error(w, `{"error": "Failed to query tasks"}`, http.StatusInternalServerError)
		return
	}

	if len(tasks) == 0 {
		w.Write([]byte("[]"))
		return
	}

	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) GetTaskById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, `{"error": "Invalid task ID"}`, http.StatusBadRequest)
		return
	}

	t, err := h.repo.GetByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		http.Error(w, `{"error": "Task not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(t)
}

func (h *TaskHandler) PostTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var input struct {
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		if errors.Is(err, io.EOF) {
			http.Error(w, `{"error": "Empty request body"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error": "Malformed JSON syntax"}`, http.StatusBadRequest)
		return
	}

	if input.Title == "" {
		http.Error(w, `{"error": "Missing required field: title"}`, http.StatusBadRequest)
		return
	}

	newTask, err := h.repo.Create(input.Title, input.Completed)
	if err != nil {
		http.Error(w, `{"error": "Failed to save task"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTask)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, `{"error": "Invalid task ID"}`, http.StatusBadRequest)
		return
	}

	var input struct {
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "Malformed JSON syntax"}`, http.StatusBadRequest)
		return
	}

	if input.Title == "" {
		http.Error(w, `{"error": "Title field cannot be empty"}`, http.StatusBadRequest)
		return
	}

	updatedTask, err := h.repo.Update(id, input.Title, input.Completed)
	if errors.Is(err, repository.ErrNotFound) {
		http.Error(w, `{"error": "Task not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"error": "Failed to update task"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updatedTask)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, `{"error": "Invalid task ID"}`, http.StatusBadRequest)
		return
	}

	err = h.repo.Delete(id)
	if errors.Is(err, repository.ErrNotFound) {
		http.Error(w, `{"error": "Task not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"error": "Failed to delete task"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

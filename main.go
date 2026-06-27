package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	server := &TaskServer{
		tasks:  []Task{},
		nextID: 1,
	}

	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /tasks", server.getTasks)
	mux.HandleFunc("GET /tasks/{id}", server.getTasksById)
	mux.HandleFunc("POST /tasks", server.postTasks)
	mux.HandleFunc("PUT /tasks/{id}", server.updateTask)
	mux.HandleFunc("DELETE /tasks/{id}", server.deleteTask)

	loggedMux := logRequest(mux)
	log.Println("Server running on :8080")

	if err := http.ListenAndServe(":8080", loggedMux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /tasks", getTasks)
	mux.HandleFunc("POST /tasks", postTasks)

	loggedMux := logRequest(mux)
	log.Println("Server running on :8080")

	if err := http.ListenAndServe(":8080", loggedMux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

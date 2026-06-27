package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	mux := http.NewServeMux()

	manager := &TaskServer{
		tasks:  []Task{},
		nextID: 1,
	}

	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /tasks", manager.getTasks)
	mux.HandleFunc("GET /tasks/{id}", manager.getTasksById)
	mux.HandleFunc("POST /tasks", manager.postTasks)
	mux.HandleFunc("PUT /tasks/{id}", manager.updateTask)
	mux.HandleFunc("DELETE /tasks/{id}", manager.deleteTask)

	loggedMux := logRequest(mux)
	
	srv := &http.Server{
		Addr:    ":8080",
		Handler: loggedMux,
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Server starting on :8080...")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed down: %v", err)
		}
	}()

	sig := <-shutdownChan
	log.Printf("Received signal %v. Shutting down gracefully...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server cleanly stopped.")
}

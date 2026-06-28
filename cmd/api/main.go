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

	"github.com/0DayMonxrch/task-manager-api/internal/database"
	"github.com/0DayMonxrch/task-manager-api/internal/handlers"
	"github.com/0DayMonxrch/task-manager-api/internal/repository"
)

func main() {
	db, err := database.InitDB("./tasks.db")
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.Close()

	repo := repository.NewTaskRepository(db)
	taskHandler := handlers.NewTaskHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", taskHandler.HandleHealth)
	mux.HandleFunc("GET /tasks", taskHandler.GetTasks)
	mux.HandleFunc("GET /tasks/{id}", taskHandler.GetTaskById)
	mux.HandleFunc("POST /tasks", taskHandler.PostTasks)
	mux.HandleFunc("PUT /tasks/{id}", taskHandler.UpdateTask)
	mux.HandleFunc("DELETE /tasks/{id}", taskHandler.DeleteTask)

	loggedMux := taskHandler.LogRequest(mux)

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

	log.Println("Closing database connections...")
	if err := db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Server cleanly stopped.")
}

package main

import (
	"sync"
	"time"
)

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"createdAt"`
}

type TaskServer struct {
	mu     sync.Mutex
	tasks  []Task
	nextID int
}

package repository

import (
	"database/sql"
	"errors"
	"time"
)

var ErrNotFound = errors.New("task not found")

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"createdAt"`
}

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) GetAll() ([]Task, error) {
	rows, err := r.db.Query("SELECT id, title, completed, created_at FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *TaskRepository) GetByID(id int) (Task, error) {
	var t Task
	err := r.db.QueryRow("SELECT id, title, completed, created_at FROM tasks WHERE id = ?", id).
		Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return t, ErrNotFound
	}
	return t, err
}

func (r *TaskRepository) Create(title string, completed bool) (Task, error) {
	var t Task
	t.Title = title
	t.Completed = completed
	t.CreatedAt = time.Now()

	result, err := r.db.Exec("INSERT INTO tasks (title, completed, created_at) VALUES (?, ?, ?)",
		t.Title, t.Completed, t.CreatedAt)
	if err != nil {
		return t, err
	}

	lastID, err := result.LastInsertId()
	if err == nil {
		t.ID = int(lastID)
	}
	return t, nil
}

func (r *TaskRepository) Update(id int, title string, completed bool) (Task, error) {
	var t Task
	result, err := r.db.Exec("UPDATE tasks SET title = ?, completed = ? WHERE id = ?", title, completed, id)
	if err != nil {
		return t, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return t, ErrNotFound
	}

	// Fetch the updated resource to pass back cleanly
	return r.GetByID(id)
}

func (r *TaskRepository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

package models

import "time"

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

func NewTask(id int, title string) Task {
	return Task{
		ID:        id,
		Title:     title,
		Done:      false,
		CreatedAt: time.Now().UTC(), // просто время, без Format
	}
}

package dao

import "time"

// Task represents a task in the system
type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"startTime"`
	ParentID  *string   `json:"parentId,omitempty"`
	SubTasks  []Task    `json:"subTasks,omitempty"`
}

// Repository defines the interface for task storage
type Repository interface {
	GetTasks() ([]Task, error)
	GetTaskByID(id string) (*Task, error)
	CreateTask(task *Task) error
	UpdateTask(task *Task) error
	DeleteTask(id string) error
	AddSubTask(parentID string, task *Task) error
}

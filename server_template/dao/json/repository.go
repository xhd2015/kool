package json

import (
	"sync"
)

type Repository struct {
	filename string
	mu       sync.RWMutex
}

func New(filename string) *Repository {
	return &Repository{
		filename: filename,
	}
}

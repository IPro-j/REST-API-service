package storage

import (
	"errors"
	"sync"

	"tasks-api/internal/models"
)

var (
	ErrNotFound = errors.New("task not found")
	ErrInvalid  = errors.New("invalid task data")
)

type MemoryStorage struct {
	mu     sync.RWMutex
	tasks  map[int]models.Task
	nextID int
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		tasks:  make(map[int]models.Task),
		nextID: 1,
	}
}

func (s *MemoryStorage) List() []models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		result = append(result, t)
	}
	return result
}

func (s *MemoryStorage) Create(t models.Task) (models.Task, error) {
	if t.Title == "" {
		return models.Task{}, ErrInvalid
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	t.ID = s.nextID
	s.nextID++

	s.tasks[t.ID] = t
	return t, nil
}

func (s *MemoryStorage) Get(id int) (models.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	return task, ok
}

func (s *MemoryStorage) Update(id int, t models.Task) (models.Task, error) {
	if t.Title == "" {
		return models.Task{}, ErrInvalid
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.tasks[id]
	if !ok {
		return models.Task{}, ErrNotFound
	}

	t.ID = id
	s.tasks[id] = t
	return t, nil
}

func (s *MemoryStorage) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.tasks[id]
	if !ok {
		return ErrNotFound
	}

	delete(s.tasks, id)
	return nil
}

package storage

import "tasks-api/internal/models"

type Storage interface {
	List() []models.Task
	Create(models.Task) (models.Task, error)
	Get(id int) (models.Task, bool)
	Update(id int, t models.Task) (models.Task, error)
	Delete(id int) error
}

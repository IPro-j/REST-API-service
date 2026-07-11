package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"tasks-api/internal/models"
	"tasks-api/internal/storage"
)

type Handler struct {
	Store storage.Storage
}

func New(s storage.Storage) *Handler {
	return &Handler{Store: s}
}

// writeJSON пишет JSON-ответ
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("error encoding JSON: %v", err)
	}
}

// writeError ошибка в едином формате
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// logRequest логирует метод и путь
func logRequest(r *http.Request) {
	log.Printf("[%s] %s", r.Method, r.URL.Path)
}

// TasksCollection обработка запрсов GET /tasks и POST /tasks
func (h *Handler) TasksCollection(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	switch r.Method {
	case http.MethodGet:
		tasks := h.Store.List()
		writeJSON(w, http.StatusOK, tasks)

	case http.MethodPost:
		var t models.Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil || t.Title == "" {
			writeError(w, http.StatusBadRequest, "invalid or missing 'title' field")
			return
		}

		// Создаём задачу с датой
		t = models.NewTask(0, t.Title)
		created, err := h.Store.Create(t)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, created)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// TaskItem обрабатка GET/PUT/DELETE /tasks/{id}
func (h *Handler) TaskItem(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	id, err := strconv.Atoi(path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		task, ok := h.Store.Get(id)
		if !ok {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeJSON(w, http.StatusOK, task)

	case http.MethodPut:
		var t models.Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil || t.Title == "" {
			writeError(w, http.StatusBadRequest, "invalid or missing 'title' field")
			return
		}

		updated, err := h.Store.Update(id, t)
		if err != nil {
			if err == storage.ErrNotFound {
				writeError(w, http.StatusNotFound, "task not found")
			} else {
				writeError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		writeJSON(w, http.StatusOK, updated)

	case http.MethodDelete:
		err := h.Store.Delete(id)
		if err != nil {
			if err == storage.ErrNotFound {
				writeError(w, http.StatusNotFound, "task not found")
			} else {
				writeError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		// 204 No Content, тело не возвращаем
		w.WriteHeader(http.StatusNoContent)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

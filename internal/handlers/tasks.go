package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
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
	// 1. Сначала сериализуем в буфер. Если тут ошибка — мы ещё не отправили статус.
	body, err := json.Marshal(data)
	if err != nil {
		log.Printf("error marshaling JSON: %v", err)

		// Отправляем честный 500, потому что не смогли сформировать корректный ответ
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "internal server error",
		})
		return
	}

	// 2. Только теперь устанавливаем заголовок и статус
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, _ = w.Write(body)
}

// writeError ошибка в едином формате
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// mapStorageError преобразует ошибки хранилища в HTTP-статус и понятное сообщение
func mapStorageError(err error) (int, string) {
	if err == nil {
		return 0, ""
	}

	// Сначала проверяем специфические ошибки хранилища
	if errors.Is(err, storage.ErrNotFound) {
		return http.StatusNotFound, "resource not found"
	}
	if errors.Is(err, storage.ErrInvalid) {
		// Пробрасываем сообщение от хранилища: оно должно быть понятным клиенту
		return http.StatusBadRequest, err.Error()
	}

	// Всё остальное считаем внутренней ошибкой сервера
	log.Printf("unexpected storage error: %v", err)
	return http.StatusInternalServerError, "internal server error"
}

// TasksCollection обработка запрсов GET /tasks и POST /tasks
func (h *Handler) TasksCollection(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		tasks := h.Store.List()
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].ID < tasks[j].ID
		})
		writeJSON(w, http.StatusOK, tasks)

	case http.MethodPost:
		var t models.Task
		const maxTitleLength = 200

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		if err := dec.Decode(&t); err != nil {
			var syntaxErr *json.SyntaxError
			if errors.As(err, &syntaxErr) {
				writeError(w, http.StatusBadRequest, "invalid JSON syntax")
				return
			}
			writeError(w, http.StatusBadRequest, err.Error()) // сюда попадут и unknown fields
			return
		}

		// Проверка: после объекта ничего не должно быть
		var extra interface{}
		if err := dec.Decode(&extra); err != io.EOF {
			writeError(w, http.StatusBadRequest, "unexpected extra data after JSON object")
			return
		}

		t.Title = strings.TrimSpace(t.Title)
		if t.Title == "" {
			writeError(w, http.StatusBadRequest, "missing or whitespace-only 'title' field")
			return
		}

		if len(t.Title) > maxTitleLength {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("'title' exceeds maximum length of %d characters", maxTitleLength))
			return
		}

		// Создаём задачу с датой
		//t = models.NewTask(0, t.Title, t.Done)
		created, err := h.Store.Create(t)
		if err != nil {
			status, msg := mapStorageError(err) // ✅ маппинг ошибок
			writeError(w, status, msg)          // 400/404/500 по ситуации
			return
		}

		writeJSON(w, http.StatusCreated, created)

	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// TaskItem обрабатка GET/PUT/DELETE /tasks/{id}
func (h *Handler) TaskItem(w http.ResponseWriter, r *http.Request) {
	//logRequest(r)
	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	id, err := strconv.Atoi(path)

	if id <= 0 {
		writeError(w, http.StatusBadRequest, "ID must be a positive integer")
		return
	}

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
		const maxTitleLength = 200

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		if err := dec.Decode(&t); err != nil {
			var syntaxErr *json.SyntaxError
			if errors.As(err, &syntaxErr) {
				writeError(w, http.StatusBadRequest, "invalid JSON syntax")
				return
			}
			writeError(w, http.StatusBadRequest, err.Error()) // сюда попадут и unknown fields
			return
		}

		// Проверка: после объекта ничего не должно быть
		var extra interface{}
		if err := dec.Decode(&extra); err != io.EOF {
			writeError(w, http.StatusBadRequest, "unexpected extra data after JSON object")
			return
		}

		t.Title = strings.TrimSpace(t.Title)
		if t.Title == "" {
			writeError(w, http.StatusBadRequest, "missing or whitespace-only 'title' field")
			return
		}

		if len(t.Title) > maxTitleLength {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("'title' exceeds maximum length of %d characters", maxTitleLength))
			return
		}

		updated, err := h.Store.Update(id, t)
		if err != nil {
			status, msg := mapStorageError(err)
			writeError(w, status, msg)
			return
		}
		writeJSON(w, http.StatusOK, updated)

	case http.MethodDelete:
		err := h.Store.Delete(id)
		if err != nil {
			status, msg := mapStorageError(err)
			writeError(w, status, msg)
			return
		}
		// 204 No Content, тело не возвращаем
		w.WriteHeader(http.StatusNoContent)

	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

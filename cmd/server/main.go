package main

import (
	"log"
	"net/http"

	"tasks-api/internal/handlers"
	"tasks-api/internal/middleware"
	"tasks-api/internal/storage"
)

func main() {
	store := storage.NewMemoryStorage()
	h := handlers.New(store)

	mux := http.NewServeMux()

	// Важно: порядок регистрации имеет значение
	// /tasks/ должен идти ПОСЛЕ /tasks, иначе /tasks будет перекрыт префиксом
	mux.HandleFunc("/tasks", h.TasksCollection)
	mux.HandleFunc("/tasks/", h.TaskItem)
	mux.HandleFunc("/health", h.Health)

	// Сначала NotFoundJSONHandler (чтобы перехватить 404), потом LoggingMiddleware (чтобы логировать всё)
	handler := middleware.LoggingMiddleware(middleware.NotFoundJSONHandler(mux))

	log.Println("server listening on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}

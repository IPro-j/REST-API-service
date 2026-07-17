package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
)

// LoggingMiddleware логирует каждый входящий запрос
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// NotFoundJSONHandler перехватывает 404 от ServeMux и превращает его в JSON.
func NotFoundJSONHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Используем recorder, чтобы «посмотреть» ответ, который отдал next (обычно это ServeMux)
		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r)

		// Если статус 404 И Content-Type не JSON — значит, это стандартный 404 от net/http
		if rec.Code == http.StatusNotFound {
			ct := rec.Header().Get("Content-Type")
			if !strings.HasPrefix(ct, "application/json") {
				// Подменяем ответ на наш JSON
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
				return
			}
		}

		// Иначе копируем оригинальный ответ как есть
		for k, vs := range rec.Header() {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(rec.Code)
		_, _ = w.Write(rec.Body.Bytes())
	})
}

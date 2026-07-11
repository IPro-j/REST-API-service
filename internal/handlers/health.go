package handlers

import (
	"net/http"
)

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	//writeJSON из tasks.go
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

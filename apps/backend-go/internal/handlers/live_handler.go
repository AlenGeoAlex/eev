package handlers

import (
	"backend-go/internal"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
)

type LiveHandler struct {
	upgrader *websocket.Upgrader
}

func (h *LiveHandler) upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return h.upgrader.Upgrade(w, r, nil)
}

func (h *LiveHandler) Live(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrade(w, r)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to upgrade connection")
		return
	}

	defer conn.Close()
	for {

	}
}

func (h *LiveHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *LiveHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, internal.ErrorResponse{
		Message: message,
	})
}

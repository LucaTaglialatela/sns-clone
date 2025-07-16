package api

import (
	"io"
	"log"
	"net/http"
)

// PingHandler interface
type PingHandler struct{}

// NewPingHandler creates a new instance of PingHandler.
func NewPingHandler() *PingHandler {
	return &PingHandler{}
}

func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	_, err := io.WriteString(w, "Pong!\n")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

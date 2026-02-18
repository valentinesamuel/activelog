package websocket

import (
	"log"
	"net/http"

	gorillaws "github.com/gorilla/websocket"
	requestcontext "github.com/valentinesamuel/activelog/internal/requestContext"
	"github.com/valentinesamuel/activelog/pkg/response"
)

var upgrader = gorillaws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now; tighten in production
		return true
	},
}

// Handler holds the hub reference for serving WebSocket connections
type Handler struct {
	hub *Hub
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// ServeWS upgrades an HTTP connection to WebSocket and registers the client
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	user, ok := requestcontext.FromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	client := newClient(h.hub, conn, user.Id)
	h.hub.register <- client

	go client.writePump()
	go client.readPump()
}

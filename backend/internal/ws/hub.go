package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents one websocket connection for a user.
type Client struct {
	UserID string
	Conn   WSConn
	mu     sync.Mutex
}

// WSConn captures the websocket connection operations used by this package.
type WSConn interface {
	Close() error
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
}

func (c *Client) WriteJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	payload, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.Conn.WriteMessage(websocket.TextMessage, payload)
}

// Hub owns active websocket clients and handles concurrent register/unregister safely.
// The clients map is only mutated from the Run loop through channels.
type Hub struct {
	register      chan *Client
	unregister    chan *Client
	deliverToUser chan userEvent
	clients       map[string]map[*Client]struct{}
}

type userEvent struct {
	userID  string
	message OutgoingEnvelope
}

func NewHub() *Hub {
	return &Hub{
		register:      make(chan *Client, 128),
		unregister:    make(chan *Client, 128),
		deliverToUser: make(chan userEvent, 512),
		clients:       make(map[string]map[*Client]struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.addClient(client)
		case client := <-h.unregister:
			h.removeClient(client)
		case evt := <-h.deliverToUser:
			h.deliverUserEvent(evt)
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) EmitToUser(userID string, message OutgoingEnvelope) {
	if userID == "" {
		return
	}
	h.deliverToUser <- userEvent{userID: userID, message: message}
}

func (h *Hub) EmitToUsers(userIDs []string, message OutgoingEnvelope) {
	seen := make(map[string]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if userID == "" {
			continue
		}
		if _, exists := seen[userID]; exists {
			continue
		}
		seen[userID] = struct{}{}
		h.EmitToUser(userID, message)
	}
}

func (h *Hub) addClient(client *Client) {
	if client == nil || client.UserID == "" {
		return
	}

	if _, ok := h.clients[client.UserID]; !ok {
		h.clients[client.UserID] = make(map[*Client]struct{})
	}
	h.clients[client.UserID][client] = struct{}{}
	log.Printf("ws connected user=%s active_connections=%d", client.UserID, len(h.clients[client.UserID]))
}

func (h *Hub) removeClient(client *Client) {
	if client == nil || client.UserID == "" {
		return
	}

	userClients, ok := h.clients[client.UserID]
	if !ok {
		return
	}

	delete(userClients, client)
	if len(userClients) == 0 {
		delete(h.clients, client.UserID)
		log.Printf("ws disconnected user=%s active_connections=0", client.UserID)
		return
	}

	log.Printf("ws disconnected user=%s active_connections=%d", client.UserID, len(userClients))
}
func (h *Hub) deliverUserEvent(evt userEvent) {
	userClients, ok := h.clients[evt.userID]
	if !ok || len(userClients) == 0 {
		return
	}

	for client := range userClients {
		if err := client.WriteJSON(evt.message); err != nil {
			log.Printf("ws deliver failed user=%s err=%v", evt.userID, err)
			_ = client.Conn.Close()
			delete(userClients, client)
		}
	}

	if len(userClients) == 0 {
		delete(h.clients, evt.userID)
	}
}

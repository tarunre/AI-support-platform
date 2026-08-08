package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/ayush/supportiq/internal/utils"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 50 * time.Second
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Origin check is handled via WEBSOCKET_ORIGIN config in the handler
		return true
	},
}

// Client represents a single WebSocket connection.
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   uint
	tenantID uuid.UUID
}

// Hub manages all active WebSocket clients and broadcasts events to them.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates and returns a new Hub. Call Run() in a goroutine to start it.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 512),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run processes client registration, unregistration, and broadcasts.
// Must be called in a dedicated goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			utils.Logger.WithField("user_id", client.userID).
				WithField("connected", h.ConnectedCount()).
				Info("WS: Client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			utils.Logger.WithField("user_id", client.userID).Info("WS: Client disconnected")

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Slow client — unregister to prevent memory leaks
					go func(c *Client) { h.unregister <- c }(client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends an arbitrary payload to all connected clients in the same tenant.
// Use BroadcastToTenant for tenant-scoped events. This method broadcasts to ALL
// clients and should only be used for system-wide events (e.g. maintenance).
func (h *Hub) Broadcast(payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		utils.Logger.WithError(err).Warn("WS: Failed to marshal broadcast payload")
		return
	}
	select {
	case h.broadcast <- data:
	default:
		utils.Logger.Warn("WS: Broadcast channel full, dropping message")
	}
}

// BroadcastToTenant sends a payload only to clients belonging to a specific tenant.
func (h *Hub) BroadcastToTenant(tenantID uuid.UUID, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		utils.Logger.WithError(err).Warn("WS: Failed to marshal tenant broadcast payload")
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.tenantID != tenantID {
			continue
		}
		select {
		case client.send <- data:
		default:
			go func(c *Client) { h.unregister <- c }(client)
		}
	}
}

// BroadcastRaw sends pre-encoded JSON bytes to all connected clients.
func (h *Hub) BroadcastRaw(data []byte) {
	select {
	case h.broadcast <- data:
	default:
		utils.Logger.Warn("WS: Broadcast channel full, dropping message")
	}
}

// BroadcastToTenantRaw sends pre-encoded JSON bytes only to clients in the given tenant.
func (h *Hub) BroadcastToTenantRaw(tenantID uuid.UUID, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.tenantID != tenantID {
			continue
		}
		select {
		case client.send <- data:
		default:
			go func(c *Client) { h.unregister <- c }(client)
		}
	}
}

// ConnectedCount returns the number of currently connected WebSocket clients.
func (h *Hub) ConnectedCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// ServeWS upgrades an HTTP request to a WebSocket connection and registers the client.
// userID and tenantID must be validated by the calling handler before invoking this.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, userID uint, tenantID uuid.UUID) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.Logger.WithError(err).Warn("WS: Upgrade failed")
		return
	}

	client := &Client{
		hub:      h,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		tenantID: tenantID,
	}

	h.register <- client

	go client.writePump()
	go client.readPump()
}

// ─── Client I/O pumps ────────────────────────────────────────────────────────

// readPump reads from the WebSocket connection. It closes the connection on
// error or when the client sends a close frame.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				utils.Logger.WithError(err).Warn("WS: Unexpected close")
			}
			break
		}
	}
}

// writePump drains the send channel and writes messages to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

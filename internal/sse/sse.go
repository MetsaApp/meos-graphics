package sse

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"meos-graphics/internal/logger"
)

// Client represents a connected SSE client
type Client struct {
	ID      string
	Channel chan Event
	Closed  bool
	mu      sync.RWMutex
}

// Event represents an SSE event
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Hub manages SSE clients
type Hub struct {
	clients    map[string]*Client
	broadcast  chan Event
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates a new SSE hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan Event, 100),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			logger.InfoLogger.Printf("SSE client registered: %s", client.ID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				client.mu.Lock()
				client.Closed = true
				close(client.Channel)
				client.mu.Unlock()
				delete(h.clients, client.ID)
				logger.InfoLogger.Printf("SSE client unregistered: %s", client.ID)
			}
			h.mu.Unlock()

		case event := <-h.broadcast:
			h.mu.RLock()
			for id, client := range h.clients {
				client.mu.RLock()
				if client.Closed {
					client.mu.RUnlock()
					continue
				}
				client.mu.RUnlock()

				select {
				case client.Channel <- event:
				default:
					// Client's channel is full, skip this event
					logger.DebugLogger.Printf("SSE client %s channel full, skipping event", id)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastUpdate sends an update event to all connected clients
func (h *Hub) BroadcastUpdate(eventType string, data interface{}) {
	event := Event{
		Type: eventType,
		Data: data,
	}

	select {
	case h.broadcast <- event:
	default:
		logger.ErrorLogger.Println("SSE broadcast channel full, dropping event")
	}
}

// HandleSSE handles SSE connections
func (h *Hub) HandleSSE(c *gin.Context) {
	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Create client
	clientID := fmt.Sprintf("%d", time.Now().UnixNano())
	client := &Client{
		ID:      clientID,
		Channel: make(chan Event, 10),
	}

	// Register client
	h.register <- client

	// Ensure cleanup on disconnect
	defer func() {
		h.unregister <- client
	}()

	// Send initial connection event
	c.SSEvent("connected", gin.H{"id": clientID})
	c.Writer.Flush()

	// Listen for client disconnect
	notify := c.Writer.CloseNotify()

	for {
		select {
		case <-notify:
			// Client disconnected
			return

		case event, ok := <-client.Channel:
			if !ok {
				// Channel closed
				return
			}

			// Send event to client
			data, err := json.Marshal(event.Data)
			if err != nil {
				logger.ErrorLogger.Printf("Failed to marshal SSE event data: %v", err)
				continue
			}

			c.SSEvent(event.Type, string(data))
			c.Writer.Flush()

		case <-time.After(30 * time.Second):
			// Send heartbeat to keep connection alive
			c.SSEvent("heartbeat", gin.H{"time": time.Now().Unix()})
			c.Writer.Flush()
		}
	}
}

// GetConnectedClients returns the number of connected clients
func (h *Hub) GetConnectedClients() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

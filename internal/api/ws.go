package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// BroadcastEvent implements the indexer.EventBroadcaster interface.
func (h *WSHub) BroadcastEvent(eventType string, data interface{}) {
	h.Broadcast(WSEvent{Type: eventType, Data: data})
}

// WSEvent represents a WebSocket event sent to clients.
type WSEvent struct {
	Type      string      `json:"type"`      // "node_status", "batch_update", "new_batch"
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// WSHub manages WebSocket connections and broadcasts events.
type WSHub struct {
	clients    map[*wsClient]bool
	mu         sync.RWMutex
	broadcast  chan []byte
}

type wsClient struct {
	conn *websocket.Conn
	send chan []byte
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func NewWSHub() *WSHub {
	h := &WSHub{
		clients:   make(map[*wsClient]bool),
		broadcast: make(chan []byte, 256),
	}
	go h.run()
	return h
}

func (h *WSHub) run() {
	for msg := range h.broadcast {
		h.mu.RLock()
		for c := range h.clients {
			select {
			case c.send <- msg:
			default:
				// Client too slow, drop
				close(c.send)
				delete(h.clients, c)
			}
		}
		h.mu.RUnlock()
	}
}

// Broadcast sends an event to all connected WebSocket clients.
func (h *WSHub) Broadcast(event WSEvent) {
	event.Timestamp = time.Now().Unix()
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	select {
	case h.broadcast <- data:
	default:
		// Channel full, skip
	}
}

// ClientCount returns the number of connected clients.
func (h *WSHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *WSHub) addClient(c *wsClient) {
	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()
	log.Printf("[ws] client connected (%d total)", h.ClientCount())
}

func (h *WSHub) removeClient(c *wsClient) {
	h.mu.Lock()
	if _, ok := h.clients[c]; ok {
		close(c.send)
		delete(h.clients, c)
	}
	h.mu.Unlock()
	log.Printf("[ws] client disconnected (%d total)", h.ClientCount())
}

// handleWS handles WebSocket upgrade and message routing.
func (s *Server) handleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[ws] upgrade error: %v", err)
		return
	}

	client := &wsClient{
		conn: conn,
		send: make(chan []byte, 64),
	}
	s.wsHub.addClient(client)

	// Writer goroutine
	go func() {
		defer conn.Close()
		for msg := range client.send {
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}()

	// Reader goroutine (keeps connection alive via ping/pong)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Ping ticker
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	// Read messages (just to detect disconnect)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
	s.wsHub.removeClient(client)
}

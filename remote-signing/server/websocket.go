package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocketHub manages WebSocket connections and message broadcasting
type WebSocketHub struct {
	clients    map[string]*WebSocketClient
	sseClients map[string]map[string]chan SSEEvent // uuid -> clientID -> channel
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	broadcast  chan WebSocketMessage
	sseEvents  chan SSEEvent
	mutex      sync.RWMutex
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[string]*WebSocketClient),
		sseClients: make(map[string]map[string]chan SSEEvent),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		broadcast:  make(chan WebSocketMessage, 256),
		sseEvents:  make(chan SSEEvent, 256),
	}
}

// Run starts the WebSocket hub
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client.ID] = client
			h.mutex.Unlock()
			log.Printf("WebSocket client connected: %s", client.ID)

			// Send welcome message
			select {
			case client.Send <- WebSocketMessage{
				Type: MessageTypeStatus,
				Payload: map[string]string{
					"status": "connected",
					"client_id": client.ID,
				},
			}:
			default:
				close(client.Send)
				h.mutex.Lock()
				delete(h.clients, client.ID)
				h.mutex.Unlock()
			}

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
				log.Printf("WebSocket client disconnected: %s", client.ID)
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			for _, client := range h.clients {
				// Check if client is interested in this UUID
				if message.UUID != "" {
					if !client.UUIDs[message.UUID] {
						continue
					}
				}

				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client.ID)
				}
			}
			h.mutex.RUnlock()

		case event := <-h.sseEvents:
			h.mutex.RLock()
			if clients, ok := h.sseClients[event.UUID]; ok {
				for clientID, clientChan := range clients {
					select {
					case clientChan <- event:
						// Event sent successfully
					default:
						// Client channel is full or closed, remove it
						delete(clients, clientID)
						close(clientChan)
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastToUUID sends a message to all clients subscribed to a specific UUID
func (h *WebSocketHub) BroadcastToUUID(uuid string, message WebSocketMessage) {
	message.UUID = uuid
	h.broadcast <- message
}

// BroadcastSSEToUUID sends an SSE event to all clients subscribed to a specific UUID
func (h *WebSocketHub) BroadcastSSEToUUID(uuid string, event SSEEvent) {
	event.UUID = uuid
	h.sseEvents <- event
}

// RegisterSSEClient registers an SSE client for a specific UUID
func (h *WebSocketHub) RegisterSSEClient(uuid, clientID string, clientChan chan SSEEvent) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.sseClients[uuid] == nil {
		h.sseClients[uuid] = make(map[string]chan SSEEvent)
	}
	h.sseClients[uuid][clientID] = clientChan
	log.Printf("SSE client registered: %s for UUID %s", clientID, uuid)
}

// UnregisterSSEClient unregisters an SSE client
func (h *WebSocketHub) UnregisterSSEClient(uuid, clientID string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if clients, ok := h.sseClients[uuid]; ok {
		if clientChan, exists := clients[clientID]; exists {
			close(clientChan)
			delete(clients, clientID)
			log.Printf("SSE client unregistered: %s for UUID %s", clientID, uuid)
		}

		// Clean up empty UUID entry
		if len(clients) == 0 {
			delete(h.sseClients, uuid)
		}
	}
}

// BroadcastToClient sends a message to a specific client
func (h *WebSocketHub) BroadcastToClient(clientID string, message WebSocketMessage) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if client, ok := h.clients[clientID]; ok {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(h.clients, clientID)
		}
	}
}

// GetClientCount returns the number of connected clients
func (h *WebSocketHub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// Upgrader configures the WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking based on config
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// HandleWebSocket handles WebSocket connections
// @Summary WebSocket connection
// @Description WebSocket endpoint for real-time notifications about signing requests
// @Tags WebSocket
// @Success 101 "WebSocket connection established"
// @Router /ws [get]
func (s *Server) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	clientID := uuid.New().String()
	client := &WebSocketClient{
		ID:       clientID,
		Conn:     conn,
		Send:     make(chan WebSocketMessage, 256),
		ClientID: c.Query("client_id"),
		UUIDs:    make(map[string]bool),
	}

	s.hub.register <- client

	// Start goroutines for reading and writing
	go s.writeWebSocket(client)
	go s.readWebSocket(client)
}

// readWebSocket handles reading messages from WebSocket client
func (s *Server) readWebSocket(client *WebSocketClient) {
	defer func() {
		s.hub.unregister <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(1024 * 1024) // 1MB limit
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message WebSocketMessage
		err := client.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle different message types
		switch message.Type {
		case MessageTypeSubscribe:
			// Subscribe client to UUID updates
			if message.UUID != "" {
				client.UUIDs[message.UUID] = true
				log.Printf("Client %s subscribed to UUID %s", client.ID, message.UUID)
			}
		case MessageTypeHeartbeat:
			// Respond to heartbeat
			client.Send <- WebSocketMessage{
				Type: MessageTypeHeartbeat,
				Payload: map[string]string{
					"status": "alive",
				},
			}
		}
	}
}

// writeWebSocket handles writing messages to WebSocket client
func (s *Server) writeWebSocket(client *WebSocketClient) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			messageBytes, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling WebSocket message: %v", err)
				return
			}

			w.Write(messageBytes)
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

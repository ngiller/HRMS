package service

import (
	"encoding/json"
	"log"
	"sync"
)

// SSEEvent represents an event to be sent to SSE clients
type SSEEvent struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

// SSEClient represents a single SSE connection
type SSEClient struct {
	UserID string
	Events chan []byte
}

// SSEHub manages all SSE client connections
type SSEHub struct {
	mu      sync.RWMutex
	clients map[string]map[*SSEClient]struct{} // userID -> set of clients
}

// NewSSEHub creates a new SSEHub
func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[string]map[*SSEClient]struct{}),
	}
}

// Subscribe registers a new SSE client for a user
func (h *SSEHub) Subscribe(userID string) *SSEClient {
	h.mu.Lock()
	defer h.mu.Unlock()

	client := &SSEClient{
		UserID: userID,
		Events: make(chan []byte, 32),
	}

	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*SSEClient]struct{})
	}
	h.clients[userID][client] = struct{}{}

	log.Printf("[SSE] Client subscribed: user=%s, total=%d", userID, h.totalClients())
	return client
}

// Unsubscribe removes a client
func (h *SSEHub) Unsubscribe(client *SSEClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.clients[client.UserID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.clients, client.UserID)
		}
	}
	log.Printf("[SSE] Client unsubscribed: user=%s", client.UserID)
}

// BroadcastToUser sends an event to all connections for a specific user
func (h *SSEHub) BroadcastToUser(userID string, event SSEEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[SSE] Failed to marshal event: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.clients[userID]; ok {
		for client := range clients {
			select {
			case client.Events <- data:
			default:
				// Client buffer full, skip
				log.Printf("[SSE] Client buffer full, dropping event for user=%s", userID)
			}
		}
	}
}

// BroadcastToUsers sends an event to all connections for multiple users
func (h *SSEHub) BroadcastToUsers(userIDs []string, event SSEEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[SSE] Failed to marshal event: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, uid := range userIDs {
		if clients, ok := h.clients[uid]; ok {
			for client := range clients {
				select {
				case client.Events <- data:
				default:
					log.Printf("[SSE] Client buffer full, dropping event for user=%s", uid)
				}
			}
		}
	}
}

// BroadcastAll sends an event to all connected clients
func (h *SSEHub) BroadcastAll(event SSEEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[SSE] Failed to marshal event: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, clients := range h.clients {
		for client := range clients {
			select {
			case client.Events <- data:
			default:
			}
		}
	}
}

func (h *SSEHub) totalClients() int {
	count := 0
	for _, clients := range h.clients {
		count += len(clients)
	}
	return count
}

// ─── Global Singleton ──────────────────────────────────────────

var GlobalSSEHub *SSEHub

// InitGlobalSSEHub initializes the global SSE hub singleton
func InitGlobalSSEHub() *SSEHub {
	hub := NewSSEHub()
	GlobalSSEHub = hub
	log.Println("[SSE] Hub initialized")
	return hub
}

// GetSSEHub returns the global SSE hub instance
func GetSSEHub() *SSEHub {
	return GlobalSSEHub
}

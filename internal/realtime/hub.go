package realtime

import (
	"context"
	"sync"
)

type Client interface {
	Send(ctx context.Context, message []byte) error
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]Client
}

func NewHub() *Hub { return &Hub{clients: make(map[string]Client)} }

func (h *Hub) Register(id string, client Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[id] = client
}

func (h *Hub) Unregister(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, id)
}

func (h *Hub) Broadcast(ctx context.Context, message []byte) error {
	h.mu.RLock()
	clients := make([]Client, 0, len(h.clients))
	for _, client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		if err := client.Send(ctx, message); err != nil {
			return err
		}
	}
	return nil
}

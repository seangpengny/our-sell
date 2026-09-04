package realtime

// WebSocket transport is intentionally kept separate from the hub. A future
// Fiber WebSocket adapter can translate frames into service calls and use the
// same business services as REST handlers without putting Fiber in services.
type WebSocket struct {
	hub *Hub
}

func NewWebSocket(hub *Hub) *WebSocket { return &WebSocket{hub: hub} }

func (w *WebSocket) Hub() *Hub { return w.hub }

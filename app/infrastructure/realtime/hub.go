// Package realtime holds the in-memory WebSocket connection registry used to
// push chat messages to online recipients. It's a single-process hub — fine
// for the current single-instance deployment (no Redis/pub-sub involved);
// horizontal scaling would need a shared broker instead, but that's not a
// concern today.
package realtime

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Envelope is the JSON shape of every message pushed over a socket, e.g.
// {"type":"message.new","payload":{...entities.Message}}.
type Envelope struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

// outboxSize bounds how many unsent envelopes queue up for a slow/stuck
// client before SendToUser starts dropping them — a push must never block
// the HTTP request that triggered it.
const outboxSize = 16

// Client wraps one live WebSocket connection. Every write goes through
// outbox and the single writer goroutine started by the hub, since gorilla
// forbids concurrent WriteMessage calls on the same connection.
type Client struct {
	conn   *websocket.Conn
	outbox chan []byte
}

type Hub struct {
	mu    sync.RWMutex
	conns map[uuid.UUID]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{conns: make(map[uuid.UUID]map[*Client]struct{})}
}

// Register starts tracking conn for userID and spins up its writer
// goroutine. Callers must call Unregister (typically once the read loop
// that detects disconnects returns) to release it.
func (h *Hub) Register(userID uuid.UUID, conn *websocket.Conn) *Client {
	client := &Client{conn: conn, outbox: make(chan []byte, outboxSize)}

	h.mu.Lock()
	if h.conns[userID] == nil {
		h.conns[userID] = make(map[*Client]struct{})
	}
	h.conns[userID][client] = struct{}{}
	h.mu.Unlock()

	go client.writeLoop()
	return client
}

func (h *Hub) Unregister(userID uuid.UUID, client *Client) {
	h.mu.Lock()
	if clients, ok := h.conns[userID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.conns, userID)
		}
	}
	h.mu.Unlock()

	close(client.outbox)
	_ = client.conn.Close()
}

// SendToUser pushes envelope to every connection userID currently has open.
// A user with no open connections is a silent no-op — they'll see the
// message on their next REST fetch, no offline delivery in v1.
func (h *Hub) SendToUser(userID uuid.UUID, envelope Envelope) {
	data, err := json.Marshal(envelope)
	if err != nil {
		log.Println("realtime: failed to marshal envelope:", err)
		return
	}

	h.mu.RLock()
	clients := h.conns[userID]
	targets := make([]*Client, 0, len(clients))
	for c := range clients {
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	for _, c := range targets {
		select {
		case c.outbox <- data:
		default:
			// Outbox full — a stuck/slow client shouldn't block delivery to
			// everyone else, so this push is dropped for them.
			log.Println("realtime: dropping envelope, outbox full for user", userID)
		}
	}
}

func (c *Client) writeLoop() {
	for data := range c.outbox {
		if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}
}

package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/archerny/mini-agent/internal/protocol"
	"github.com/archerny/mini-agent/internal/runtime"
)

const (
	// Server sends ping every 30 seconds.
	pingInterval = 30 * time.Second

	// If no pong is received within this time, the connection is dead.
	pongTimeout = 40 * time.Second

	// Max time to wait for a write to complete.
	writeTimeout = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	// Allow all origins in development (MVP).
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ---------------------------------------------------------------------------
// Hub — manages all active WebSocket clients
// ---------------------------------------------------------------------------

// Hub manages WebSocket client connections and broadcasts events to them.
type Hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}
	engine  *runtime.Engine
}

// NewHub creates a new WebSocket hub and subscribes to the event stream.
func NewHub(engine *runtime.Engine) *Hub {
	h := &Hub{
		clients: make(map[*client]struct{}),
		engine:  engine,
	}

	// Subscribe to the event stream — fan out to all WebSocket clients.
	engine.EventStream.Subscribe(func(evt *protocol.Event) {
		h.broadcast(evt)
	})

	return h
}

// broadcast sends an event to all connected clients.
func (h *Hub) broadcast(evt *protocol.Event) {
	data, err := json.Marshal(evt)
	if err != nil {
		log.Printf("[ws-hub] failed to marshal event: %v", err)
		return
	}

	h.mu.RLock()
	clients := make([]*client, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, c := range clients {
		c.send(data)
	}
}

// register adds a client to the hub.
func (h *Hub) register(c *client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
	log.Printf("[ws-hub] client connected (total: %d)", h.clientCount())
}

// unregister removes a client from the hub.
func (h *Hub) unregister(c *client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	log.Printf("[ws-hub] client disconnected (total: %d)", h.clientCount())
}

func (h *Hub) clientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// ServeWS handles the WebSocket upgrade and manages the connection.
// Route: GET /ws/events
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ws-hub] upgrade failed: %v", err)
		return
	}

	c := newClient(conn, h)
	h.register(c)

	// Start read and write pumps.
	go c.readPump()
	go c.writePump()
}

// ---------------------------------------------------------------------------
// client — a single WebSocket connection
// ---------------------------------------------------------------------------

type client struct {
	conn    *websocket.Conn
	hub     *Hub
	sendCh  chan []byte
	closeCh chan struct{}
	once    sync.Once
}

func newClient(conn *websocket.Conn, hub *Hub) *client {
	return &client{
		conn:    conn,
		hub:     hub,
		sendCh:  make(chan []byte, 256),
		closeCh: make(chan struct{}),
	}
}

// send queues a message for sending. Non-blocking; drops if buffer full.
func (c *client) send(data []byte) {
	select {
	case c.sendCh <- data:
	default:
		// Client is too slow, drop the message.
		log.Printf("[ws-client] send buffer full, dropping message")
	}
}

// close shuts down the client (idempotent).
func (c *client) close() {
	c.once.Do(func() {
		close(c.closeCh)
		c.hub.unregister(c)
		_ = c.conn.Close()
	})
}

// readPump reads messages from the WebSocket (mainly for pong handling).
// The WebSocket is server→client (unidirectional push), so we mostly
// just read pong responses and detect disconnections.
func (c *client) readPump() {
	defer c.close()

	c.conn.SetReadLimit(512)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongTimeout))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[ws-client] read error: %v", err)
			}
			return
		}
	}
}

// writePump writes messages and pings to the WebSocket.
func (c *client) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.close()
	}()

	for {
		select {
		case <-c.closeCh:
			return

		case data, ok := <-c.sendCh:
			if !ok {
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

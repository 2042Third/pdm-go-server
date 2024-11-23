package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

// Upgrade HTTP to WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins
}

// Message represents the JSON structure of incoming messages
type Message struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
}

// Connection represents a single WebSocket connection
type Connection struct {
	ws     *websocket.Conn
	send   chan []byte
	userID string
}

// Hub manages active WebSocket connections
type Hub struct {
	connections map[*Connection]bool
	register    chan *Connection
	unregister  chan *Connection
	mu          sync.Mutex
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		connections: make(map[*Connection]bool),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
	}
}

// Start the Hub to handle register/unregister
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.connections[conn] = true
			h.mu.Unlock()
			//log.Printf("Client connected: %s", conn.userID)

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.connections[conn]; ok {
				delete(h.connections, conn)
				close(conn.send)
				//log.Printf("Client disconnected: %s", conn.userID)
			}
			h.mu.Unlock()
		}
	}
}

// ReadPump reads messages from the WebSocket
func (c *Connection) ReadPump(h *Hub) {
	defer func() {
		h.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(512)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
				//log.Printf("Client closed connection: %v", err)
			} else {
				log.Printf("Error reading message: %v", err)
			}
			return
		}

		// Process the message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			c.send <- []byte("Invalid message format")
			continue
		}
		response := fmt.Sprintf("%s: %d", msg.Msg, msg.Code)
		c.send <- []byte(response)
	}
}

// WritePump writes messages to the WebSocket
func (c *Connection) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Channel closed
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			//log.Printf("%s sending message: %s", c.userID, message)
			if err := c.ws.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error writing message: %v", err)
				return
			}
		case <-ticker.C:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWS handles WebSocket requests from clients
func ServeWS(h *Hub, w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	conn := &Connection{ws: ws, send: make(chan []byte, 256), userID: r.RemoteAddr}
	h.register <- conn

	go conn.WritePump()
	conn.ReadPump(h)
}

// Main function to start the WebSocket server
func main() {
	hub := NewHub()
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWS(hub, w, r)
	})

	port := "8082"
	log.Printf("WebSocket server started on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

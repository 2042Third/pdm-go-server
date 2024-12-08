package handlers

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

// ClientConnection represents a WebSocket connection with its associated metadata
type ClientConnection struct {
	conn     *websocket.Conn
	userID   string
	queue    amqp.Queue // Client's specific queue
	consumer string     // Consumer tag for cleanup
}

// WebSocketHandler manages WebSocket connections and RabbitMQ integration
type WebSocketHandler struct {
	upgrader   websocket.Upgrader
	clients    map[string][]*ClientConnection
	clientsMu  sync.RWMutex
	rabbitChan *amqp.Channel
	exchange   string
}

func NewWebSocketHandler(rabbitChan *amqp.Channel, exchange string) (*WebSocketHandler, error) {
	handler := &WebSocketHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients:    make(map[string][]*ClientConnection),
		rabbitChan: rabbitChan,
		exchange:   exchange,
	}

	// Declare the topic exchange
	err := rabbitChan.ExchangeDeclare(
		exchange,
		"topic",
		true,  // Durable
		false, // Auto-deleted
		false, // Internal
		false, // No-wait
		nil,
	)
	if err != nil {
		return nil, err
	}

	return handler, nil
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "Missing userId", http.StatusBadRequest)
		return
	}

	// First set up the WebSocket connection
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Immediately set up ping/pong handlers BEFORE any read/write operations
	conn.SetPingHandler(func(string) error {
		return conn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(10*time.Second))
	})

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start ping ticker early
	go func() {
		ticker := time.NewTicker(54 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
					log.Printf("Error sending ping: %v", err)
					return
				}
			}
		}
	}()

	// Create an exclusive queue for this specific client
	queue, err := h.rabbitChan.QueueDeclare(
		"",    // Let RabbitMQ generate a unique name
		false, // Non-durable
		true,  // Delete when unused
		true,  // Exclusive
		false, // No-wait
		nil,
	)
	if err != nil {
		log.Printf("Failed to declare queue: %v", err)
		conn.Close()
		return
	}

	// Bind the queue to the exchange with user-specific routing key
	routingKey := "note_update." + userID
	err = h.rabbitChan.QueueBind(
		queue.Name,
		routingKey,
		h.exchange,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed to bind queue: %v", err)
		conn.Close()
		return
	}

	// Start consuming messages
	msgs, err := h.rabbitChan.Consume(
		queue.Name,
		"",    // Let RabbitMQ generate a consumer tag
		true,  // Auto-ack
		true,  // Exclusive
		false, // No-local
		false, // No-wait
		nil,
	)
	if err != nil {
		log.Printf("Failed to start consuming: %v", err)
		conn.Close()
		return
	}

	client := &ClientConnection{
		conn:   conn,
		userID: userID,
		queue:  queue,
	}

	// Register client
	h.registerClient(client)
	defer h.unregisterClient(client)

	// Handle RabbitMQ messages in a goroutine
	go func() {
		for msg := range msgs {
			// Send directly to this client's WebSocket
			if err := conn.WriteMessage(websocket.TextMessage, msg.Body); err != nil {
				log.Printf("Error writing to WebSocket: %v", err)
				return
			}
		}
	}()

	// Handle WebSocket messages
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading from WebSocket: %v", err)
			break
		}

		// Echo back the message
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Printf("Error writing to WebSocket: %v", err)
			break
		}
	}

}

func (h *WebSocketHandler) registerClient(client *ClientConnection) {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	h.clients[client.userID] = append(h.clients[client.userID], client)
	log.Printf("New client registered for user %s. Total clients: %d",
		client.userID, len(h.clients[client.userID]))
}

func (h *WebSocketHandler) unregisterClient(client *ClientConnection) {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	userClients := h.clients[client.userID]
	for i, c := range userClients {
		if c == client {
			h.clients[client.userID] = append(userClients[:i], userClients[i+1:]...)
			break
		}
	}

	client.conn.Close()

	if len(h.clients[client.userID]) == 0 {
		delete(h.clients, client.userID)
	}
}

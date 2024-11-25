package main

import (
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"syncing/config"
	"syncing/handlers"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	// Initialize PostgreSQL connection
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer config.CloseDB(db)

	// Initialize RabbitMQ connection
	rabbitMQ, ch, err := config.InitRabbitMQ()
	if err != nil {
		log.Fatalf("Error initializing RabbitMQ: %v", err)
	}
	defer rabbitMQ.Close()
	defer ch.Close()

	syncHandler := handlers.NewSyncHandler(db)

	// Start RabbitMQ consumer goroutine
	go syncHandler.ConsumeRabbitMQMessages(ch)

	// Create WebSocket handler
	wsHandler, err := handlers.NewWebSocketHandler(ch, "notes_exchange")
	if err != nil {
		log.Fatal(err)
	}

	// Set up HTTP route
	http.HandleFunc("/ws", wsHandler.HandleWebSocket)

	// Start server
	log.Fatal(http.ListenAndServe(":8082", nil))
}

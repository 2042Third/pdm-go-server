package main

import (
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"syncing/config"
	"syncing/handlers"

	"github.com/gorilla/mux"
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

	// Set up WebSocket routes
	r := mux.NewRouter()
	r.HandleFunc("/ws", handlers.HandleWebSocket)

	// Start HTTP server
	log.Println("Starting server on :8082...")
	if err := http.ListenAndServe(":8082", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

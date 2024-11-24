package handlers

import (
	"log"

	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

type Task struct {
	ID      uint   `gorm:"primaryKey"`
	Message string `gorm:"type:text;not null"`
}

func ConsumeRabbitMQMessages(ch *amqp.Channel, db *gorm.DB) {
	msgs, err := ch.Consume(
		"task_queue", // Queue name
		"",           // Consumer tag
		true,         // Auto-acknowledge
		false,        // Exclusive
		false,        // No-local
		false,        // No-wait
		nil,          // Arguments
	)
	if err != nil {
		log.Fatalf("Failed to start consuming messages: %v", err)
	}

	// Ensure the database has the Task table
	db.AutoMigrate(&Task{})

	log.Println("Waiting for messages...")

	for msg := range msgs {
		log.Printf("Received a message: %s", msg.Body)

		// Save the message to the database
		task := Task{Message: string(msg.Body)}
		if err := db.Create(&task).Error; err != nil {
			log.Printf("Failed to save message to database: %v", err)
		} else {
			log.Printf("Message saved to database with ID: %d", task.ID)
		}
	}
}

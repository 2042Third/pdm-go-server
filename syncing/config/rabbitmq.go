package config

import (
	"log"

	"github.com/streadway/amqp"
)

func InitRabbitMQ() (*amqp.Connection, *amqp.Channel, error) {
	// Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://admin:admin@localhost:5672/")
	if err != nil {
		return nil, nil, err
	}

	// Open a channel
	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	// Declare a queue
	_, err = ch.QueueDeclare(
		"task_queue", // Queue name
		true,         // Durable
		false,        // Delete when unused
		false,        // Exclusive
		false,        // No-wait
		nil,          // Arguments
	)
	if err != nil {
		return nil, nil, err
	}

	log.Println("Connected to RabbitMQ successfully!")
	return conn, ch, nil
}

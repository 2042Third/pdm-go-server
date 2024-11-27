package config

import (
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
)

func InitRabbitMQ() (*amqp.Connection, *amqp.Channel, error) {

	rabbitUser := os.Getenv("RABBITMQ_USER")
	rabbitPass := os.Getenv("RABBITMQ_PASSWORD")
	rabbitHost := os.Getenv("RABBITMQ_HOST")
	rabbitPort := os.Getenv("RABBITMQ_PORT")

	// Connect to RabbitMQ
	rabbitmqUrl := fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitUser, rabbitPass, rabbitHost, rabbitPort)
	conn, err := amqp.Dial(rabbitmqUrl)
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

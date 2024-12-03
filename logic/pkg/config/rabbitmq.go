package config

import (
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
)

var RabbitMQConnection *amqp.Connection
var RabbitMQChannel *amqp.Channel

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

	// Declare the queue
	_, err = ch.QueueDeclare(
		"logic_to_sync", // Queue name
		true,            // Durable
		false,           // Delete when unused
		false,           // Exclusive
		false,           // No-wait
		nil,             // Arguments
	)
	// Declare the queue
	_, err = ch.QueueDeclare(
		"logic_to_sync", // Queue name
		true,            // Durable
		false,           // Delete when unused
		false,           // Exclusive
		false,           // No-wait
		nil,             // Arguments
	)
	if err != nil {
		return nil, nil, err
	}

	log.Println("RabbitMQ initialized for Logic Server")
	return conn, ch, nil
}

func CloseRabbitMQ() {
	if RabbitMQChannel != nil {
		RabbitMQChannel.Close()
	}
	if RabbitMQConnection != nil {
		RabbitMQConnection.Close()
	}
}

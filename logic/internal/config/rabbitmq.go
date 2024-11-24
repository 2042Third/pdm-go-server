package config

import (
	"log"

	"github.com/streadway/amqp"
)

var RabbitMQConnection *amqp.Connection
var RabbitMQChannel *amqp.Channel

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

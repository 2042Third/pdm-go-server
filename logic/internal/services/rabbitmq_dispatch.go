package services

import (
	"encoding/json"
	"log"
	"pdm-go-server/internal/config"

	"github.com/streadway/amqp"
)

type RabbitMQCtx struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

// InitRabbitMQ initializes the RabbitMQ connection and channel
func NewRabbitMQHandler() (*RabbitMQCtx, error) {
	conn, channel, err := config.InitRabbitMQ()
	if err != nil {
		return nil, err
	}
	return &RabbitMQCtx{
		Conn:    conn,
		Channel: channel,
	}, nil
}

// DispatchRabbitMQMessage sends a message to the "logic_to_sync" queue
func (c *RabbitMQCtx) DispatchRabbitMQMessage(taskType string, payload map[string]interface{}) error {
	message := map[string]interface{}{
		"type":    taskType,
		"payload": payload,
	}
	messageBody, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = c.Channel.Publish(
		"",              // Exchange
		"logic_to_sync", // Routing key (queue name)
		false,           // Mandatory
		false,           // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		},
	)
	if err != nil {
		return err
	}

	log.Printf("Message dispatched to RabbitMQ: %s", string(messageBody))
	return nil
}

// DispatchNoteUpdate sends a "note update" task to RabbitMQ
func (c *RabbitMQCtx) DispatchNoteUpdate(noteID string, content string) {
	payload := map[string]interface{}{
		"noteId":  noteID,
		"content": content,
	}

	if err := c.DispatchRabbitMQMessage("note_update", payload); err != nil {
		log.Printf("Failed to dispatch note update: %v", err)
	}
}

// DispatchAddSession sends an "add session" task to RabbitMQ
func (c *RabbitMQCtx) DispatchAddRefresh(userID, refreshKey string) {
	payload := map[string]interface{}{
		"userId":     userID,
		"refreshKey": refreshKey,
	}

	if err := c.DispatchRabbitMQMessage("add_session", payload); err != nil {
		log.Printf("Failed to dispatch add session: %v", err)
	}
}

// DispatchAddSession sends an "add session" task to RabbitMQ
func (c *RabbitMQCtx) DispatchAddSession(userID, sessionKey string, expiration int64) {
	payload := map[string]interface{}{
		"userId":     userID,
		"sessionKey": sessionKey,
		"expiration": float64(expiration),
	}

	if err := c.DispatchRabbitMQMessage("add_session", payload); err != nil {
		log.Printf("Failed to dispatch add session: %v", err)
	}
}

func (c *RabbitMQCtx) DispatchDeleteSession(userID, sessionKey string) {
	payload := map[string]interface{}{
		"userId":     userID,
		"sessionKey": sessionKey,
	}

	if err := c.DispatchRabbitMQMessage("delete_session", payload); err != nil {
		log.Printf("Failed to dispatch delete session: %v", err)
	}
}

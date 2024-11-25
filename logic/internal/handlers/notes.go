package handlers

import (
	"encoding/json"
	"github.com/streadway/amqp"
)

// Publish note update to available clients
func publishNoteUpdate(ch *amqp.Channel, userID string, update interface{}) error {
	msg, err := json.Marshal(update)
	if err != nil {
		return err
	}

	return ch.Publish(
		"notes_exchange",      // exchange
		"note_update."+userID, // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msg,
		},
	)
}

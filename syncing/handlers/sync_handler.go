package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"syncing/models"
	"time"

	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

type SyncHandler struct {
	DB *gorm.DB
}

func NewSyncHandler(db *gorm.DB) *SyncHandler {
	return &SyncHandler{DB: db}
}

func (h *SyncHandler) ConsumeRabbitMQMessages(ch *amqp.Channel) {
	msgs, err := ch.Consume(
		"logic_to_sync", // Queue name
		"",              // Consumer tag
		true,            // Auto-acknowledge
		false,           // Exclusive
		false,           // No-local
		false,           // No-wait
		nil,             // Arguments
	)
	if err != nil {
		log.Fatalf("Failed to start consuming messages: %v", err)
	}

	log.Println("Sync Server listening for RabbitMQ messages...")

	for msg := range msgs {
		var message map[string]interface{}
		if err := json.Unmarshal(msg.Body, &message); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		taskType, _ := message["type"].(string)
		payload, _ := message["payload"].(map[string]interface{})

		switch taskType {
		case "note_update":
			h.handleNoteUpdate(payload)
		case "add_refresh":
			h.handleAddRefreshKey(payload)
		case "add_session":
			h.handleAddSessionKey(payload)
		case "delete_session":
			h.handleInvalidateSessionKey(payload)
		default:
			log.Printf("Unknown task type: %s", taskType)
		}
	}
}

func (h *SyncHandler) handleNoteUpdate(payload map[string]interface{}) {
	noteID, _ := payload["noteId"].(string)
	content, _ := payload["content"].(string)

	log.Printf("Received RabbitMQ for note update for %s\n", noteID)

	var note models.Notes
	if err := h.DB.First(&note, "id = ?", noteID).Error; err != nil {
		log.Printf("Note not found: %v", err)
		return
	}

	note.Content = content
	note.UpdateTime = note.UpdateTime
	if err := h.DB.Save(&note).Error; err != nil {
		log.Printf("Failed to update note: %v", err)
	} else {
		log.Printf("Note updated successfully: %s", noteID)
	}
}

func (h *SyncHandler) handleAddRefreshKey(payload map[string]interface{}) {
	userIDStr, _ := payload["userId"].(string)
	//sessionKey, _ := payload["sessionKey"].(string)
	refreshKey, _ := payload["refreshKey"].(string)

	log.Printf("Received RabbitMQ for refreshKey for %s\n", userIDStr)

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		return
	}

	session := models.RefreshKey{
		UserID:     uint(userID),
		RefreshKey: refreshKey,
	}

	if err := h.DB.Create(&session).Error; err != nil {
		log.Printf("Failed to add session: %v", err)
	} else {
		log.Printf("Session added successfully for user: %d", userID)
	}
}

func (h *SyncHandler) handleAddSessionKey(payload map[string]interface{}) {
	userIDStr, _ := payload["userId"].(string)
	sessionKey, _ := payload["sessionKey"].(string)
	expiration, _ := payload["expiration"].(float64)

	log.Printf("Received RabbitMQ for sessionKey for %s\n", userIDStr)

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		return
	}

	parsedExp := time.Unix(int64(expiration), 0)
	fmt.Printf("Parsed expiration: %v from %d\n", parsedExp, expiration)
	session := models.SessionKey{
		UserID:         uint(userID),
		SessionKey:     sessionKey,
		ExpirationTime: parsedExp,
	}

	if err := h.DB.Create(&session).Error; err != nil {
		log.Printf("Failed to add session: %v", err)
	} else {
		log.Printf("Session added successfully for user: %d", userID)
	}
}

func (h *SyncHandler) handleInvalidateSessionKey(payload map[string]interface{}) {
	userIDStr, _ := payload["userId"].(string)
	sessionKey, _ := payload["sessionKey"].(string)

	log.Printf("Received RabbitMQ for sessionKey invalidation for %s with %s\n", userIDStr, sessionKey)

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		return
	}

	if sessionKey == "" {
		log.Printf("No session key provided for invalidation")
		return
	}

	// Update the valid field to "0" instead of deleting
	if err := h.DB.Model(&models.SessionKey{}).
		Where("user_id = ? AND session_key = ?", userID, sessionKey).
		Update("valid", "0").Error; err != nil {
		log.Printf("Failed to invalidate session: %v", err)
	} else {
		log.Printf("Session invalidated successfully for user: %d", userID)
	}
}

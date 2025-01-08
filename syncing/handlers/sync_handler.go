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
		case "add_session":
			h.handleAddSessionKey(payload)
		case "add_session_refresh":
			h.handleAddRefreshKey(payload)
		case "delete_session":
			h.handleInvalidateSessionKey(payload)
		default:
			log.Printf("Unknown task type: %s", taskType)
		}
	}
}

func (h *SyncHandler) handleNoteUpdate(payload map[string]interface{}) {
	// For numeric ID, first assert to float64, then convert to uint
	noteIDFloat, ok := payload["noteid"].(float64)
	if !ok {
		log.Printf("Invalid note ID for note update: %v", payload)
		return
	}
	noteID := uint(noteIDFloat)

	// String assertions
	hash, ok := payload["h"].(string)
	if !ok {
		log.Printf("Invalid content hash \"h\" for note update: %v", payload)
		return
	}

	headHash, ok := payload["intgrh"].(string)
	if !ok {
		log.Printf("Invalid heading hash \"intgrh\" for note update: %v", payload)
		return
	}

	content, ok := payload["content"].(string)
	if !ok {
		log.Printf("Invalid content for note update: %v", payload)
		return
	}

	heading, ok := payload["heading"].(string)
	if !ok {
		log.Printf("Invalid heading for note update: %v", payload)
		return
	}

	// For deleted flag, first assert to float64, then convert to int
	deletedFloat, ok := payload["deleted"].(float64)
	if !ok {
		log.Printf("Invalid deleted for note update: %v", payload)
		return
	}
	deleted := int(deletedFloat)

	log.Printf("Received RabbitMQ for note update for %s\n", noteID)

	var note models.Notes
	if err := h.DB.First(&note, "noteid = ?", noteID).Error; err != nil {
		log.Printf("Note not found: %v", err)
		return
	}

	updateTime, ok := payload["update_time"].(float64)
	if !ok {
		log.Printf("Invalid update time for note update: %v", payload)
		return
	}

	note.Content = content
	note.UpdateTime = time.Now()
	note.H = hash
	note.Heading = heading
	note.Intgrh = headHash
	note.Deleted = deleted
	note.UpdateTime = time.Unix(int64(updateTime), 0)

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

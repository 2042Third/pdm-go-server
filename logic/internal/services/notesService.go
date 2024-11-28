package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"pdm-go-server/internal/models"
	"time"
)

func GetNotes(S *Storage, ctx context.Context, userID uint) ([]models.Notes, error) {
	var notes []models.Notes

	keysPattern := fmt.Sprintf("user:%d:note:*", userID)
	keys, err := S.Ch.Keys(ctx, keysPattern)
	if err != nil {
		log.Printf("Failed to retrieve keys: %v", err)
		return nil, err
	}
	if len(keys) == 0 {
		log.Printf("Cache miss or failure to find note in cache: %v", err)

	} else {
		log.Printf("Cache hit - found %d keys", len(keys))
		for _, key := range keys {
			jsonData, err := S.Ch.Get(ctx, key)
			if err == nil {
				// Cache hit - need to deserialize
				var note models.Notes
				err = json.Unmarshal([]byte(jsonData), &note)
				if err == nil {
					notes = append(notes, note)
				} else {
					// If unmarshal fails, log and continue to next key
					log.Printf("Failed to unmarshal cached note: %v", err)
				}
			} else {
				// Log cache miss or error
				log.Printf("Cache retrieve of existing key failed %v", err)
			}
		}
	}

	// Cache miss or unmarshal error, get from DB
	err = S.DB.Model(&models.Notes{}).
		Select("noteid", "userid", "heading", "time", "h", "update_time", "intgrh", "content", "deleted").
		Where("userid = ?", userID).
		Find(&notes).Error
	if err != nil {
		// Handle the error
		fmt.Println("Error:", err)
	} else {
		// Use the notes slice
		//fmt.Println("Notes:", notes)
	}

	// Cache the result for next time
	for _, note := range notes {
		key := fmt.Sprintf("user:%d:note:%d", note.UserID, note.NoteID)
		bytes, err := json.Marshal(note)
		if err == nil {
			jsonData := string(bytes)
			err = S.Ch.Set(ctx, key, jsonData, 24*time.Hour)
			if err != nil {
				log.Printf("Failed to cache notes: %v", err)
			}
		} else {
			log.Printf("Failed to marshal note: %v", err)
		}
	}
	return notes, nil
}

func GetNoteByID(S *Storage, ctx context.Context, userID uint, noteID uint) (models.Notes, error) {
	var note models.Notes

	key := fmt.Sprintf("user:%d:note:%d", userID, noteID)
	jsonData, err := S.Ch.Get(ctx, key)
	if err == nil {
		// Cache hit - need to deserialize
		err = json.Unmarshal([]byte(jsonData), &note)
		if err == nil {
			return note, nil
		}
		// If unmarshal fails, log and continue to DB
		log.Printf("Failed to unmarshal cached noteInfo: %v", err)
	}

	// Cache miss or unmarshal error, get from DB
	err = S.DB.First(&note, noteID).Error
	if err != nil {
		return note, err
	}

	// Cache the result for next time
	bytes, err := json.Marshal(note)
	jsonData = string(bytes)
	if err == nil {
		err = S.Ch.Set(ctx, key, jsonData, 24*time.Hour)
		if err != nil {
			log.Printf("Failed to cache noteInfo: %v", err)
		}
	}

	return note, nil
}

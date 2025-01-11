package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"pdm-logic-server/pkg/models"
	"time"
)

func (s *Storage) GetNotes(ctx context.Context, userID string, cacheTTL int) ([]models.Notes, error) {
	var notes []models.Notes

	keysPattern := fmt.Sprintf("user:%s:note:*", userID)
	keys, err := s.Ch.Keys(ctx, keysPattern)
	if err != nil {
		log.Printf("Failed to retrieve keys: %v", err)
		return nil, err
	}
	if len(keys) == 0 {
		log.Printf("Cache miss or failure to find note in cache: %v", err)

	} else {
		log.Printf("Cache hit - found %d keys", len(keys))
		for _, key := range keys {
			jsonData, err := s.Ch.Get(ctx, key)
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
	err = s.DB.Model(&models.Notes{}).
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
		key := fmt.Sprintf("user:%s:note:%s", note.UserID, note.NoteID)
		bytes, err := json.Marshal(note)
		if err == nil {
			jsonData := string(bytes)
			err = s.Ch.Set(ctx, key, jsonData, time.Duration(cacheTTL)*time.Minute)
			if err != nil {
				log.Printf("Failed to cache notes: %v", err)
			}
		} else {
			log.Printf("Failed to marshal note: %v", err)
		}
	}

	return notes, nil
}

func (s *Storage) GetNoteByID(ctx context.Context, userID string, noteID string, cacheTTL int) (models.Notes, error) {
	var note models.Notes

	key := fmt.Sprintf("user:%s:note:%s", userID, noteID)
	jsonData, err := s.Ch.Get(ctx, key)
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
	err = s.DB.First(&note, noteID).Error
	if err != nil {
		return note, err
	}

	// Cache the result for next time
	bytes, err := json.Marshal(note)
	jsonData = string(bytes)
	if err == nil {
		err = s.Ch.Set(ctx, key, jsonData, time.Duration(cacheTTL)*time.Minute)
		if err != nil {
			log.Printf("Failed to cache noteInfo: %v", err)
		}
	}

	return note, nil
}

// CreateNote creates a new note for the given user in the storage layer.
// It stores the note in the database and caches it for 24 hours.
//
// Parameters:
//   - ctx: context.Context for request cancellation and timeouts
//   - userId: uint representing the ID of the user who owns the note
//
// Returns:
//   - models.Notes: the newly created note
//   - error: nil if successful, otherwise contains the error that occurred
//
// The cache key is formatted as "user:{userId}:note:{noteId}". If caching fails,
// the error is logged but the function will still return successfully.
func (s *Storage) CreateNote(ctx context.Context, userId string, cacheTTL int) (models.Notes, error) {
	note := models.Notes{
		UserID: userId,
	}

	// Save the note to the database
	db := s.DB.Create(&note)
	if db.Error != nil {
		return note, db.Error
	}

	// Cache the result for next time
	key := fmt.Sprintf("user:%s:note:%s", note.UserID, note.NoteID)
	bytes, err := json.Marshal(note)
	if err == nil {
		jsonData := string(bytes)
		err = s.Ch.Set(ctx, key, jsonData, time.Duration(cacheTTL)*time.Minute)
		if err != nil {
			log.Printf("Failed to cache note: %v", err)
		}
	} else {
		log.Printf("Failed to marshal note: %v", err)
	}

	return note, nil
}

func (s *Storage) UpdateNote(ctx context.Context, note models.Notes, cacheTTL int) error {
	note.UpdateTime = time.Now()

	log.Printf("[DEBUG, func (s *Storage) UpdateNote] note.Time: %v", note.Time)

	// Save the note to the database through rabbitmq
	err := s.R.DispatchNoteUpdate(note)
	if err != nil {
		log.Printf("Failed to dispatch note update: %v", err)
		return err
	}

	// Cache the changed note
	key := fmt.Sprintf("user:%s:note:%s", note.UserID, note.NoteID)
	bytes, err := json.Marshal(note)
	if err == nil {
		jsonData := string(bytes)
		err = s.Ch.Set(ctx, key, jsonData, time.Duration(cacheTTL)*time.Minute)
		if err != nil {
			log.Printf("Failed to cache note: %v", err)
		}
	} else {
		log.Printf("Failed to marshal note: %v", err)
	}

	return nil
}

func (s *Storage) DeleteNote(ctx context.Context, userId string, req models.DeleteNoteRequest) error {

	// Save the note to the database through rabbitmq
	err := s.R.DispatchNoteDelete(req)
	if err != nil {
		log.Printf("Failed to dispatch note update: %v", err)
		return err
	}

	// Delete the note from the cache
	key := fmt.Sprintf("user:%s:note:%s", userId, req.NoteID)
	err = s.Ch.Delete(ctx, key)
	if err != nil {
		log.Printf("Failed to delete note from cache: %v", err)
	}

	return nil
}

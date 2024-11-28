package handlers

import (
	"context"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/streadway/amqp"
	"log"
	"net/http"
	"pdm-go-server/internal/auth"
	"pdm-go-server/internal/models"
	"pdm-go-server/internal/services"
	"strconv"
)

type NotesHandler struct {
	S           *services.Storage
	AuthService *auth.AuthService
}

func NewNotesHandler(storage *services.Storage, authService *auth.AuthService) *NotesHandler {
	return &NotesHandler{S: storage, AuthService: authService}
}

func (h *NotesHandler) CreateNote(c echo.Context) error {
	//ctx := context.Background()
	log.Println("Login request received")
	creds := new(models.Notes)
	if err := c.Bind(creds); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
	}
	return c.JSON(http.StatusInternalServerError, map[string]string{
		"message": "Not implemented",
	})
}

func (h *NotesHandler) GetNotes(c echo.Context) error {
	ctx := context.Background()

	userEmail := c.Get("email").(string)
	log.Println("Get notes request received from user " + userEmail)

	userId, err := h.S.Ch.HGet(ctx, "userEmail:userId", userEmail)
	if err != nil {
		log.Printf("Failed to get userId for userEmail %s: %v", userEmail, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, HGet userEmail:userId",
		})
	}

	intUserId, err := strconv.Atoi(userId)
	if err != nil {
		log.Printf("Failed to convert userId to int: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Failed to convert userId to int",
		})
	}
	notes, err := services.GetNotes(h.S, ctx, uint(intUserId))
	if err != nil {
		log.Printf("Failed to get notes: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Failed to get notes",
		})
	}

	return c.JSON(http.StatusOK, notes)
}

func (h *NotesHandler) UpdateNotes(c echo.Context) error {
	//ctx := context.Background()

	return c.JSON(http.StatusInternalServerError, map[string]string{
		"message": "Not implemented",
	})
}

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

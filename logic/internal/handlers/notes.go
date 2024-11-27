package handlers

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/streadway/amqp"
	"log"
	"net/http"
	"pdm-go-server/internal/auth"
	"pdm-go-server/internal/models"
	"pdm-go-server/internal/services"
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
	//ctx := context.Background()

	return c.JSON(http.StatusInternalServerError, map[string]string{
		"message": "Not implemented",
	})
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

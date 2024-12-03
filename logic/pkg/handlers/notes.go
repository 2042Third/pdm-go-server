package handlers

import (
	"context"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/streadway/amqp"
	"log"
	"net/http"
	"pdm-logic-server/pkg/errors"
	"pdm-logic-server/pkg/models"
	"strconv"
)

type NotesHandler struct {
	*BaseHandler
}

func NewNotesHandler(base *BaseHandler) *NotesHandler {
	return &NotesHandler{BaseHandler: base}
}

func (h *NotesHandler) GetNotes(c echo.Context) error {
	ctx := context.Background()

	userEmail := c.Get("email").(string)
	userId, err := h.getUserId(ctx, userEmail)
	if err != nil {
		return err
	}

	notes, err := h.storage.GetNotes(ctx, userId)
	if err != nil {
		return errors.NewAppError(http.StatusInternalServerError, "Failed to fetch notes", err)
	}

	return c.JSON(http.StatusOK, notes)
}

func (h *NotesHandler) getUserId(ctx context.Context, email string) (uint, error) {
	userIdStr, err := h.storage.Ch.HGet(ctx, "userEmail:userId", email)
	if err != nil {
		return 0, errors.NewAppError(http.StatusInternalServerError, "Failed to get user ID", err)
	}

	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		return 0, errors.NewAppError(http.StatusInternalServerError, "Invalid user ID format", err)
	}

	return uint(userId), nil
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

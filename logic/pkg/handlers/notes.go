package handlers

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"pdm-logic-server/pkg/errors"
	"pdm-logic-server/pkg/models"
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

	notes, err := h.storage.GetNotes(ctx, userId, h.config.Redis.NotesCacheTTLMinutes)
	if err != nil {
		return errors.NewAppError(http.StatusInternalServerError, "Failed to fetch notes", err)
	}

	return c.JSON(http.StatusOK, notes)
}

func (h *NotesHandler) getUserId(ctx context.Context, email string) (string, error) {
	userIdStr, err := h.storage.Ch.HGet(ctx, "userEmail:userId", email)
	if err != nil {
		return "", errors.NewAppError(http.StatusInternalServerError, "Failed to get user ID", err)
	}

	return userIdStr, nil
}

func (h *NotesHandler) CreateNote(c echo.Context) error {
	ctx := context.Background()

	userId := c.Get("userId").(string)
	note, err := h.storage.CreateNote(ctx, userId, h.config.Redis.NotesCacheTTLMinutes)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to create note",
			"error":   err.Error(),
		})
	}

	fmt.Printf("Created note: %v\n", note.NoteID)

	return c.JSON(http.StatusOK, note)
}

func (h *NotesHandler) UpdateNotes(c echo.Context) error {
	var req models.Notes
	if err := c.Bind(&req); err != nil {
		return errors.NewAppError(http.StatusBadRequest, "Invalid request format", err)
	}

	if err := c.Validate(&req); err != nil {
		return errors.NewAppError(http.StatusBadRequest, "Invalid request data", err)
	}

	log.Printf("[DEBUG, func (h *NotesHandler) UpdateNotes] req.Time: %v", req.Time)
	ctx := context.Background()

	err := h.storage.UpdateNote(ctx, req, h.config.Redis.NotesCacheTTLMinutes)
	if err != nil {
		return errors.NewAppError(http.StatusInternalServerError, "Failed to update note", err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "note updated",
	})
}

func (h *NotesHandler) DeleteNotes(c echo.Context) error {
	var req models.DeleteNoteRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewAppError(http.StatusBadRequest, "Invalid request format", err)
	}

	if err := c.Validate(&req); err != nil {
		return errors.NewAppError(http.StatusBadRequest, "Invalid request data", err)
	}

	ctx := context.Background()

	userId := c.Get("userId").(string)
	err := h.storage.DeleteNote(ctx, userId, req)
	if err != nil {
		return errors.NewAppError(http.StatusInternalServerError, "Failed to update note", err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "note deleted",
	})
}

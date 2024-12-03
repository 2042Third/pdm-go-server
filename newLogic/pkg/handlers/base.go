package handlers

import (
	"github.com/sirupsen/logrus"
	"pdm-logic-server/pkg/services"
)

type BaseHandler struct {
	storage     *services.Storage
	authService *services.AuthService
	log         *logrus.Logger
}

func NewBaseHandler(storage *services.Storage, authService *services.AuthService, logger *logrus.Logger) *BaseHandler {
	return &BaseHandler{
		storage:     storage,
		authService: authService,
		log:         logger,
	}
}

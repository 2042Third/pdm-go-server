package middleware

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"pdm-logic-server/pkg/errors"
)

func ErrorHandler(err error, c echo.Context) error {
	if e, ok := err.(*errors.AppError); ok {
		return c.JSON(e.Code, map[string]string{
			"message": e.Message,
		})
	}

	if he, ok := err.(*echo.HTTPError); ok {
		return c.JSON(he.Code, map[string]string{
			"message": he.Message.(string),
		})
	}

	return c.JSON(http.StatusInternalServerError, map[string]string{
		"message": "Internal server error",
	})
}

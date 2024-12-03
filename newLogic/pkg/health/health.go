package health

import (
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
	"pdm-logic-server/pkg/cache"
	"pdm-logic-server/pkg/db"
	"time"
)

type HealthChecker struct {
	db    *db.Database
	cache cache.Cache
}

type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

func NewHealthChecker(db *db.Database, cache cache.Cache) *HealthChecker {
	return &HealthChecker{
		db:    db,
		cache: cache,
	}
}

func (h *HealthChecker) Handler(c echo.Context) error {
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services:  make(map[string]string),
	}

	// Check database
	//if err := h.db.PingContext(ctx); err != nil {
	//	status.Status = "unhealthy"
	//	status.Services["database"] = "unhealthy"
	//} else {
	//	status.Services["database"] = "healthy"
	//}

	// Check Redis
	//if err := h.cache.Ping(ctx); err != nil {
	//	status.Status = "unhealthy"
	//	status.Services["cache"] = "unhealthy"
	//} else {
	//	status.Services["cache"] = "healthy"
	//}

	if status.Status == "healthy" {
		return c.JSON(http.StatusOK, status)
	}
	return c.JSON(http.StatusServiceUnavailable, status)
}

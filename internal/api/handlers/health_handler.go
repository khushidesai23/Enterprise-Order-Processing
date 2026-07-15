package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/khushidesai23/Enterprise-Order-Processing/config"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/response"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/database"
)

type HealthHandler struct {
	cfg *config.Config
	db  *database.Database
}

func NewHealthHandler(
	cfg *config.Config,
	db *database.Database,
) *HealthHandler {

	return &HealthHandler{
		cfg: cfg,
		db:  db,
	}
}

func (h *HealthHandler) Root(c *gin.Context) {

	response.OK(c, "welcome", gin.H{
		"application": h.cfg.AppName,
		"status":      "running",
	})
}

func (h *HealthHandler) Ping(c *gin.Context) {

	response.OK(c, "pong", gin.H{
		"timestamp": time.Now().UTC(),
	})
}

func (h *HealthHandler) Version(c *gin.Context) {

	response.OK(c, "version", gin.H{
		"version": "v1.0.0",
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {

	if !h.db.Healthy() {

		response.Error(
			c,
			http.StatusServiceUnavailable,
			"database unavailable",
		)

		return
	}

	response.OK(c, "ready", gin.H{
		"database": "connected",
	})
}

func (h *HealthHandler) Health(c *gin.Context) {

	dbStatus := "down"

	if h.db.Healthy() {
		dbStatus = "up"
	}

	response.OK(c, "healthy", gin.H{
		"status": "UP",
		"database": gin.H{
			"status": dbStatus,
		},
		"timestamp": time.Now().UTC(),
	})
}
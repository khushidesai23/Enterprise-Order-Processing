package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/database"
)

func Ping(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func Health(c *gin.Context) {

	sqlDB, err := database.GetDB().DB()

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "database unavailable",
		})

		return
	}

	if err := sqlDB.Ping(); err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "database disconnected",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "UP",
		"database": "UP",
	})
}
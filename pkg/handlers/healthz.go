package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stv0g/gose/pkg/server"
)

// HandleConfigWith returns runtime configuration to the frontend
func HandleHealthz(c *gin.Context) {
	svrs := c.MustGet("servers").(server.List)

	// TODO: check health status of notifier?
	// TODO: check health status of shortener?
	// shortener := c.MustGet("shortener").(*shortener.Shortener)

	for _, svr := range svrs {
		if !svr.Healthy() {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}
}
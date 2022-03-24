package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/stv0g/gose/pkg/config"
)

type featureResponse struct {
	Shortener  bool `json:"shortener"`
	NotifyMail bool `json:"notify_mail"`
}

type configResponse struct {
	ExpirationClasses []config.ExpirationClass `json:"expiration_classes"`
	Features          featureResponse          `json:"features"`
}

// HandleConfig returns runtime configuration to the frontend
func HandleConfig(c *gin.Context) {
	cfg, _ := c.MustGet("cfg").(*config.Config)

	c.JSON(200, &configResponse{
		ExpirationClasses: cfg.S3.Expiration.Classes,
		Features: featureResponse{
			Shortener:  cfg.Shortener != nil,
			NotifyMail: false,
		},
	})
}

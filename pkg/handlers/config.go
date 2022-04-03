package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/stv0g/gose/pkg/config"
	"github.com/stv0g/gose/pkg/server"
)

type featureResponse struct {
	ShortURL      bool `json:"short_url"`
	NotifyMail    bool `json:"notify_mail"`
	NotifyBrowser bool `json:"notify_browser"`
}

type configResponse struct {
	Servers  []config.S3ServerConfig `json:"servers"`
	Features featureResponse         `json:"features"`
}

// HandleConfig returns runtime configuration to the frontend
func HandleConfig(c *gin.Context) {
	cfg := c.MustGet("config").(*config.Config)
	svrs := c.MustGet("servers").(server.List)

	svrsResp := []config.S3ServerConfig{}
	for _, svr := range svrs {
		svrsResp = append(svrsResp, svr.Config.S3ServerConfig)
	}

	c.JSON(200, &configResponse{
		Servers: svrsResp,
		Features: featureResponse{
			ShortURL:      cfg.Shortener != nil,
			NotifyMail:    cfg.Notification.Mail != nil,
			NotifyBrowser: true,
		},
	})
}

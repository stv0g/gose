package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/containrrr/shoutrrr/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stv0g/gose/pkg/config"
	"github.com/stv0g/gose/pkg/notifier"
	"github.com/stv0g/gose/pkg/server"
)

func HandleDownload(c *gin.Context) {
	var err error

	svrs := c.MustGet("servers").(server.List)
	cfg := c.MustGet("config").(*config.Config)

	key := c.Param("key")
	key = key[1:]

	svr, ok := svrs[c.Param("server")]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid server"})
		return
	}

	parts := strings.Split(key, "/")
	if len(parts) != 2 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid key"})
		return
	}

	if _, err := uuid.Parse(parts[0]); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid uuid in key"})
		return
	}

	req, _ := svr.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(svr.Config.Bucket),
		Key:    aws.String(key),
	})

	u, _, err := req.PresignRequest(10 * time.Second)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to presign request: %s", err)})
		return
	}

	go func(svr server.Server, key string) {
		if cfg.Notification != nil && cfg.Notification.Downloads {
			if notif, err := notifier.NewNotifier(cfg.Notification.Template, cfg.Notification.URLs...); err != nil {
				log.Fatalf("Failed to create notification sender: %s", err)
			} else {
				if err := notif.Notify(svr, key, types.Params{
					"Title": "New download",
				}); err != nil {
					fmt.Printf("Failed to send notification: %s", err)
				}
			}
		}
	}(svr, key)

	c.Redirect(http.StatusTemporaryRedirect, u)
}

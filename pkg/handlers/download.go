package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stv0g/gose/pkg/config"
	"github.com/stv0g/gose/pkg/notifier"
)

func HandleDownload(c *gin.Context) {
	var err error

	svc, _ := c.MustGet("s3").(*s3.S3)
	cfg, _ := c.MustGet("cfg").(*config.Config)
	notifier, _ := c.MustGet("notifier").(*notifier.Notifier)

	key := c.Param("key")
	key = key[1:]

	parts := strings.Split(key, "/")
	if len(parts) != 2 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid key"})
		return
	}

	if _, err := uuid.Parse(parts[0]); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid uuid in key"})
		return
	}

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(cfg.S3.Bucket),
		Key:    aws.String(key),
	})

	u, _, err := req.PresignRequest(10 * time.Second)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to presign request: %s", err)})
		return
	}

	if notifier != nil && cfg.Notification.Downloads {
		go func(s3svc *s3.S3, cfg *config.Config, key string) {
			if err := notifier.Notify(svc, cfg, key, "New download"); err != nil {
				fmt.Printf("Failed to send notification: %s", err)
			}
		}(svc, cfg, key)
	}

	c.Redirect(http.StatusTemporaryRedirect, u)
}

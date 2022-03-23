package handlers

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/stv0g/gose/pkg/config"
	"github.com/stv0g/gose/pkg/notifier"
)

type part struct {
	PartNumber int64  `json:"part_number"`
	Checksum   string `json:"checksum"`
	ETag       string `json:"etag"`
}

type completionRequest struct {
	Key      string `json:"key"`
	UploadID string `json:"upload_id"`
	Parts    []part `json:"parts"`
}

type completionResponse struct {
	ETag string `json:"etag"`
}

// HandleComplete handles a completed upload
func HandleComplete(c *gin.Context) {
	svc, _ := c.MustGet("s3").(*s3.S3)
	cfg, _ := c.MustGet("cfg").(*config.Config)
	notifier, _ := c.MustGet("notifier").(*notifier.Notifier)

	var req completionRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "malformed request"})
		return
	}

	parts := []*s3.CompletedPart{}
	for _, part := range req.Parts {
		parts = append(parts, &s3.CompletedPart{
			PartNumber: aws.Int64(part.PartNumber),
			ETag:       aws.String(part.ETag),
		})
	}

	respCompleteMPU, err := svc.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(cfg.S3.Bucket),
		Key:      aws.String(req.Key),
		UploadId: aws.String(req.UploadID),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: parts,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if notifier != nil && cfg.Notification.Uploads {
		go func(s3svc *s3.S3, cfg *config.Config, key string) {
			if err := notifier.Notify(svc, cfg, key, "New upload"); err != nil {
				fmt.Printf("Failed to send notification: %s", err)
			}
		}(svc, cfg, *respCompleteMPU.Key)
	}

	c.JSON(200, &completionResponse{
		ETag: *respCompleteMPU.ETag,
	})
}

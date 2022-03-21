package handlers

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/stv0g/gose/backend/config"
	"github.com/stv0g/gose/backend/notifier"
)

type Part struct {
	PartNumber int64  `json:"part_number"`
	Checksum   string `json:"checksum"`
	ETag       string `json:"etag"`
}

type CompletionRequest struct {
	Key      string `json:"key"`
	UploadId string `json:"upload_id"`
	Parts    []Part `json:"parts"`
}

type CompletionResponse struct {
	ETag string `json:"etag"`
}

func HandleComplete(c *gin.Context) {
	svc, _ := c.MustGet("s3").(*s3.S3)
	cfg, _ := c.MustGet("cfg").(*config.Config)
	notifier, _ := c.MustGet("notifier").(*notifier.Notifier)

	var req CompletionRequest
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
		UploadId: aws.String(req.UploadId),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: parts,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if notifier != nil {
		go func(s3svc *s3.S3, cfg *config.Config, key string) {
			if err := notifier.Notify(svc, cfg, key); err != nil {
				fmt.Printf("Failed to send notification: %s", err)
			}
		}(svc, cfg, *respCompleteMPU.Key)
	}

	c.JSON(200, &CompletionResponse{
		ETag: *respCompleteMPU.ETag,
	})
}

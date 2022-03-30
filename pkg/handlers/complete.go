package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/containrrr/shoutrrr/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/stv0g/gose/pkg/config"
	"github.com/stv0g/gose/pkg/notifier"
	"github.com/stv0g/gose/pkg/server"
)

type part struct {
	PartNumber int64  `json:"part_number"`
	Checksum   string `json:"checksum"`
	ETag       string `json:"etag"`
}

type completionRequest struct {
	Server     string  `json:"server"`
	Key        string  `json:"key"`
	UploadID   string  `json:"upload_id"`
	Parts      []part  `json:"parts"`
	NotifyMail *string `json:"notify_mail"`
	Expiration *string `json:"expiration"`
}

type completionResponse struct {
	ETag string `json:"etag"`
}

// HandleComplete handles a completed upload
func HandleComplete(c *gin.Context) {
	svrs := c.MustGet("servers").(server.List)
	cfg := c.MustGet("config").(*config.Config)

	var req completionRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "malformed request"})
		return
	}

	svr, ok := svrs[req.Server]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid server"})
		return
	}

	// Ceph's RadosGW does not yet support tagging during the initiation of multi-part uploads.
	// So we tag here with a separate request instead of the MPU initiate req.
	//  See: https://github.com/ceph/ceph/pull/38275
	var expiration string
	if req.Expiration == nil {
		if len(svr.Config.Expiration) > 0 {
			expiration = svr.Config.Expiration[0].ID
		}
	} else {
		if !svr.HasExpirationClass(*req.Expiration) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid expiration class"})
			return
		}

		expiration = *req.Expiration
	}

	// Prepare MPU completion request
	parts := []*s3.CompletedPart{}
	for _, part := range req.Parts {
		parts = append(parts, &s3.CompletedPart{
			PartNumber: aws.Int64(part.PartNumber),
			ETag:       aws.String(part.ETag),
		})
	}

	respCompleteMPU, err := svr.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(svr.Config.Bucket),
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

	// Tag object with expiration tag here
	if _, err := svr.PutObjectTagging(&s3.PutObjectTaggingInput{
		Bucket: aws.String(svr.Config.Bucket),
		Key:    aws.String(req.Key),
		Tagging: &s3.Tagging{
			TagSet: []*s3.Tag{
				{
					Key:   aws.String("expiration"),
					Value: aws.String(expiration),
				},
			},
		},
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to tag object"})
		return
	}

	// Send notifications
	go func(key string) {
		if cfg.Notification != nil && cfg.Notification.Uploads {
			if notif, err := notifier.NewNotifier(cfg.Notification.Template, cfg.Notification.URLs...); err != nil {
				log.Fatalf("Failed to create notification sender: %s", err)
			} else {
				if err := notif.Notify(svr, key, types.Params{
					"Title": "New upload",
				}); err != nil {
					fmt.Printf("Failed to send notification: %s", err)
				}
			}
		}

		if cfg.Notification.Mail != nil && req.NotifyMail != nil {
			u := fmt.Sprintf("%s&ToAddresses=%s", cfg.Notification.Mail.URL, *req.NotifyMail)
			if notif, err := notifier.NewNotifier(cfg.Notification.Mail.Template, u); err != nil {
				log.Fatalf("Failed to create notification sender: %s", err)
			} else {
				if err := notif.Notify(svr, key, types.Params{
					"Title": "New upload",
				}); err != nil {
					fmt.Printf("Failed to send notification: %s", err)
				}
			}
		}
	}(*respCompleteMPU.Key)

	c.JSON(200, &completionResponse{
		ETag: *respCompleteMPU.ETag,
	})
}

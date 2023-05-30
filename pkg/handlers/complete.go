// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/containrrr/shoutrrr/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/stv0g/gose/pkg/config"
	"github.com/stv0g/gose/pkg/notifier"
	"github.com/stv0g/gose/pkg/server"
	"github.com/stv0g/gose/pkg/utils"
)

type completionRequest struct {
	Server     string  `json:"server"`
	ETag       string  `json:"etag"`
	UploadID   string  `json:"upload_id"`
	Parts      []part  `json:"parts"`
	NotifyMail *string `json:"notify_mail"`
	Expiration *string `json:"expiration"`
}

type completionResponse struct {
	ETag string `json:"etag"`
	URL  string `json:"url"`
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

	if !utils.IsValidETag(req.ETag) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid etag"})
		return
	}

	// Ceph's RadosGW does not yet support tagging during the initiation of multi-part uploads.
	// So we tag here with a separate request instead of the MPU initiate req.
	//  See: https://github.com/ceph/ceph/pull/38275
	var exp *config.Expiration
	if req.Expiration == nil {
		if len(svr.Config.Expiration) > 0 {
			exp = &svr.Config.Expiration[0]
		}
	} else {
		if exp = svr.GetExpirationClass(*req.Expiration); exp == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid expiration class"})
			return
		}
	}

	if len(req.Parts) > int(svr.Config.MaxUploadSize/svr.Config.PartSize) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "max upload size exceeded"})
		return
	}

	// Prepare MPU completion request
	parts := []*s3.CompletedPart{}
	for _, part := range req.Parts {
		parts = append(parts, &s3.CompletedPart{
			PartNumber: aws.Int64(part.Number),
			ETag:       aws.String(part.ETag),
		})
	}

	respCompleteMPU, err := svr.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(svr.Config.Bucket),
		Key:      aws.String(req.ETag),
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
	if exp != nil {
		if _, err := svr.PutObjectTagging(&s3.PutObjectTaggingInput{
			Bucket: aws.String(svr.Config.Bucket),
			Key:    aws.String(req.ETag),
			Tagging: &s3.Tagging{
				TagSet: []*s3.Tag{
					{
						Key:   aws.String("expiration"),
						Value: aws.String(exp.ID),
					},
				},
			},
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to tag object"})
			return
		}
	}

	// Retrieve meta-data
	obj, err := svr.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(svr.Config.Bucket),
		Key:    aws.String(req.ETag),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get object"})
		return
	}

	var url string
	if u, ok := obj.Metadata["Original-Short-Url"]; ok {
		url = *u
	} else {
		url = svr.GetObjectURL(req.ETag).String()
	}

	// Send notifications
	go func(key string) {
		if cfg.Notification != nil && cfg.Notification.Uploads {
			if notif, err := notifier.NewNotifier(cfg.Notification.Template, cfg.Notification.URLs...); err != nil {
				log.Fatalf("Failed to create notification sender: %s", err)
			} else {
				if err := notif.Notify(url, obj, types.Params{
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
				if err := notif.Notify(url, obj, types.Params{
					"Title": "New upload",
				}); err != nil {
					fmt.Printf("Failed to send notification: %s", err)
				}
			}
		}
	}(req.ETag)

	c.JSON(200, &completionResponse{
		URL:  url,
		ETag: strings.Trim(*respCompleteMPU.ETag, "\""),
	})
}

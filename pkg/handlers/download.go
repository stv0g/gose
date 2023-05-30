// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/containrrr/shoutrrr/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/stv0g/gose/pkg/config"
	"github.com/stv0g/gose/pkg/notifier"
	"github.com/stv0g/gose/pkg/server"
	"github.com/stv0g/gose/pkg/utils"
	"github.com/vfaronov/httpheader"
)

// HandleDownload handles a request for downloading a file
func HandleDownload(c *gin.Context) {
	var err error

	svrs := c.MustGet("servers").(server.List)
	cfg := c.MustGet("config").(*config.Config)

	etag := c.Param("etag")
	fileName := c.Param("filename")
	svrName := c.Param("server")

	svr, ok := svrs[svrName]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid server"})
		return
	}

	if !utils.IsValidETag(etag) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid etag"})
		return
	}

	// Retrieve meta-data
	obj, err := svr.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(svr.Config.Bucket),
		Key:    aws.String(etag),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get object"})
		return
	}

	// RFC8187
	contentDisposition := "attachment; filename*=" + httpheader.EncodeExtValue(fileName, "")

	req, _ := svr.GetObjectRequest(&s3.GetObjectInput{
		Bucket:                     aws.String(svr.Config.Bucket),
		Key:                        aws.String(etag),
		ResponseContentDisposition: aws.String(contentDisposition),
		ResponseContentType:        aws.String(*obj.ContentType),
	})

	signedURL, _, err := req.PresignRequest(10 * time.Second)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to presign request: %s", err)})
		return
	}

	var shortURL string
	if u, ok := obj.Metadata["Original-Short-Url"]; ok {
		shortURL = *u
	} else {
		shortURL = svr.GetObjectURL(etag).String()
	}

	go func(svr server.Server, key string) {
		if cfg.Notification != nil && cfg.Notification.Downloads {
			if notif, err := notifier.NewNotifier(cfg.Notification.Template, cfg.Notification.URLs...); err != nil {
				log.Fatalf("Failed to create notification sender: %s", err)
			} else {
				if err := notif.Notify(shortURL, obj, types.Params{
					"Title": "New download",
				}); err != nil {
					fmt.Printf("Failed to send notification: %s", err)
				}
			}
		}
	}(svr, etag)

	c.Redirect(http.StatusTemporaryRedirect, signedURL)
}

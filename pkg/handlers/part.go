// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/stv0g/gose/pkg/server"
	"github.com/stv0g/gose/pkg/utils"
)

type partRequest struct {
	Server   string `json:"server"`
	ETag     string `json:"etag"`
	UploadID string `json:"upload_id"`
	Number   int    `json:"number"`
	Length   int    `json:"length"`
}

type partResponse struct {
	URL string `json:"url"`
}

// HandlePart initiates a new upload
func HandlePart(c *gin.Context) {
	svrs := c.MustGet("servers").(server.List)

	var req partRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "malformed request"})
		return
	}

	svr, ok := svrs[req.Server]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid server"})
		return
	}

	if req.Number <= 0 || req.Number >= utils.MaxPartCount {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid part number"})
		return
	}

	if !utils.IsValidETag(req.ETag) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid etag"})
		return
	}

	if req.Length > int(svr.Config.PartSize) {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid part size"})
		return
	}

	// For creating PutObject presigned URLs.
	partReq, _ := svr.UploadPartRequest(&s3.UploadPartInput{
		Bucket:        aws.String(svr.Config.Bucket),
		Key:           aws.String(req.ETag),
		UploadId:      aws.String(req.UploadID),
		ContentLength: aws.Int64(int64(req.Length)),
		PartNumber:    aws.Int64(int64(req.Number)),
	})

	u, _, err := partReq.PresignRequest(1 * time.Hour)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, partResponse{
		URL: u,
	})
}

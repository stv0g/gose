package handlers

import (
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stv0g/gose/pkg/config"
	"github.com/stv0g/gose/pkg/shortener"
)

type initiateRequest struct {
	ContentLength int64   `json:"content_length"`
	ContentType   string  `json:"content_type"`
	Filename      string  `json:"filename"`
	Expiration    *string `json:"expiration"`
	SSEKey        *string `json:"sse_key"`
	Shorten       bool    `json:"bool"`
}

type initiationResponse struct {
	Parts    []string `json:"parts"`
	Key      string   `json:"key"`
	UploadID string   `json:"upload_id"`
	PartSize int64    `json:"part_size"`
	URL      string   `json:"url"`
}

// HandleInitiate initiates a new upload
func HandleInitiate(c *gin.Context) {
	var err error

	svc, _ := c.MustGet("s3").(*s3.S3)
	cfg, _ := c.MustGet("cfg").(*config.Config)
	shortener, _ := c.MustGet("shortener").(*shortener.Shortener)

	var req initiateRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "malformed request"})
		return
	}

	if req.ContentLength <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid content length"})
		return
	}

	if req.ContentLength > int64(cfg.S3.MaxUploadSize) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is too large"})
		return
	}

	// TODO: perform proper validation of filenames
	// See: https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-keys.html
	if len(req.Filename) > 128 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}

	uid, err := uuid.NewRandom()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate UUID"})
		return
	}

	key := path.Join(uid.String(), req.Filename)

	u, _ := url.Parse(cfg.Server.BaseURL)
	u.Path += "api/v1/download/" + key
	if req.Shorten {
		if shortener == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "shortened URL requested but nut supported"})
			return
		}

		u, err = shortener.Shorten(u)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	var expiration string
	if req.Expiration == nil {
		expiration = cfg.S3.Expiration.Default
	} else {
		if !cfg.S3.Expiration.Supported(*req.Expiration) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid expiration class"})
			return
		}

		expiration = *req.Expiration
	}

	tags := url.Values{
		"expiration": []string{expiration},
	}

	reqCreateMPU := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(cfg.S3.Bucket),
		Key:    aws.String(key),
		Metadata: aws.StringMap(map[string]string{
			"uploaded-by": c.ClientIP(),
			"url":         u.String(),
		}),
		Tagging: aws.String(tags.Encode()),
	}

	respCreateMPU, err := svc.CreateMultipartUpload(reqCreateMPU)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	parts := []string{}
	numParts := req.ContentLength / int64(cfg.S3.PartSize)
	if req.ContentLength%int64(cfg.S3.PartSize) > 0 {
		numParts++
	}

	for partNum := int64(1); partNum <= numParts; partNum++ {
		partSize := int64(cfg.S3.PartSize)
		if partNum == numParts {
			partSize = req.ContentLength % int64(cfg.S3.PartSize)
		}

		// For creating PutObject presigned URLs
		req, _ := svc.UploadPartRequest(&s3.UploadPartInput{
			Bucket:            aws.String(cfg.S3.Bucket),
			Key:               aws.String(key),
			ContentLength:     aws.Int64(partSize),
			UploadId:          respCreateMPU.UploadId,
			PartNumber:        &partNum,
			ChecksumAlgorithm: aws.String("SHA256"),
		})

		u, _, err := req.PresignRequest(1 * time.Hour)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		parts = append(parts, u)
	}

	c.JSON(http.StatusOK, initiationResponse{
		URL:      u.String(),
		Parts:    parts,
		UploadID: *respCreateMPU.UploadId,
		Key:      key,
		PartSize: int64(cfg.S3.PartSize),
	})
}

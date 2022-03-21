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

type InitiateRequest struct {
	ContentLength int64   `json:"content_length"`
	ContentType   string  `json:"content_type"`
	Filename      string  `json:"filename"`
	Expiration    *string `json:"expiration"`
	SSEKey        *string `json:"sse_key"`
}

type InitiationResponse struct {
	Parts    []string `json:"parts"`
	Key      string   `json:"key"`
	UploadId string   `json:"upload_id"`
	PartSize int64    `json:"part_size"`
	URL      string   `json:"url"`
}

func HandleInitiate(c *gin.Context) {
	var err error

	svc, _ := c.MustGet("s3").(*s3.S3)
	cfg, _ := c.MustGet("cfg").(*config.Config)
	shortener, _ := c.MustGet("shortener").(*shortener.Shortener)

	var req InitiateRequest

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

	u := cfg.S3.GetObjectUrl(key)
	if shortener != nil {
		u, err = shortener.Shorten(u)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	tags := url.Values{}

	if req.Expiration != nil {
		tags["expiration"] = []string{*req.Expiration}
	}

	reqCreateMPU := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(cfg.S3.Bucket),
		Key:    aws.String(key),
		ACL:    aws.String("public-read"),
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

	c.JSON(http.StatusOK, InitiationResponse{
		URL:      u.String(),
		Parts:    parts,
		UploadId: *respCreateMPU.UploadId,
		Key:      key,
		PartSize: int64(cfg.S3.PartSize),
	})
}

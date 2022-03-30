package handlers

import (
	"log"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stv0g/gose/pkg/config"
	"github.com/stv0g/gose/pkg/server"
	"github.com/stv0g/gose/pkg/shortener"
)

type initiateRequest struct {
	Server        string `json:"server"`
	ContentLength int64  `json:"content_length"`
	ContentType   string `json:"content_type"`
	Filename      string `json:"filename"`
	ShortenLink   bool   `json:"shorten_link"`
}

type initiateResponse struct {
	Parts    []string `json:"parts"`
	Key      string   `json:"key"`
	UploadID string   `json:"upload_id"`
	PartSize int64    `json:"part_size"`
	URL      string   `json:"url"`
}

// HandleInitiate initiates a new upload
func HandleInitiate(c *gin.Context) {
	var err error

	svrs := c.MustGet("servers").(server.List)
	cfg := c.MustGet("config").(*config.Config)
	shortener := c.MustGet("shortener").(*shortener.Shortener)

	var req initiateRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "malformed request"})
		return
	}

	svr, ok := svrs[req.Server]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid server"})
		return
	}

	if req.ContentLength <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid content length"})
		return
	}

	if req.ContentLength > int64(svr.Config.MaxUploadSize) {
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

	u, _ := url.Parse(cfg.BaseURL)
	u.Path += filepath.Join("api/v1/download", req.Server, key)
	if req.ShortenLink {
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

	reqCreateMPU := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(svr.Config.Bucket),
		Key:    aws.String(key),
		Metadata: aws.StringMap(map[string]string{
			"uploaded-by": c.ClientIP(),
			"url":         u.String(),
		}),
	}

	log.Printf(" req: %+#v", reqCreateMPU)

	respCreateMPU, err := svr.CreateMultipartUpload(reqCreateMPU)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	parts := []string{}
	numParts := req.ContentLength / int64(svr.Config.PartSize)
	if req.ContentLength%int64(svr.Config.PartSize) > 0 {
		numParts++
	}

	for partNum := int64(1); partNum <= numParts; partNum++ {
		partSize := int64(svr.Config.PartSize)
		if partNum == numParts {
			partSize = req.ContentLength % int64(svr.Config.PartSize)
		}

		// For creating PutObject presigned URLs
		req, _ := svr.UploadPartRequest(&s3.UploadPartInput{
			Bucket:            aws.String(svr.Config.Bucket),
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

	c.JSON(http.StatusOK, initiateResponse{
		URL:      u.String(),
		Parts:    parts,
		UploadID: *respCreateMPU.UploadId,
		Key:      key,
		PartSize: int64(svr.Config.PartSize),
	})
}

package handlers

import (
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/stv0g/gose/pkg/config"
	"github.com/stv0g/gose/pkg/server"
	"github.com/stv0g/gose/pkg/shortener"
	"github.com/stv0g/gose/pkg/utils"
)

const (
	// MaxFileNameLength is the maximum file length which can be uploaded
	MaxFileNameLength = 256
)

type initiateRequest struct {
	Server   string `json:"server"`
	ETag     string `json:"etag"`
	FileName string `json:"filename"`
	ShortURL bool   `json:"short_url"`
	Type     string `json:"type"`
}

type initiateResponse struct {
	ETag string `json:"etag"`

	// We do not have a URL for resumed uploads due to limitations of the S3 API.
	URL string `json:"url,omitempty"`

	// An empty UploadID indicate that the file already existed
	UploadID string `json:"upload_id,omitempty"`

	Parts []part `json:"parts"`
}

// HandleInitiate initiates a new upload
func HandleInitiate(c *gin.Context) {
	var err error

	svrs := c.MustGet("servers").(server.List)
	shortener := c.MustGet("shortener").(*shortener.Shortener)
	cfg := c.MustGet("config").(*config.Config)

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

	if len(req.FileName) > MaxFileNameLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}

	if req.Type == "" {
		req.Type = "binary/octet-stream"
	}
	if _, _, err := mime.ParseMediaType(req.Type); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
		return
	}

	if !utils.IsValidETag(req.ETag) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid etag"})
		return
	}

	resp := initiateResponse{
		ETag:  req.ETag,
		Parts: []part{},
	}

	// Check if an object with this key already exists
	respObj, err := svr.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(svr.Config.Bucket),
		Key:    aws.String(resp.ETag),
	})

	u, _ := url.Parse(cfg.BaseURL)
	u.Path += filepath.Join("api/v1/download", req.Server, resp.ETag, req.FileName)

	// Object already exists
	if err == nil {
		if req.ShortURL {
			origShortURL, okURL := respObj.Metadata["Original-Short-Url"]
			origFileName, okName := respObj.Metadata["Original-Filename"]
			if okName && okURL && req.FileName == *origFileName {
				// This file is uploaded with the same name
				// So we can reuse the already shortened link
				resp.URL = *origShortURL
			} else {
				if shortener == nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "shortened URL requested but nut supported"})
					return
				}

				if u, err = shortener.Shorten(u); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				resp.URL = u.String()
			}
		} else {
			resp.URL = u.String()
		}
	} else {
		// Check if an upload has already been started
		respUploads, err := svr.ListMultipartUploads(&s3.ListMultipartUploadsInput{
			Bucket:     aws.String(svr.Config.Bucket),
			Prefix:     aws.String(resp.ETag),
			MaxUploads: aws.Int64(1),
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get uploads"})
			return
		}

		if len(respUploads.Uploads) > 0 {
			upload := respUploads.Uploads[0]

			respParts, err := svr.ListParts(&s3.ListPartsInput{
				Bucket:   aws.String(svr.Config.Bucket),
				Key:      aws.String(resp.ETag),
				UploadId: upload.UploadId,
			})
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get parts"})
				return
			}

			for _, p := range respParts.Parts {
				resp.Parts = append(resp.Parts, part{
					Number: *p.PartNumber,
					ETag:   strings.Trim(*p.ETag, "\""),
					Length: int(*p.Size),
				})
			}

			resp.UploadID = *upload.UploadId
		} else {
			meta := map[string]string{
				"Original-Uploader": c.ClientIP(),
				"Original-Filename": req.FileName,
			}

			// Shorten link
			if req.ShortURL {
				if shortener == nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "shortened URL requested but nut supported"})
					return
				}

				u, err = shortener.Shorten(u)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				meta["Original-Short-Url"] = u.String()
			}

			respCreateMPU, err := svr.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
				Bucket:      aws.String(svr.Config.Bucket),
				Key:         aws.String(resp.ETag),
				Metadata:    aws.StringMap(meta),
				ContentType: aws.String(req.Type),
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			resp.URL = u.String()
			resp.UploadID = *respCreateMPU.UploadId
		}
	}

	c.JSON(http.StatusOK, resp)
}

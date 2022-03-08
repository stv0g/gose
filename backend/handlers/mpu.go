package handlers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"

	"github.com/stv0g/gose/backend/config"
)

type MpuResponse struct {
	UploadID string `json:"upload_id"`
}

func HandleInitiateMPU(c *gin.Context) {
	svc, _ := c.MustGet("s3svc").(*s3.S3)
	cfg, _ := c.MustGet("cfg").(*config.Config)

	// Extract the object key from the path
	key := c.Params.ByName("key")

	out, err := svc.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(cfg.S3.Bucket),
		Key:    aws.String(key),
	})

	c.JSON(&MpuResponse{
		UploadID: err.UploadID
	})
}

func HandleCompleteMPU(c *gin.Context) {
	svc, _ := c.MustGet("s3svc").(*s3.S3)
	cfg, _ := c.MustGet("cfg").(*config.Config)

	// Extract the object key from the path
	key := c.Params.ByName("key")

	svc.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket: aws.String(cfg.S3.Bucket),
		Key:    aws.String(key),
	})
}

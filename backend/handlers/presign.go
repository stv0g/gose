package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"

	"github.com/stv0g/gose/backend/config"
)

// PresignResponse provides the Go representation of the JSON value that will be
// sent to the client.
type PresignResponse struct {
	Method string      `json:"method"`
	URL    string      `json:"url"`
	Header http.Header `json:"header"`
}

func HandlePresign(c *gin.Context) {
	var u string
	var err error
	var signedHeaders http.Header

	svc, _ := c.MustGet("s3svc").(*s3.S3)
	cfg, _ := c.MustGet("cfg").(*config.Config)

	var contentLen int64
	// Optionally the Content-Length header can be included with the signature
	// of the request. This is helpful to ensure the content uploaded is the
	// size that is expected. Constraints like these can be further expanded
	// with headers such as `Content-Type`. These can be enforced by the service
	// requiring the client to satisfying those constraints when uploading
	//
	// In addition the client could provide the service with a SHA256 of the
	// content to be uploaded. This prevents any other third party from uploading
	// anything else with the presigned URL
	if contLenStr, exists := c.GetQuery("contentLength"); exists {
		contentLen, err = strconv.ParseInt(contLenStr, 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to parse request content length, %v", err)
			c.String(http.StatusBadRequest, err.Error())
			return
		}
	}

	// Extract the object key from the path
	key := c.Params.ByName("key")
	method, _ := c.GetQuery("method")

	switch method {
	case "PUT":
		// For creating PutObject presigned URLs
		sdkReq, _ := svc.PutObjectRequest(&s3.PutObjectInput{
			Bucket: aws.String(cfg.S3.Bucket),
			Key:    aws.String(key),

			// If ContentLength is 0 the header will not be included in the signature.
			ContentLength: aws.Int64(contentLen),
		})
		u, signedHeaders, err = sdkReq.PresignRequest(15 * time.Minute)

	case "GET":
		// For creating GetObject presigned URLs
		fmt.Println("Received request to presign GetObject for,", key)
		sdkReq, _ := svc.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(cfg.S3.Bucket),
			Key:    aws.String(key),
		})
		u, signedHeaders, err = sdkReq.PresignRequest(15 * time.Minute)

	default:
		fmt.Fprintf(os.Stderr, "invalid method provided, %s, %v\n", method, err)
		err = fmt.Errorf("invalid request")
	}

	c.JSON(http.StatusOK, PresignResponse{
		Method: method,
		URL:    u,
		Header: signedHeaders,
	})
}

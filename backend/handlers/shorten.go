package handlers

import (
	"net/url"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/stv0g/Gose/backend/shortener"
)

type ShortenResponse struct {
	Url      *url.URL `json:"url"`
	ShortUrl *url.URL `json:"shorturl"`
}

func HandleShorten(c *gin.Context) {
	var err error

	s3svc, _ := c.MustGet("s3").(*s3.S3)
	shortener, _ := c.MustGet("shortener").(*shortener.Shortener)

	// Extract the object key from the path
	key := c.Params.ByName("key")

	req, _ := s3svc.GetObjectRequest(&s3.GetObjectInput{
		Key: &key,
	})

	r := ShortenResponse{
		Url: req.HTTPRequest.URL,
	}

	r.ShortUrl, err = shortener.Shorten(req.HTTPRequest.URL.String())
	if err != nil {
		c.AbortWithError(500, err)
	}

	c.JSON(200, &r)
}

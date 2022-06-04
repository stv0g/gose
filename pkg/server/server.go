package server

import (
	"net/url"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stv0g/gose/pkg/config"
)

// Server is a abstraction of an S3 server/bucket
type Server struct {
	*s3.S3

	Config *config.S3Server
}

// GetURL returns the full endpoint URL of the S3 server
func (s *Server) GetURL() *url.URL {
	u := &url.URL{}

	if s.Config.NoSSL {
		u.Scheme = "http"
	} else {
		u.Scheme = "https"
	}

	if s.Config.PathStyle {
		u.Host = s.Config.Endpoint
		u.Path = "/" + s.Config.Bucket
	} else {
		u.Host = s.Config.Bucket + "." + s.Config.Endpoint
		u.Path = ""
	}

	return u
}

// GetObjectURL returns the full URL to an object based on its key
func (s *Server) GetObjectURL(key string) *url.URL {
	u := s.GetURL()
	u.Path += "/" + key

	return u
}

// GetExpirationClass gets the expiration class by name
func (s *Server) GetExpirationClass(cls string) *config.Expiration {
	for _, c := range s.Config.Expiration {
		if c.ID == cls {
			return &c
		}
	}

	return nil
}

package server

import (
	"net/url"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stv0g/gose/pkg/config"
)

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
		u.Host = s.Endpoint
		u.Path = "/" + s.Config.Bucket
	} else {
		u.Host = s.Config.Bucket + "." + s.Endpoint
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

func (s *Server) HasExpirationClass(cls string) bool {
	for _, c := range s.Config.Expiration {
		if c.ID == cls {
			return true
		}
	}

	return false
}

// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stv0g/gose/pkg/config"
)

const (
	ImplementationAWS                = "AmazonS3"
	ImplementationMinio              = "MinIO"
	ImplementationGoogleCloudStorage = "UploadServer"
	ImplementationDigitalOceanSpaces = "DigitalOceanSpaces"
	ImplementationUnknown            = "Unknown"
)

// Server is a abstraction of an S3 server/bucket.
type Server struct {
	*s3.S3

	Config *config.S3Server
}

// GetURL returns the full endpoint URL of the S3 server.
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

// GetObjectURL returns the full URL to an object based on its key.
func (s *Server) GetObjectURL(key string) *url.URL {
	u := s.GetURL()
	u.Path += "/" + key

	return u
}

// GetExpirationClass gets the expiration class by name.
func (s *Server) GetExpirationClass(cls string) *config.Expiration {
	for _, c := range s.Config.Expiration {
		if c.ID == cls {
			return &c
		}
	}

	return nil
}

func (s *Server) DetectImplementation() string {
	if strings.Contains(s.Config.Endpoint, "digitaloceanspaces.com") {
		return ImplementationDigitalOceanSpaces
	} else if strings.Contains(s.Config.Endpoint, "storage.googleapis.com") {
		return ImplementationGoogleCloudStorage
	} else {
		req, _ := s.S3.ListBucketsRequest(&s3.ListBucketsInput{})
		req.Retryer = retryer{
			DefaultRetryer: client.DefaultRetryer{
				NumMaxRetries: 10,
			},
		}
		if err := req.Send(); err == nil {
			if svr := req.HTTPResponse.Header.Get("Server"); svr != "" {
				return svr
			}

			return ImplementationUnknown
		}

		return ImplementationUnknown
	}
}

// Healthy returns true if the S3 server is reachable and responds to our authenticated requests.
func (s *Server) Healthy() bool {
	_, err := s.S3.ListObjects(&s3.ListObjectsInput{
		Bucket:  aws.String(s.Config.Bucket),
		MaxKeys: aws.Int64(0),
	})

	return err == nil
}

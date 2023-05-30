// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Setup initializes the S3 bucket (life-cycle rules & CORS)
func (s *Server) Setup() error {
	if s.Config.Implementation == "" {
		s.Config.Implementation = s.DetectImplementation()
		log.Printf("Detected %s S3 implementation for server %s", s.Config.Implementation, s.GetURL())
	} else {
		log.Printf("Using %s S3 implementation for server %s", s.Config.Implementation, s.GetURL())
	}

	// MinIO does not support the setup of bucket CORS rules and MPU abortion lifecycle
	if s.Config.Implementation == ImplementationMinio {
		s.Config.Setup.CORS = false
		s.Config.Setup.AbortIncompleteUploads = 0
	}

	// Create bucket if it does not exist yet
	if _, err := s.GetBucketPolicy(&s3.GetBucketPolicyInput{
		Bucket: aws.String(s.Config.Bucket),
	}); err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeNoSuchBucket && s.Config.Setup.Bucket {
			if _, err := s.CreateBucket(&s3.CreateBucketInput{
				Bucket: aws.String(s.Config.Bucket),
			}); err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", *&s.Config.Bucket, err)
			}
		}
	}

	// Set CORS configuration for bucket
	if s.Config.Setup.CORS {
		corsRule := &s3.CORSRule{
			AllowedHeaders: aws.StringSlice([]string{"Authorization"}),
			AllowedOrigins: aws.StringSlice([]string{"*"}),
			MaxAgeSeconds:  aws.Int64(3000),
			AllowedMethods: aws.StringSlice([]string{"PUT", "GET"}),
			ExposeHeaders:  aws.StringSlice([]string{"ETag"}),
		}

		if _, err := s.PutBucketCors(&s3.PutBucketCorsInput{
			Bucket: aws.String(s.Config.Bucket),
			CORSConfiguration: &s3.CORSConfiguration{
				CORSRules: []*s3.CORSRule{
					corsRule,
				},
			},
		}); err != nil {
			return fmt.Errorf("failed to set bucket %s's CORS rules: %w", s.Config.Bucket, err)
		}
	}

	if s.Config.Setup.Lifecycle {
		// Create lifecycle policies
		lcRules := []*s3.LifecycleRule{}

		if s.Config.Setup.AbortIncompleteUploads > 0 {
			lcRules = append(lcRules, &s3.LifecycleRule{
				ID:     aws.String("Abort Multipart Uploads"),
				Status: aws.String("Enabled"),
				AbortIncompleteMultipartUpload: &s3.AbortIncompleteMultipartUpload{
					DaysAfterInitiation: aws.Int64(31),
				},
				Filter: &s3.LifecycleRuleFilter{
					Prefix: aws.String("/"),
				},
			})
		}

		for _, cls := range s.Config.Expiration {
			lcRules = append(lcRules, &s3.LifecycleRule{
				ID:     aws.String(fmt.Sprintf("Expiration after %s", cls.Title)),
				Status: aws.String("Enabled"),
				Filter: &s3.LifecycleRuleFilter{
					Tag: &s3.Tag{
						Key:   aws.String("expiration"),
						Value: aws.String(cls.ID),
					},
				},
				Expiration: &s3.LifecycleExpiration{
					Days: aws.Int64(cls.Days),
				},
			})
		}

		if len(lcRules) > 0 {
			if _, err := s.PutBucketLifecycleConfiguration(&s3.PutBucketLifecycleConfigurationInput{
				Bucket: aws.String(s.Config.Bucket),
				LifecycleConfiguration: &s3.BucketLifecycleConfiguration{
					Rules: lcRules,
				},
			}); err != nil {
				return fmt.Errorf("failed to set bucket %s's lifecycle rules: %w", s.Config.Bucket, err)
			}
		}
	}

	return nil
}

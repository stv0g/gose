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
	s.Implementation = s.DetectImplementation()
	log.Printf("Detected %s S3 implementation for server %s\n", s.Implementation, s.GetURL())

	// Create bucket if it does not exist yet
	if _, err := s.GetBucketPolicy(&s3.GetBucketPolicyInput{
		Bucket: aws.String(s.Config.Bucket),
	}); err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeNoSuchBucket && s.Config.CreateBucket {
			if _, err := s.CreateBucket(&s3.CreateBucketInput{
				Bucket: aws.String(s.Config.Bucket),
			}); err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", *&s.Config.Bucket, err)
			}
		}
	}

	// Set CORS configuration for bucket
	if s.Implementation != ImplementationMinio {
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

	// Create lifecycle policies
	lcRules := []*s3.LifecycleRule{}

	if s.Implementation != ImplementationMinio {
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

	// lc, err := svc.GetBucketLifecycleConfiguration(&s3.GetBucketLifecycleConfigurationInput{
	// 	Bucket: aws.String(.Bucket),
	// })
	// if err != nil {
	// 	return fmt.Errorf("failed get life-cycle rules: %w", err)
	// }
	// log.Printf("Life-cycle rules: %+#v\n", lc)

	return nil
}

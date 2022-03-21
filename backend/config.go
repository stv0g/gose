package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stv0g/gose/backend/config"
)

func getAnonymousReadPolicy(bucket string) string {
	policy := `
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "PublicRead",
            "Effect": "Allow",
            "Principal": "*",
            "Action": [
                "s3:GetObject",
                "s3:GetObjectVersion"
            ],
            "Resource": [
                "arn:aws:s3:::{{bucket}}/*"
            ]
        }
    ]
}
`

	return strings.ReplaceAll(policy, "{{bucket}}", bucket)
}

func configBucket(svc *s3.S3, cfg *config.Config) error {

	corsRule := &s3.CORSRule{
		AllowedHeaders: aws.StringSlice([]string{"Authorization"}),
		AllowedOrigins: aws.StringSlice([]string{"*"}),
		MaxAgeSeconds:  aws.Int64(3000),
		AllowedMethods: aws.StringSlice([]string{"PUT", "GET"}),
		ExposeHeaders:  aws.StringSlice([]string{"ETag"}),
	}

	if _, err := svc.PutBucketCors(&s3.PutBucketCorsInput{
		Bucket: aws.String(cfg.S3.Bucket),
		CORSConfiguration: &s3.CORSConfiguration{
			CORSRules: []*s3.CORSRule{
				corsRule,
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to set bucket %s's CORS rules: %w", cfg.S3.Bucket, err)
	}

	lcRules := []*s3.LifecycleRule{
		{
			ID:     aws.String("Abort Multipart Uploads"),
			Status: aws.String("Enabled"),
			AbortIncompleteMultipartUpload: &s3.AbortIncompleteMultipartUpload{
				DaysAfterInitiation: aws.Int64(31),
			},
			Filter: &s3.LifecycleRuleFilter{
				Prefix: aws.String("/"),
			},
		},
	}

	for _, cls := range cfg.S3.Expiration.Classes {
		lcRules = append(lcRules, &s3.LifecycleRule{
			ID:     aws.String(fmt.Sprintf("Expiration after %s", cls.Title)),
			Status: aws.String("Enabled"),
			Filter: &s3.LifecycleRuleFilter{
				Tag: &s3.Tag{
					Key:   aws.String("expiration"),
					Value: aws.String(cls.Tag),
				},
			},
			Expiration: &s3.LifecycleExpiration{
				Days: aws.Int64(cls.Days),
			},
		})
	}

	if _, err := svc.PutBucketLifecycleConfiguration(&s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(cfg.S3.Bucket),
		LifecycleConfiguration: &s3.BucketLifecycleConfiguration{
			Rules: lcRules,
		},
	}); err != nil {
		return fmt.Errorf("failed to set bucket %s's lifecycle rules: %w", cfg.S3.Bucket, err)
	}

	// lc, err := svc.GetBucketLifecycleConfiguration(&s3.GetBucketLifecycleConfigurationInput{
	// 	Bucket: aws.String(cfg.S3.Bucket),
	// })
	// if err != nil {
	// 	return fmt.Errorf("failed get life-cycle rules: %w", err)
	// }
	// log.Printf("Life-cycle rules: %+#v\n", lc)

	return nil
}

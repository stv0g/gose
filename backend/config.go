package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func filterMethods(methods []string) []string {
	filtered := make([]string, 0, len(methods))
	for _, m := range methods {
		v := strings.ToUpper(m)
		switch v {
		case "POST", "GET", "PUT", "PATCH", "DELETE":
			filtered = append(filtered, v)
		}
	}

	return filtered
}

func configBucket(svc *s3.S3, bucket *string) {
	methods := filterMethods(flag.Args())

	rule := s3.CORSRule{
		AllowedHeaders: aws.StringSlice([]string{"Authorization"}),
		AllowedOrigins: aws.StringSlice([]string{"*"}),
		MaxAgeSeconds:  aws.Int64(3000),

		// Add HTTP methods CORS request that were specified in the CLI.
		AllowedMethods: aws.StringSlice(methods),
	}

	params := s3.PutBucketCorsInput{
		Bucket: bucket,
		CORSConfiguration: &s3.CORSConfiguration{
			CORSRules: []*s3.CORSRule{&rule},
		},
	}

	_, err := svc.PutBucketCors(&params)
	if err != nil {
		// Print the error message
		fmt.Printf("Unable to set Bucket %q's CORS, %v\n", bucket, err)
	}

	// Print the updated CORS config for the bucket
	fmt.Printf("Updated bucket %q CORS for %v\n", bucket, methods)
}

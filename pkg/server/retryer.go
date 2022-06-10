package server

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
)

type retryer struct {
	client.DefaultRetryer
}

func (d retryer) RetryRules(r *request.Request) time.Duration {
	ra := r.HTTPResponse.Header.Get("Retry-after")
	svr := r.HTTPResponse.Header.Get("Server")

	// MinIO returns a "Retry-after: 120" header when the server is still initializing
	// Waiting 120 seconds is far too long as most servers finish their initialization faster
	// We reduce the time to 1 second and let the other retry rules handle the rest.
	if ra != "" && svr == "MinIO" && r.HTTPResponse.StatusCode == 503 {
		r.HTTPResponse.Header.Set("Retry-after", "1")
	}

	return d.DefaultRetryer.RetryRules(r)
}

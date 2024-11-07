// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stv0g/gose/pkg/config"
)

// List is a list of servers.
type List map[string]Server

// NewList creates a new server list.
func NewList(svrs []config.S3Server) List {
	svcs := List{}
	sess := session.Must(session.NewSession())

	for i := range svrs {
		svr := &svrs[i]

		svcs[svr.ID] = Server{
			S3: s3.New(sess, &aws.Config{
				Region:           aws.String(svr.Region),
				Endpoint:         aws.String(svr.Endpoint),
				S3ForcePathStyle: aws.Bool(svr.PathStyle),
				DisableSSL:       aws.Bool(svr.NoSSL),
				Credentials:      credentials.NewStaticCredentials(svr.AccessKey, svr.SecretKey, ""),
			}),
			Config: svr,
		}
	}

	return svcs
}

// Setup initialize all servers in the list.
func (sl List) Setup() error {
	for _, svc := range sl {
		if err := svc.Setup(); err != nil {
			return err
		}
	}

	return nil
}

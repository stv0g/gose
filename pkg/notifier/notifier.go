// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package notifier

import (
	"bytes"
	"fmt"
	"math"
	"net"
	"net/http"
	"regexp"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/router"
	"github.com/containrrr/shoutrrr/pkg/types"
	"github.com/stv0g/gose/pkg/utils"
)

type notifierArgs struct {
	URL              string
	FileSize         int64
	FileSizeHuman    string
	FileName         string
	FileType         string
	UploaderIP       string
	UploaderHostname string
	Env              map[string]string
	ExpiryRuleID     string
	ExpiryDate       time.Time
	UploadDate       time.Time
}

// Notifier sends notifications via various channels
type Notifier struct {
	*router.ServiceRouter

	template *template.Template
}

// NewNotifier creates a new notifier instance
func NewNotifier(tpl string, urls ...string) (*Notifier, error) {
	sender, err := shoutrrr.CreateSender(urls...)
	if err != nil {
		return nil, err
	}

	t := template.New("action")

	t, err = t.Parse(tpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse notification template: %w", err)
	}

	return &Notifier{
		ServiceRouter: sender,
		template:      t,
	}, nil
}

// Notify sends a notification
func (n *Notifier) Notify(url string, obj *s3.HeadObjectOutput, params types.Params) error {
	env, err := utils.EnvToMap()
	if err != nil {
		return fmt.Errorf("failed to get env: %w", err)
	}

	data := notifierArgs{
		FileName:      *obj.Metadata["Original-Filename"],
		FileSize:      *obj.ContentLength,
		FileSizeHuman: humanizeBytes(*obj.ContentLength),
		FileType:      *obj.ContentType,
		Env:           env,
		URL:           url,
		UploadDate:    *obj.LastModified,
	}

	if obj.Expiration != nil {
		re := regexp.MustCompile(`([a-z-]+)="([^"]+)"`)
		for _, m := range re.FindAllStringSubmatch(*obj.Expiration, -1) {
			switch m[1] {
			case "expiry-date":
				if expiryTime, err := http.ParseTime(m[2]); err == nil {
					data.ExpiryDate = expiryTime
				}

			case "rule-id":
				data.ExpiryRuleID = m[2]
			}
		}
	}

	if upl, ok := obj.Metadata["Original-Uploader"]; ok {
		data.UploaderIP = *upl

		if addrs, err := net.LookupAddr(data.UploaderIP); err != nil && len(addrs) > 0 {
			data.UploaderHostname = addrs[0]
		}
	}

	var tpl bytes.Buffer
	if err := n.template.Execute(&tpl, data); err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	msg := tpl.String()

	if errs := n.Send(msg, &params); errs != nil {
		for _, err := range errs {
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

func humanizeBytes(s int64) string {
	sizes := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	base := 1024.0

	if s < 10 {
		return fmt.Sprintf("%d B", s)
	}
	e := math.Floor(logn(float64(s), base))
	suffix := sizes[int(e)]
	val := math.Floor(float64(s)/math.Pow(base, e)*10+0.5) / 10
	f := "%.0f %s"
	if val < 10 {
		f = "%.1f %s"
	}

	return fmt.Sprintf(f, val, suffix)
}

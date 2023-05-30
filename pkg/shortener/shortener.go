// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package shortener

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/stv0g/gose/pkg/config"
	"github.com/stv0g/gose/pkg/utils"
)

// Shortener is an URL shortener instance
type Shortener struct {
	config.ShortenerConfig
}

type shortenerArgs struct {
	URL        string
	URLEscaped string
	Env        map[string]string
}

// NewShortener creates a new URL shortener instance
func NewShortener(c *config.ShortenerConfig) (*Shortener, error) {
	s := new(Shortener)

	s.ShortenerConfig = *c

	return s, nil
}

func (s *Shortener) getRequest(u string) (*http.Request, error) {
	t := template.New("action")

	var err error
	t, err = t.Parse(s.Endpoint)
	if err != nil {
		return &http.Request{}, err
	}

	env, err := utils.EnvToMap()
	if err != nil {
		return nil, fmt.Errorf("failed to get env: %w", err)
	}

	data := shortenerArgs{
		URL:        u,
		URLEscaped: url.QueryEscape(u),
		Env:        env,
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		return &http.Request{}, err
	}

	tplURL := tpl.String()

	return http.NewRequest(s.Method, tplURL, nil)
}

// Shorten shorten a passed long URL into a short one using the shortener service
func (s *Shortener) Shorten(long *url.URL) (*url.URL, error) {
	req, err := s.getRequest(long.String())
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: time.Second * 10}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid API response: %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var shortURL *url.URL

	switch s.Response {
	case "raw":
		shortURL, err = url.Parse(string(body))
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("Unknown shortener response type: %s", s.Response)
	}

	return shortURL, nil
}

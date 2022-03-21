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

type Shortener struct {
	config.ShortenerConfig
}

type ShortenerArgs struct {
	Url        string
	UrlEscaped string
	Env        map[string]string
}

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

	data := ShortenerArgs{
		Url:        u,
		UrlEscaped: url.QueryEscape(u),
		Env:        env,
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		return &http.Request{}, err
	}

	tUrl := tpl.String()

	return http.NewRequest(s.Method, tUrl, nil)
}

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

	var shortUrl *url.URL

	switch s.Response {
	case "raw":
		shortUrl, err = url.Parse(string(body))
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("Unknown shortener response type: %s", s.Response)
	}

	return shortUrl, nil
}

package shortener

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/stv0g/gose/backend/config"
)

type Shortener struct {
	config.ShortenerConfig
}

type ShortenerArgs struct {
	Url string
}

func NewShortener(c config.ShortenerConfig) *Shortener {
	s := new(Shortener)

	s.ShortenerConfig = c

	return s
}

func (s *Shortener) getRequest(url string) (*http.Request, error) {
	t := template.New("action")

	var err error
	t, err = t.Parse(s.Endpoint)
	if err != nil {
		return &http.Request{}, err
	}

	data := ShortenerArgs{
		Url: url,
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		return &http.Request{}, err
	}

	tUrl := tpl.String()

	return http.NewRequest(s.Method, tUrl, nil)
}

func (s *Shortener) Shorten(longUrl string) (*url.URL, error) {
	req, err := s.getRequest(longUrl)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: time.Second * 10}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid API response")
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

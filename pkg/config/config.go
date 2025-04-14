// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"flag"
	"fmt"
	"log"
	"strings"

	units "github.com/docker/go-units"
	"github.com/mitchellh/mapstructure"
	"github.com/mozillazg/go-slugify"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	// MinPartSize is the minimally supported part size for an S3 multi-part upload part.
	MinPartSize size = 5 << 20 // 5MiB

	// DefaultPartSize is the default size of the chunks used for Multi-part Upload if not provided by the configuration.
	DefaultPartSize size = 16 << 20 // 16MiB

	// DefaultMaxUploadSize is the maximum upload size if not provided by the configuration.
	DefaultMaxUploadSize size = 1 << 40 // 1TiB

	// DefaultRegion is the default S3 region if not provided by the configuration.
	DefaultRegion = "us-east-1"

	// DefaultBucket is the default S3 bucket name to use if not provided by the configuration.
	DefaultBucket = "gose-uploads"
)

// DefaultExpiration is list of default expiration classes.
var DefaultExpiration = []Expiration{
	{
		ID:    "1day",
		Title: "1day",
		Days:  1,
	},
	{
		ID:    "1week",
		Title: "1 week",
		Days:  7,
	},
	{
		ID:    "1month",
		Title: "1 month",
		Days:  31,
	},
	{
		ID:    "1year",
		Title: "1 year",
		Days:  365,
	},
}

type size int64

func (s *size) UnmarshalText(text []byte) error {
	sz, err := units.FromHumanSize(string(text))
	if err != nil {
		return err
	}

	*s = size(sz)
	return nil
}

// Expiration describes how long files are kept before getting deleted.
type Expiration struct {
	ID    string `json:"id" yaml:"id"`
	Title string `json:"title" yaml:"title"`

	Days int64 `json:"days" yaml:"days"`
}

// S3ServerConfig is the public part of S3Server.
type S3ServerConfig struct {
	ID    string `json:"id" yaml:"id"`
	Title string `json:"title" yaml:"title"`

	Implementation string       `json:"implementation" yaml:"implementation"`
	MaxUploadSize  size         `json:"max_upload_size" yaml:"max_upload_size"`
	PartSize       size         `json:"part_size" yaml:"part_size"`
	Expiration     []Expiration `json:"expiration" yaml:"expiration"`
}

// S3ServerSetup describes initial configuration for an S3 server/bucket.
type S3ServerSetup struct {
	Bucket                 bool `json:"bucket" yaml:"bucket"`
	CORS                   bool `json:"cors" yaml:"cors"`
	Lifecycle              bool `json:"lifecycle" yaml:"lifecycle"`
	AbortIncompleteUploads int  `json:"abort_incomplete_uploads" yaml:"abort_incomplete_uploads"`
}

// S3Server describes an S3 server
type S3Server struct {
	// S3ServerConfig is the public info about an S3 server shared with the frontend.
	S3ServerConfig `json:",squash"`

	Endpoint  string `json:"endpoint" yaml:"endpoint"`
	Bucket    string `json:"bucket" yaml:"bucket"`
	Region    string `json:"region" yaml:"region"`
	PathStyle bool   `json:"path_style" yaml:"path_style"`
	NoSSL     bool   `json:"no_ssl" yaml:"no_ssl"`
	AccessKey string `json:"access_key" yaml:"access_key"`
	SecretKey string `json:"secret_key" yaml:"secret_key"`

	Setup S3ServerSetup `json:"setup" yaml:"setup"`
}

// ShortenerConfig contains Link-shortener specific configuration.
type ShortenerConfig struct {
	Endpoint string `json:"endpoint" yaml:"endpoint"`
	Method   string `json:"method" yaml:"method"`
	Response string `json:"response" yaml:"response"`
}

// NotificationConfig contains notification specific configuration.
type NotificationConfig struct {
	URLs     []string `json:"urls" yaml:"urls"`
	Template string   `json:"template" yaml:"template"`

	Uploads   bool `json:"uploads" yaml:"uploads"`
	Downloads bool `json:"downloads" yaml:"downloads"`

	Mail *struct {
		URL      string `json:"url" yaml:"url"`
		Template string `json:"template" yaml:"template"`
	} `json:"mail" yaml:"mail"`
}

// Config contains the main configuration.
type Config struct {
	*viper.Viper `json:"-" yaml:"-"`

	// Default or single server config values.
	S3Server `json:",squash" yaml:"default"`

	// Multiple server config values.
	Servers []S3Server `json:"servers" yaml:"servers,omitempty"`

	// Host is the local machine IP Address to bind the HTTP Server to.
	Listen string `json:"listen" yaml:"listen,omitempty"`

	// Directory of frontend assets if not bundled.
	Static string `json:"static" yaml:"static,omitempty"`

	// BaseURL at which Gose is accessible.
	BaseURL string `json:"base_url" yaml:"base_url,omitempty"`

	Shortener    *ShortenerConfig    `json:"shortener" yaml:"shortener,omitempty"`
	Notification *NotificationConfig `json:"notification" yaml:"notification,omitempty"`
}

// NewConfig returns a new decoded Config struct.
func NewConfig(configFile string) (*Config, error) {
	// Create config structure.
	cfg := &Config{
		Viper: viper.New(),
	}

	cfg.SetDefault("listen", ":8080")
	cfg.SetDefault("static", "./dist")
	cfg.SetDefault("base_url", "http://localhost:8080")
	cfg.SetDefault("notification.uploads", true)
	cfg.SetDefault("notification.downloads", false)
	cfg.SetDefault("max_upload_size", DefaultMaxUploadSize)
	cfg.SetDefault("part_size", DefaultPartSize)
	cfg.SetDefault("expiration", DefaultExpiration)
	cfg.SetDefault("endpoint", "")
	cfg.SetDefault("bucket", DefaultBucket)
	cfg.SetDefault("region", DefaultRegion)
	cfg.SetDefault("path_style", false)
	cfg.SetDefault("no_ssl", false)
	cfg.SetDefault("access_key", "")
	cfg.SetDefault("secret_key", "")
	cfg.SetDefault("implementation", "")
	cfg.SetDefault("setup.bucket", true)
	cfg.SetDefault("setup.cors", true)
	cfg.SetDefault("setup.lifecycle", true)
	cfg.SetDefault("setup.abort_incomplete_uploads", 31)

	cfg.BindEnv("access_key", "AWS_ACCESS_KEY_ID")
	cfg.BindEnv("secret_key", "AWS_SECRET_ACCESS_KEY")

	replacer := strings.NewReplacer(".", "_")
	cfg.SetEnvKeyReplacer(replacer)
	cfg.SetEnvPrefix("gose")
	cfg.AutomaticEnv()

	if configFile != "" {
		cfg.SetConfigFile(configFile)

		if err := cfg.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	if err := cfg.UnmarshalExact(cfg, viper.DecodeHook(func(c *mapstructure.DecoderConfig) {
		c.DecodeHook = mapstructure.TextUnmarshallerHookFunc()
		c.TagName = "json"
	})); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Use the default values as the single server if no others are configured.
	if len(cfg.Servers) == 0 {
		cfg.Servers = append(cfg.Servers, cfg.S3Server)
	}

	// Some normalization and default values for servers.
	for i := range cfg.Servers {
		svr := &cfg.Servers[i]

		if svr.ID == "" {
			svr.ID = slugify.Slugify(svr.Endpoint)
		}

		if svr.Title == "" {
			svr.Title = svr.Endpoint
		}

		if svr.Region == "" {
			svr.Region = cfg.Region
		}

		if svr.MaxUploadSize == 0 {
			svr.MaxUploadSize = cfg.MaxUploadSize
		}

		if svr.PartSize == 0 {
			svr.PartSize = cfg.PartSize
		}

		if svr.AccessKey == "" {
			svr.AccessKey = cfg.AccessKey
		}

		if svr.SecretKey == "" {
			svr.SecretKey = cfg.SecretKey
		}

		if svr.Expiration == nil {
			svr.Expiration = []Expiration{}
		}
	}

	if err := cfg.Check(); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	log.Printf("Loaded configuration:\n")
	bs, _ := yaml.Marshal(cfg)
	fmt.Print(string(bs))

	return cfg, nil
}

func (c *Config) Check() error {
	for _, svr := range c.Servers {
		if svr.PartSize < MinPartSize {
			return fmt.Errorf("part_size must be larger than %s (it is currently %s)",
				units.HumanSize(float64(MinPartSize)),
				units.HumanSize(float64(svr.PartSize)))
		}
	}

	return nil
}

// ParseFlags will create and parse the CLI flags
// and return the path to be used elsewhere.
func ParseFlags() (string, bool, error) {
	// String that contains the configured configuration path.
	var configPath string
	var showVersion bool

	// Set up a CLI flag called "-config" to allow users
	// to supply the configuration file.
	flag.StringVar(&configPath, "config", "", "path to config file")

	// Set up a CLI flag called "-version" to print the programs version and build details.
	flag.BoolVar(&showVersion, "version", false, "show version information")

	// Actually parse the flags.
	flag.Parse()

	// Return the configuration path.
	return configPath, showVersion, nil
}

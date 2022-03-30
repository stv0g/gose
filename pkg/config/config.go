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
	"gopkg.in/yaml.v2"
)

const (
	DefaultPartSize      size = 16e6 // 16MB
	DefaultMaxUploadSize size = 1e12 // 1TB
	DefaultRegion             = "us-east-1"
)

type size int64

func (s *size) UnmarshalText(text []byte) error {
	sz, err := units.FromHumanSize(string(text))
	if err != nil {
		return err
	}

	*s = size(sz)
	return nil
}

// Expiration describes how long files are kept before getting deleted
type Expiration struct {
	ID    string `mapstructure:"id" json:"id"`
	Title string `mapstructure:"title" json:"title"`

	Days int64 `mapstructure:"days" json:"days"`
}

// S3ServerConfig is the public part of S3Server
type S3ServerConfig struct {
	ID    string `mapstructure:"id" json:"id"`
	Title string `mapstructure:"title" json:"title"`

	Expiration []Expiration `mapstructure:"expiration" json:"expiration"`
}

// S3Server describes an S3 server
type S3Server struct {
	S3ServerConfig `mapstructure:",squash"`

	Endpoint  string `mapstructure:"endpoint"`
	Bucket    string `mapstructure:"bucket"`
	Region    string `mapstructure:"region"`
	PathStyle bool   `mapstructure:"path_style"`
	NoSSL     bool   `mapstructure:"no_ssl"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`

	MaxUploadSize size `mapstructure:"max_upload_size"`
	PartSize      size `mapstructure:"part_size"`
}

// ShortenerConfig contains Link-shortener specific configuration
type ShortenerConfig struct {
	Endpoint string `mapstructure:"endpoint"`
	Method   string `mapstructure:"method"`
	Response string `mapstructure:"response"`
}

// NotificationConfig contains notification specific configuration
type NotificationConfig struct {
	URLs     []string `mapstructure:"urls"`
	Template string   `mapstructure:"template"`

	Uploads   bool `mapstructure:"uploads"`
	Downloads bool `mapstructure:"downloads"`

	Mail *struct {
		URL      string `mapstructure:"url"`
		Template string `mapstructure:"template"`
	} `mapstructure:"mail"`
}

// Config contains the main configuration
type Config struct {
	S3Server `mapstructure:",squash"`

	*viper.Viper `mapstructure:"-"`

	// Host is the local machine IP Address to bind the HTTP Server to
	Listen string `mapstructure:"listen"`

	// Directory of frontend assets if not bundled
	Static string `mapstructure:"static"`

	// BaseURL at which Gose is accessible
	BaseURL string `mapstructure:"base_url"`

	Servers      []S3Server          `mapstructure:"servers"`
	Shortener    *ShortenerConfig    `mapstructure:"shortener"`
	Notification *NotificationConfig `mapstructure:"notification"`
}

// NewConfig returns a new decoded Config struct
func NewConfig(configFile string) (*Config, error) {
	// Create cfg structure
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
	cfg.SetDefault("region", DefaultRegion)

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

	if err := cfg.UnmarshalExact(cfg, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc())); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Some normalization and default values for servers
	for i, _ := range cfg.Servers {
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

	log.Printf("Loaded configuration:\n")
	bs, _ := yaml.Marshal(cfg)
	fmt.Print(string(bs))

	return cfg, nil
}

// ParseFlags will create and parse the CLI flags
// and return the path to be used elsewhere
func ParseFlags() (string, error) {
	// String that contains the configured configuration path
	var configPath string

	// Set up a CLI flag called "-config" to allow users
	// to supply the configuration file
	flag.StringVar(&configPath, "config", "", "path to config file")

	// Actually parse the flags
	flag.Parse()

	// Return the configuration path
	return configPath, nil
}

package config

import (
	"flag"
	"fmt"
	"net/url"
	"strings"

	units "github.com/docker/go-units"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type Size int64

func (s *Size) UnmarshalText(text []byte) error {
	if sz, err := units.FromHumanSize(string(text)); err != nil {
		return err
	} else {
		*s = (Size)(sz)
		return nil
	}
}

type ExpirationClass struct {
	Tag   string `mapstructure:"tag"`
	Days  int64  `mapstructure:"days"`
	Title string `mapstructure:"title"`
}

type S3Config struct {
	Endpoint  string `mapstructure:"endpoint"`
	Bucket    string `mapstructure:"bucket"`
	Region    string `mapstructure:"region"`
	PathStyle bool   `mapstructure:"path_style"`
	NoSSL     bool   `mapstructure:"no_ssl"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`

	MaxUploadSize Size `mapstructure:"max_upload_size"`
	PartSize      Size `mapstructure:"part_size"`

	Expiration struct {
		Default string            `mapstructure:"default_class"`
		Classes []ExpirationClass `mapstructure:"classes"`
	} `mapstructure:"expiration"`
}

type ServerConfig struct {
	// Host is the local machine IP Address to bind the HTTP Server to
	Listen string `mapstructure:"listen"`

	Static string `mapstructure:"static"`
}

type ShortenerConfig struct {
	Endpoint string `mapstructure:"endpoint"`
	Method   string `mapstructure:"method"`
	Response string `mapstructure:"response"`
}

type NotificationConfig struct {
	URLs     []string `mapstructure:"urls"`
	Template string   `mapstructure:"template"`
}

type Config struct {
	*viper.Viper `mapstructure:"-"`

	S3           *S3Config           `mapstructure:"s3"`
	Server       *ServerConfig       `mapstructure:"server"`
	Shortener    *ShortenerConfig    `mapstructure:"shortener"`
	Notification *NotificationConfig `mapstructure:"notification"`
}

func (c *S3Config) GetUrl() *url.URL {
	u := &url.URL{}

	if c.NoSSL {
		u.Scheme = "http"
	} else {
		u.Scheme = "https"
	}

	if c.PathStyle {
		u.Host = c.Endpoint
		u.Path = "/" + c.Bucket
	} else {
		u.Host = c.Bucket + "." + c.Endpoint
		u.Path = ""
	}

	return u
}

func (c *S3Config) GetObjectUrl(key string) *url.URL {
	u := c.GetUrl()
	u.Path += "/" + key

	return u
}

// NewConfig returns a new decoded Config struct
func NewConfig(configFile string) (*Config, error) {
	// Create cfg structure
	cfg := &Config{
		Viper: viper.New(),
	}

	cfg.SetDefault("s3.max_upload_size", "1TB")
	cfg.SetDefault("s3.part_size", "5MB")
	cfg.SetDefault("server.listen", ":8080")
	cfg.SetDefault("server.static", "./dist")

	replacer := strings.NewReplacer(".", "_")
	cfg.SetEnvKeyReplacer(replacer)
	cfg.SetEnvPrefix("gose")
	cfg.AutomaticEnv()

	cfg.BindEnv("s3.access_key", "AWS_ACCESS_KEY_ID", "MINIO_ACCESS_KEY")
	cfg.BindEnv("s3.secret_key", "AWS_SECRET_ACCESS_KEY", "MINIO_SECRET_KEY")

	cfg.SetConfigFile(configFile)

	if err := cfg.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := cfg.UnmarshalExact(cfg, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc())); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

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
	flag.StringVar(&configPath, "config", "./config.yaml", "path to config file")

	// Actually parse the flags
	flag.Parse()

	// Return the configuration path
	return configPath, nil
}
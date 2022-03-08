package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

type MpuConfig struct {
	Enabled       bool `yaml:"enabled"`
	ThresholdSize int  `yaml:"threshold_size"`
}

type S3Config struct {
	Endpoint  string `yaml:"endpoint"`
	Bucket    string `yaml:"bucket"`
	Region    string `yaml:"region"`
	PathStyle bool   `yaml:"path_style"`
	NoSSL     bool   `yaml:"no_ssl"`
}

type ServerConfig struct {
	// Host is the local machine IP Address to bind the HTTP Server to
	Bind string `yaml:"bind"`
}

type ShortenerConfig struct {
	Endpoint string `yaml:"endpoint"`
	Method   string `yaml:"method"`
	Response string `yaml:"response"`
}

type Config struct {
	Mpu       MpuConfig       `yaml:"mpu"`
	S3        S3Config        `yaml:"s3"`
	Server    ServerConfig    `yaml:"server"`
	Shortener ShortenerConfig `yaml:"shortener"`
}

func (c *S3Config) GetUrl() string {
	url := ""

	if c.NoSSL {
		url += "http://"
	} else {
		url += "https://"
	}

	if c.PathStyle {
		url += c.Endpoint + "/" + c.Bucket
	} else {
		url += c.Bucket + "." + c.Endpoint
	}

	return url
}

func (c *S3Config) GetObjectUrl(u *uuid.UUID, key string) string {
	url := c.GetUrl()

	url += u.String() + "/" + key

	return url
}

// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{
		Mpu: MpuConfig{
			Enabled: false,
		},
		Server: ServerConfig{
			Bind: ":8080",
		},
	}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read
func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
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

	// Validate the path first
	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}

	// Return the configuration path
	return configPath, nil
}

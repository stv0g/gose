package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/stv0g/gose/pkg/config"
	"github.com/stv0g/gose/pkg/handlers"
	"github.com/stv0g/gose/pkg/notifier"

	"github.com/stv0g/gose/pkg/shortener"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

const apiBase = "/api/v1"

func main() {
	// Generate our config based on the config supplied
	// by the user in the flags
	cfgFile, err := config.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.NewConfig(cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	// Run the server
	run(cfg)
}

// APIMiddleware will add the db connection to the context
func APIMiddleware(svc *s3.S3, shortener *shortener.Shortener, cfg *config.Config, notifier *notifier.Notifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("s3", svc)
		c.Set("cfg", cfg)
		c.Set("shortener", shortener)
		c.Set("notifier", notifier)
		c.Next()
	}
}

func run(cfg *config.Config) {
	var err error

	sess := session.Must(session.NewSession())
	svc := s3.New(sess, &aws.Config{
		Region:           aws.String(cfg.S3.Region),
		Endpoint:         &cfg.S3.Endpoint,
		S3ForcePathStyle: &cfg.S3.PathStyle,
		DisableSSL:       &cfg.S3.NoSSL,
		Credentials:      credentials.NewStaticCredentials(cfg.S3.AccessKey, cfg.S3.SecretKey, ""),
	})

	// if err := configBucket(svc, cfg); err != nil {
	// 	log.Fatalf("Failed to setup bucket: %s", err)
	// }

	var short *shortener.Shortener
	if cfg.Shortener != nil {
		if short, err = shortener.NewShortener(cfg.Shortener); err != nil {
			log.Fatalf("Failed to create link shortener: %s", err)
		}
	}

	var notif *notifier.Notifier
	if cfg.Notification != nil {
		if notif, err = notifier.NewNotifier(cfg.Notification); err != nil {
			log.Fatalf("Failed to create notification sender: %s", err)
		}
	}

	router := gin.Default()
	router.Use(APIMiddleware(svc, short, cfg, notif))
	router.Use(StaticMiddleware(cfg))

	router.GET(apiBase+"/config", handlers.HandleConfig)
	router.POST(apiBase+"/initiate", handlers.HandleInitiate)
	router.POST(apiBase+"/complete", handlers.HandleComplete)

	server := &http.Server{
		Addr:           cfg.Server.Listen,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("GoSƐ %s, commit %s, built at %s by %s", version, commit, date, builtBy)
	log.Printf("Listening on: http://%s", server.Addr)

	server.ListenAndServe()
}

func exitError(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}

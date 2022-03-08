package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/stv0g/gose/backend/config"
	_ "github.com/stv0g/gose/backend/docs"
	"github.com/stv0g/gose/backend/handlers"

	"github.com/stv0g/gose/backend/shortener"
)

// @title Gose API
// @version 1.0
// @description A terascale uploader

// @contact.name Steffen Vogel
// @contact.email post@steffenvogel.de

// @license.name GPL v3.0
// @license.url https://www.gnu.org/licenses/gpl-3.0.en.html

// @host gose.0l.de
// @BasePath /api/v1
func main() {
	// Generate our config based on the config supplied
	// by the user in the flags
	cfgPath, err := config.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	// Run the server
	run(cfg)
}

// ApiMiddleware will add the db connection to the context
func ApiMiddleware(svc *s3.S3, shortener *shortener.Shortener, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("s3svc", svc)
		c.Set("cfg", cfg)
		c.Set("shortener", shortener)
		c.Next()
	}
}

func run(cfg *config.Config) {
	// Create a AWS SDK for Go Session that will load credentials using the SDK's
	// default credential change.
	sess := session.Must(session.NewSession())

	// Create a new S3 service client that will be use by the service to generate
	// presigned URLs with. Not actual API requests will be made with this client.
	// The credentials loaded when the Session was created above will be used
	// to sign the requests with.
	svc := s3.New(sess, &aws.Config{
		Region:           aws.String(cfg.S3.Region),
		Endpoint:         &cfg.S3.Endpoint,
		S3ForcePathStyle: &cfg.S3.PathStyle,
		DisableSSL:       &cfg.S3.NoSSL,
	})

	const apiBase = "/api/v1"

	short := shortener.NewShortener(cfg.Shortener)

	router := gin.Default()
	router.Use(ApiMiddleware(svc, short, cfg))

	router.Use(static.Serve("/", static.LocalFile("./dist", false)))

	url := ginSwagger.URL("http://localhost:8080/swagger/doc.json") // The url pointing to API definition
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	router.GET(apiBase+"/mpu/initiate/*key", handlers.HandleMPU)
	router.GET(apiBase+"/mpu/complete/*key", handlers.HandleMPU)
	router.GET(apiBase+"/presign/*key", handlers.HandlePresign)
	router.POST(apiBase+"/shorten/*key", handlers.HandleShorten)

	server := &http.Server{
		Addr:           cfg.Server.Bind,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Println("Starting Server nn:", "http://"+server.Addr)

	server.ListenAndServe()
}

func exitError(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}

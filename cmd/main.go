// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/stv0g/gose/pkg/config"
	"github.com/stv0g/gose/pkg/handlers"
	"github.com/stv0g/gose/pkg/server"

	"github.com/stv0g/gose/pkg/shortener"
)

var (
	version = "0.0.0"
	commit  = "unknown"
	date    = "unknown"
	builtBy = "unknown"
)

const apiBase = "/api/v1"

func main() {
	log.Printf("GoSƐ %s, commit %s, built at %s by %s\n", version, commit, date, builtBy)

	// Generate our config based on the config supplied
	// by the user in the flags.
	cfgFile, showVersion, err := config.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	if showVersion {
		return
	}

	cfg, err := config.NewConfig(cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	// Run the server.
	run(cfg)
}

// APIMiddleware will add the db connection to the context.
func APIMiddleware(svrs server.List, shortener *shortener.Shortener, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("servers", svrs)
		c.Set("config", cfg)
		c.Set("shortener", shortener)
		c.Next()
	}
}

func run(cfg *config.Config) {
	svrs := server.NewList(cfg.Servers)

	log.Printf("Initializing S3 servers. Please wait...")
	if err := svrs.Setup(); err != nil {
		log.Fatalf("Failed to setup servers: %s", err)
	}
	log.Printf("Initialization of %d servers completed.", len(svrs))

	var err error
	var short *shortener.Shortener
	if cfg.Shortener != nil {
		if short, err = shortener.NewShortener(cfg.Shortener); err != nil {
			log.Fatalf("Failed to create link shortener: %s", err)
		}
	}

	router := gin.Default()
	router.Use(APIMiddleware(svrs, short, cfg))
	router.Use(StaticMiddleware(cfg))

	router.GET(apiBase+"/config", handlers.HandleConfigWith(version, commit, date))
	router.GET(apiBase+"/healthz", handlers.HandleHealthz)
	router.POST(apiBase+"/initiate", handlers.HandleInitiate)
	router.POST(apiBase+"/part", handlers.HandlePart)
	router.POST(apiBase+"/complete", handlers.HandleComplete)
	router.GET(apiBase+"/download/:server/:etag/:filename", handlers.HandleDownload)
	router.HEAD(apiBase+"/download/:server/:etag/:filename", handlers.HandleDownload)

	server := &http.Server{
		Addr:           cfg.Listen,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Listening on: http://%s", server.Addr)

	server.ListenAndServe()
}

func exitError(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}

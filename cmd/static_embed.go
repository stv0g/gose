//go:build embed

package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/stv0g/gose/frontend"
	"github.com/stv0g/gose/pkg/config"
)

type embedFileSystem struct {
	http.FileSystem
}

func (e embedFileSystem) Exists(prefix string, path string) bool {
	_, err := e.Open(path)
	if err != nil {
		return false
	}
	return true
}

func embedFolder(fsEmbed embed.FS, targetPath string) static.ServeFileSystem {
	fsys, err := fs.Sub(fsEmbed, targetPath)
	if err != nil {
		panic(err)
	}
	return embedFileSystem{
		FileSystem: http.FS(fsys),
	}
}

// StaticMiddleware serves static assets package by Webpack
func StaticMiddleware(cfg *config.Config) gin.HandlerFunc {
	return static.Serve("/", embedFolder(frontend.Files, "dist"))
}

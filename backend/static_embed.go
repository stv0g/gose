//go:build embed

//go:generate npm --prefix ../frontend install
//go:generate npm --prefix ../frontend run-script build -- --output-path=../backend/dist/

package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/stv0g/gose/backend/config"
)

//go:embed dist
var embeddedFiles embed.FS

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

func EmbedFolder(fsEmbed embed.FS, targetPath string) static.ServeFileSystem {
	fsys, err := fs.Sub(fsEmbed, targetPath)
	if err != nil {
		panic(err)
	}
	return embedFileSystem{
		FileSystem: http.FS(fsys),
	}
}

func StaticMiddleware(cfg *config.Config) gin.HandlerFunc {
	return static.Serve("/", EmbedFolder(embeddedFiles, "dist"))
}

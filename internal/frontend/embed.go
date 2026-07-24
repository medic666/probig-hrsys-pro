package frontend

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist/*
var embeddedFiles embed.FS

func GetFileSystem() http.FileSystem {
	sub, err := fs.Sub(embeddedFiles, "dist")
	if err != nil {
		panic("failed to get embedded frontend: " + err.Error())
	}
	return http.FS(sub)
}

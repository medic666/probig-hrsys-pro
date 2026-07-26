package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed dist/*
var distFS embed.FS

func serveStatic(r *gin.Engine) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Println("WARNING: embedded dist not available")
		return
	}

	entries, err := distFS.ReadDir("dist")
	if err != nil || len(entries) <= 1 {
		log.Println("INFO: frontend not embedded, using dev mode")
		return
	}

	indexData, indexErr := fs.ReadFile(sub, "index.html")
	if indexErr != nil {
		log.Println("WARNING: index.html not found in embed")
		return
	}

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "接口不存在"})
			return
		}

		if strings.HasPrefix(path, "/assets/") {
			filePath := strings.TrimPrefix(path, "/assets/")
			data, err := fs.ReadFile(sub, filepath.Join("assets", filePath))
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}
			contentType := mimeType(filePath)
			c.Data(http.StatusOK, contentType, data)
			return
		}

		if path == "/favicon.ico" {
			data, err := fs.ReadFile(sub, "favicon.ico")
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}
			c.Data(http.StatusOK, "image/x-icon", data)
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", indexData)
	})
}

func mimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".js":
		return "application/javascript"
	case ".css":
		return "text/css"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".svg":
		return "image/svg+xml"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".json":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}

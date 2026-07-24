package embed

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed dist
var embeddedFiles embed.FS

func ServeFrontend(r *gin.Engine) {
	distFS, err := fs.Sub(embeddedFiles, "dist")
	if err != nil {
		panic("failed to load embedded frontend: " + err.Error())
	}

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "not found", "data": nil})
			return
		}
		fileServer := http.FileServer(http.FS(distFS))
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}

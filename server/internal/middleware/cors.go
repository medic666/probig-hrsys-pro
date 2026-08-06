package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS 可配置跨域支持：allowedOrigins 为空则不启用（同源部署零影响）
func CORS(allowedOrigins []string) gin.HandlerFunc {
	origins := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		if o = strings.TrimSpace(o); o != "" {
			origins[o] = true
		}
	}
	return func(c *gin.Context) {
		if len(origins) > 0 {
			origin := c.GetHeader("Origin")
			if origins[origin] {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
				// 暴露滑动续期响应头，跨域场景下前端才能读到
				c.Header("Access-Control-Expose-Headers", "X-New-Token")
			}
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}
		}
		c.Next()
	}
}

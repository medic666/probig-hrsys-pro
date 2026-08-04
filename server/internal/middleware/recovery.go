package middleware

import (
	"net"
	"os"
	"runtime/debug"
	"strings"

	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				stack := string(debug.Stack())
				utils.Errorf("panic recovered: %v\n%s", err, stack)

				if brokenPipe {
					c.Error(err.(error))
					c.Abort()
					return
				}

				utils.Error(c, "服务器内部错误")
				c.Abort()
			}
		}()
		c.Next()
	}
}

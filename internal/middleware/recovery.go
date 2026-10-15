package middleware

import (
	"runtime/debug"

	"tuyou/pkg/logger"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery Panic 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack := debug.Stack()

				// 记录错误日志
				path := c.FullPath()
				if path == "" {
					path = "<unmatched>"
				}
				logger.Error("Panic recovered",
					zap.String("stack", string(stack)),
					zap.String("path", path),
					zap.String("method", c.Request.Method),
				)

				// 响应客户端
				response.InternalError(c, "服务器内部错误")
				c.Abort()
			}
		}()

		c.Next()
	}
}

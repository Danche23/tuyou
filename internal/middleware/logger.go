package middleware

import (
	"time"

	"tuyou/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// writeRequestLog 允许测试注入内存日志，生产环境写入统一日志系统。
var writeRequestLog = logger.Info

// Logger 仅记录请求元信息。查询参数、请求体和原始 URL 都可能包含敏感数据。
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = "<unmatched>"
		}
		writeRequestLog("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}

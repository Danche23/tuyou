package middleware

import "github.com/gin-gonic/gin"

// RequestLogger 保留旧调用入口，统一复用不读取请求体的安全日志实现。
func RequestLogger() gin.HandlerFunc {
	return Logger()
}

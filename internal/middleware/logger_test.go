package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestRequestLogDoesNotExposePrivateData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name        string
		middleware  gin.HandlerFunc
		contentType string
		body        string
	}{
		{name: "standard-json", middleware: Logger(), contentType: "application/json", body: `{"password":"body-secret","code":"body-code"}`},
		{name: "legacy-json", middleware: RequestLogger(), contentType: "application/json", body: `{"password":"body-secret","code":"body-code"}`},
		{name: "multipart-with-boundary", middleware: RequestLogger(), contentType: "multipart/form-data; boundary=test-boundary", body: "file-content-secret"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(&output), zap.InfoLevel)
			previous := writeRequestLog
			writeRequestLog = func(message string, fields ...zap.Field) {
				zap.New(core).Info(message, fields...)
			}
			t.Cleanup(func() { writeRequestLog = previous })
			engine := gin.New()
			engine.Use(tc.middleware)
			engine.POST("/items/:id", func(c *gin.Context) {
				got, err := io.ReadAll(c.Request.Body)
				if err != nil || string(got) != tc.body {
					t.Errorf("处理函数未收到完整请求体: got=%q err=%v", got, err)
				}
				c.Status(http.StatusNoContent)
			})
			req := httptest.NewRequest(http.MethodPost, "/items/path-secret?token=query-secret", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", tc.contentType)
			engine.ServeHTTP(httptest.NewRecorder(), req)
			logData := output.String()
			for _, secret := range []string{"body-secret", "body-code", "file-content-secret", "query-secret", "path-secret"} {
				if strings.Contains(string(logData), secret) {
					t.Errorf("日志包含敏感请求内容 %q", secret)
				}
			}
			if !strings.Contains(string(logData), "/items/:id") || !strings.Contains(string(logData), "status") {
				t.Errorf("日志应保留路由模板及状态，实际: %s", logData)
			}
		})
	}
}

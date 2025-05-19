package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"gil_teacher/app/core/logger"

	"github.com/gin-gonic/gin"
)

// formatQueryParams 将 url.Values 格式化为更易读的字符串
func formatQueryParams(params url.Values) string {
	if len(params) == 0 {
		return ""
	}

	var parts []string
	for key, values := range params {
		// 对于单值参数，直接显示键值对
		if len(values) == 1 {
			parts = append(parts, fmt.Sprintf("%s=%s", key, values[0]))
		} else if len(values) > 1 {
			// 对于多值参数，显示键和值的数组
			parts = append(parts, fmt.Sprintf("%s=%v", key, values))
		}
	}

	return strings.Join(parts, "&")
}

// Gin Logger middleware
func (m *Middleware) GinRequestLogger(logger *logger.ContextLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Capture request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// Restore the request body for subsequent middleware/handlers
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Format query parameters using existing helper function
		queryParams := formatQueryParams(c.Request.URL.Query())

		// Log request information
		logger.Info(c, "[Request] Method: %s, Path: %s, Query: %s, Body: %s",
			c.Request.Method,
			c.Request.URL.Path,
			queryParams,
			string(requestBody),
		)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Log response information (without body)
		logger.Info(c, "[Response] Method: %s, Path: %s, Status: %d, Duration: %v",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
		)
	}
}

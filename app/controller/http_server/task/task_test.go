package controller_task

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 测试不同的Content-Disposition头格式
func TestContentDispositionHeader(t *testing.T) {
	// 初始化Gin引擎
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 创建CSV文件内容
	csvContent := []byte("id,name\n1,test\n")

	// 测试用例
	testCases := []struct {
		name     string
		filename string
		header   string
	}{
		{
			name:     "Standard ASCII",
			filename: "report.csv",
			header:   fmt.Sprintf(`attachment; filename="%s"`, "report.csv"),
		},
		{
			name:     "RFC 5987 Format",
			filename: "测试报告.csv",
			header:   fmt.Sprintf(`attachment; filename*=UTF-8''%s`, url.QueryEscape("测试报告.csv")),
		},
		{
			name:     "GBK Encoded",
			filename: "测试报告.csv",
			header:   fmt.Sprintf(`attachment; filename="%s"`, "测试报告.csv"), // GBK编码会在handler中处理
		},
		{
			name:     "Quoted UTF-8",
			filename: "测试报告.csv",
			header:   fmt.Sprintf(`attachment; filename="%s"`, "测试报告.csv"),
		},
		{
			name:     "Combined Format",
			filename: "测试报告.csv",
			header:   fmt.Sprintf(`attachment; filename="report.csv"; filename*=UTF-8''%s`, url.QueryEscape("测试报告.csv")),
		},
	}

	// 为每个测试用例创建路由
	for i, tc := range testCases {
		path := fmt.Sprintf("/export/%d", i)
		
		router.GET(path, func(header string) gin.HandlerFunc {
			return func(c *gin.Context) {
				c.Header("Content-Type", "text/csv; charset=utf-8")
				c.Header("Content-Disposition", header)
				c.Header("Content-Transfer-Encoding", "binary")
				c.Header("Cache-Control", "no-cache")
				c.Writer.Write(csvContent)
			}
		}(tc.header))
	}

	// 测试每个用例
	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/export/%d", i)
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 检查状态码
			assert.Equal(t, http.StatusOK, w.Code)
			
			// 检查响应头
			contentDisposition := w.Header().Get("Content-Disposition")
			t.Logf("Content-Disposition: %s", contentDisposition)
			
			// 检查响应体
			assert.Equal(t, csvContent, w.Body.Bytes())
		})
	}
} 
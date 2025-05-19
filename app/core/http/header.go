package http

import (
	"fmt"
	"net/http"
)

const (
	// HeaderAuthorization 认证头
	HeaderAuthorization = "Authorization"
	// HeaderContentType 内容类型头
	HeaderContentType = "Content-Type"
	// HeaderAccept 接受类型头
	HeaderAccept = "Accept"
	// HeaderUserAgent 用户代理头
	HeaderUserAgent = "User-Agent"

	// ContentTypeJSON JSON内容类型
	ContentTypeJSON = "application/json"
	// ContentTypeForm 表单内容类型
	ContentTypeForm = "application/x-www-form-urlencoded"
)

// SetBearerToken 设置Bearer令牌认证头
func SetBearerToken(req *http.Request, token string) {
	req.Header.Set(HeaderAuthorization, fmt.Sprintf("Bearer %s", token))
}

// SetAuthorization 设置Authorization头
func SetAuthorization(req *http.Request, token string) {
	req.Header.Set(HeaderAuthorization, token)
}

// SetJSONHeaders 设置JSON请求头
func SetJSONHeaders(req *http.Request) {
	req.Header.Set(HeaderContentType, ContentTypeJSON)
	req.Header.Set(HeaderAccept, ContentTypeJSON)
}

// SetFormHeaders 设置表单请求头
func SetFormHeaders(req *http.Request) {
	req.Header.Set(HeaderContentType, ContentTypeForm)
}

// SetUserAgent 设置用户代理
func SetUserAgent(req *http.Request, userAgent string) {
	req.Header.Set(HeaderUserAgent, userAgent)
}

// SetCustomHeader 设置自定义请求头
func SetCustomHeader(req *http.Request, key, value string) {
	req.Header.Set(key, value)
}

package auth

import (
	"fmt"

	"gil_teacher/app/conf"
	"gil_teacher/app/third_party/errorx"
	"gil_teacher/app/third_party/response"
	"gil_teacher/app/third_party/time"
	"gil_teacher/libs/encodingx"
	"gil_teacher/libs/httpx"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"
)

const AuthInfo_ = "auth_info"

const (
	OpenApiInfoPath = "/openapi/info"
)

type AdminAuth struct {
	httpClient *httpx.HttpClient
	conf       *conf.Config
	log        *log.Helper
}

func NewAdminAuth(c *conf.Config, logger log.Logger) *AdminAuth {
	return &AdminAuth{
		conf:       c,
		httpClient: httpx.NewHttpClient(logger, httpx.DialTimeout(time.ParseDuration(c.AdminAuth.Timeout))),
		log:        log.NewHelper(log.With(logger, "x_module", "middleware/NewAdminAuth")),
	}
}

type AdminAuthResponse struct {
	Status       int           `json:"status"`
	Code         int           `json:"code"`
	Message      string        `json:"message"`
	Data         AdminAuthInfo `json:"data"`
	ResponseTime int           `json:"response_time"`
}

type AdminAuthInfo struct {
	AdminId int    `json:"admin_id"`
	Name    string `json:"name"`
	Status  int    `json:"status"`
}

func (a *AdminAuth) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("token") // 前端调用后台必须携带有效的 Authorization
		if token == "" {
			response.NewResponse().Error(c, errorx.ErrAuthFail)
			c.Abort()
			return
		}

		// 根据token请求openapi接口获取用户信息
		url := fmt.Sprintf("%s%s", a.conf.AdminAuth.Domain, OpenApiInfoPath)
		resp := &AdminAuthResponse{}

		header := map[string]string{
			"Content-Type": "application/json",
			"token":        token,
		}
		body := map[string]int32{
			"system_id": int32(a.conf.AdminAuth.SystemId),
		}

		err := a.httpClient.Post(c, url, body, header, resp)
		if err != nil {
			a.log.WithContext(c).Infof("AdminAuth.Auth error:%v", err)
			response.NewResponse().Error(c, errorx.ErrAuthFail)
			c.Abort()
			return
		}
		if resp.Code != 0 {
			a.log.WithContext(c).Infof("AdminAuth.Auth error:%v", encodingx.ToJson(resp))
			response.NewResponse().Error(c, errorx.ErrAuthFail)
			c.Abort()
			return
		}

		// 将用户信息存入context
		c.Set(AuthInfo_, resp.Data)

		c.Next()
	}
}

// GetAdminInfo 获取当前登录用户信息
func GetAdminInfo(ctx *gin.Context) (info AdminAuthInfo, ok bool) {
	if adminInfo, exists := ctx.Get(AuthInfo_); exists {
		info, ok = adminInfo.(AdminAuthInfo)
	}

	return
}

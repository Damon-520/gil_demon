package router

import (
	"net/http"

	"gil_teacher/app/conf"
	"gil_teacher/app/controller/grpc_server/live_http"
	apies "gil_teacher/app/third_party/elasticsearch"
	"gil_teacher/app/third_party/middlewares/auth"
	"gil_teacher/app/third_party/response"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/openzipkin/zipkin-go"
	zipkinmiddleware "github.com/openzipkin/zipkin-go/middleware/http"
)

type GinRouter *gin.Engine

type responseWriter struct {
	gin.ResponseWriter
}

func RegisterRouter(
	r *gin.Engine,
	c *conf.Config,
	adminAuth *auth.AdminAuth,
	liveRoomHttp *live_http.LiveRoomHttp,
	tracer *zipkin.Tracer, // 接收 Zipkin Tracer
	esClient *elasticsearch.Client,
) GinRouter {

	// 创建 Zipkin 中间件 TODO
	zipkinMiddleware := zipkinmiddleware.NewServerMiddleware(tracer, zipkinmiddleware.SpanName("my-server"))
	ginZipkinMiddleware := func(c *gin.Context) {
		zipkinMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
			c.Writer = &responseWriter{ResponseWriter: c.Writer}
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	}
	_ = ginZipkinMiddleware
	// 添加根路由策略
	r.Any("/", func(ctx *gin.Context) {
		response.NewResponse().JsonRaw(ctx, nil)
	})

	// ############ 不验证权限 ############
	// noAuthGroup := r.Group("/demoApi")

	// Elasticsearch 操作示例
	r.GET("/es-search", func(c *gin.Context) {
		// 在这里执行 Elasticsearch 查询
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					"name": "Test Document",
				},
			},
		}

		res, err := apies.SearchDocument("test_index", query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	})

	// ############ 严格验证权限(必须验证) ############
	// strictAuthGroup := r.Group("/demoApi", ginZipkinMiddleware) // .Use(adminAuth.Auth())

	// {
	// ============ 直播间模块 ============
	// strictAuthGroup.POST("/v1/live/room/list", liveRoomHttp.List)     // 直播间列表
	// strictAuthGroup.POST("/v1/live/room/info", liveRoomHttp.Info)     // 直播间详情
	// strictAuthGroup.POST("/v1/live/room/create", liveRoomHttp.Create) // 直播间创建
	// strictAuthGroup.POST("/v1/live/room/edit", liveRoomHttp.Edit)     // 直播间编辑
	// strictAuthGroup.POST("/v1/live/room/update", liveRoomHttp.Update) // 直播间更新
	// strictAuthGroup.POST("/v1/live/room/es", liveRoomHttp.Es)         // 测试es

	// ============ 直播商品模块 ============

	// ============ 直播视频模块 ============

	// ============ 直播视频打点模块 ============

	// ============ 直播优惠券活动模块 ============
	// }

	return r
}

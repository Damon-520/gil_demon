package http

import (
	"gil_teacher/app/conf"
	"gil_teacher/app/service/db_test_service"
	"gil_teacher/app/third_party/errorx"
	"gil_teacher/app/third_party/response"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"
)

// DBTestController 数据库测试控制器
type DBTestController struct {
	log      *log.Helper
	dbTestSc *db_test_service.DBTestService
	config   *conf.Config
}

// NewDBTestController 创建数据库测试控制器
func NewDBTestController(
	logger log.Logger,
	dbTestSc *db_test_service.DBTestService,
	config *conf.Config,
) *DBTestController {
	return &DBTestController{
		log:      log.NewHelper(log.With(logger, "x_module", "controller/http/NewDBTestController")),
		dbTestSc: dbTestSc,
		config:   config,
	}
}

// RegisterRoutes 注册路由
func (c *DBTestController) RegisterRoutes(r *gin.Engine) {
	group := r.Group("/api/v1/db_test")
	{
		group.GET("/postgresql", c.TestPostgreSQL)
		group.GET("/postgresql/readwrite", c.TestPostgreSQLReadWrite)
		group.GET("/clickhouse", c.TestClickHouse)
		group.GET("/clickhouse/readwrite", c.TestClickHouseReadWrite)
		group.GET("/clickhouse/read", c.TestClickHouseRead)
		group.GET("/clickhouse/write", c.TestClickHouseWrite)
		group.GET("/all", c.TestAllDB)
	}
}

// TestPostgreSQL 测试PostgreSQL连接
func (c *DBTestController) TestPostgreSQL(ctx *gin.Context) {
	result, err := c.dbTestSc.TestPostgreSQLConnection(ctx)
	if err != nil {
		c.log.WithContext(ctx).Errorf("测试PostgreSQL连接失败: %v", err)
		response.NewResponse().Error(ctx, errorx.ErrServer)
		return
	}

	response.NewResponse().JsonRaw(ctx, gin.H{
		"connected": result,
	})
}

// TestPostgreSQLReadWrite 测试PostgreSQL读写分离连接
func (c *DBTestController) TestPostgreSQLReadWrite(ctx *gin.Context) {
	results := c.dbTestSc.TestPostgreSQLReadWriteConnections(ctx)
	response.NewResponse().JsonRaw(ctx, results)
}

// TestClickHouse 测试ClickHouse连接
func (c *DBTestController) TestClickHouse(ctx *gin.Context) {
	result, err := c.dbTestSc.TestClickHouseConnection(ctx)
	if err != nil {
		c.log.WithContext(ctx).Errorf("测试ClickHouse连接失败: %v", err)
		response.NewResponse().Error(ctx, errorx.ErrServer)
		return
	}

	response.NewResponse().JsonRaw(ctx, gin.H{
		"connected": result,
	})
}

// TestClickHouseReadWrite 测试ClickHouse读写分离连接
func (c *DBTestController) TestClickHouseReadWrite(ctx *gin.Context) {
	results := c.dbTestSc.TestClickHouseReadWriteConnections(ctx)
	response.NewResponse().JsonRaw(ctx, results)
}

// TestClickHouseRead 测试ClickHouse读连接
func (c *DBTestController) TestClickHouseRead(ctx *gin.Context) {
	result, err := c.dbTestSc.TestClickHouseReadConnection(ctx)
	if err != nil {
		c.log.WithContext(ctx).Errorf("测试ClickHouse读库连接失败: %v", err)
		response.NewResponse().Error(ctx, errorx.ErrServer)
		return
	}

	response.NewResponse().JsonRaw(ctx, gin.H{
		"connected": result,
	})
}

// TestClickHouseWrite 测试ClickHouse写连接
func (c *DBTestController) TestClickHouseWrite(ctx *gin.Context) {
	result, err := c.dbTestSc.TestClickHouseWriteConnection(ctx)
	if err != nil {
		c.log.WithContext(ctx).Errorf("测试ClickHouse写库连接失败: %v", err)
		response.NewResponse().Error(ctx, errorx.ErrServer)
		return
	}

	response.NewResponse().JsonRaw(ctx, gin.H{
		"connected": result,
	})
}

// TestAllDB 测试所有数据库连接
func (c *DBTestController) TestAllDB(ctx *gin.Context) {
	results := c.dbTestSc.TestAllDBConnections(ctx)
	response.NewResponse().JsonRaw(ctx, results)
}

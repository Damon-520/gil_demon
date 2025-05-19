package dao

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gil_teacher/app/conf"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DBTestClient 数据库连接测试客户端
type DBTestClient struct {
	log           *log.Helper
	pgConfig      *conf.PostgreSQL
	pgWriteConfig *conf.PostgreSQL
	pgReadConfig  *conf.PostgreSQL
	chConfig      *conf.Clickhouse
	chWriteConfig *conf.Clickhouse
	chReadConfig  *conf.Clickhouse
}

// NewDBTestClient 创建数据库测试客户端
func NewDBTestClient(c *conf.Data, logger log.Logger) *DBTestClient {
	return &DBTestClient{
		log:           log.NewHelper(log.With(logger, "x_module", "dao/NewDBTestClient")),
		pgConfig:      c.PostgreSQL,
		pgWriteConfig: c.PostgreSQLWrite,
		pgReadConfig:  c.PostgreSQLRead,
		chConfig:      c.Clickhouse,
		chWriteConfig: c.ClickhouseWrite,
		chReadConfig:  c.ClickhouseRead,
	}
}

// TestPostgreSQLConnection 测试PostgreSQL连接
func (c *DBTestClient) TestPostgreSQLConnection() (bool, error) {
	if c.pgConfig == nil && c.pgWriteConfig == nil {
		return false, fmt.Errorf("PostgreSQL配置为空")
	}

	c.log.Info("测试PostgreSQL连接...")

	// 使用旧配置
	if c.pgConfig != nil {
		return c.testSinglePostgreSQLConnection(c.pgConfig, "默认")
	}

	// 测试写库
	writeOK, writeErr := c.testSinglePostgreSQLConnection(c.pgWriteConfig, "写库")
	if !writeOK {
		return false, writeErr
	}

	// 测试读库(如果配置了)
	if c.pgReadConfig != nil {
		return c.testSinglePostgreSQLConnection(c.pgReadConfig, "读库")
	}

	return true, nil
}

// testSinglePostgreSQLConnection 测试单个PostgreSQL连接
func (c *DBTestClient) testSinglePostgreSQLConnection(config *conf.PostgreSQL, dbType string) (bool, error) {
	if config == nil {
		return false, fmt.Errorf("PostgreSQL %s 配置为空", dbType)
	}

	c.log.Infof("测试PostgreSQL %s 连接...", dbType)

	// 创建GORM配置
	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	}

	// 连接PostgreSQL
	db, err := gorm.Open(postgres.Open(config.Source), gormConfig)
	if err != nil {
		c.log.Errorf("PostgreSQL %s 连接失败: %v", dbType, err)
		return false, err
	}

	// 获取原生SQL连接用于关闭
	sqlDB, err := db.DB()
	if err != nil {
		c.log.Errorf("获取PostgreSQL %s 原生连接失败: %v", dbType, err)
		return false, err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxIdleTime(time.Duration(config.ConnMaxIdleTime) * time.Second)
	sqlDB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifeTime) * time.Second)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		c.log.Errorf("PostgreSQL %s ping失败: %v", dbType, err)
		return false, err
	}

	// 执行简单查询
	var version string
	row := db.Raw("SELECT version()").Row()
	if err := row.Scan(&version); err != nil {
		c.log.Errorf("PostgreSQL %s 查询版本失败: %v", dbType, err)
		return false, err
	}

	c.log.Infof("PostgreSQL %s 连接成功, 版本: %s", dbType, version)
	sqlDB.Close()
	return true, nil
}

// TestPostgreSQLReadWriteConnections 测试PostgreSQL读写分离连接
func (c *DBTestClient) TestPostgreSQLReadWriteConnections() map[string]interface{} {
	results := make(map[string]interface{})

	// 测试写库
	if c.pgWriteConfig != nil {
		writeResult, writeErr := c.testSinglePostgreSQLConnection(c.pgWriteConfig, "写库")
		results["write"] = map[string]interface{}{
			"connected": writeResult,
			"error":     writeErr,
		}
	} else if c.pgConfig != nil {
		// 使用旧配置
		writeResult, writeErr := c.testSinglePostgreSQLConnection(c.pgConfig, "默认")
		results["write"] = map[string]interface{}{
			"connected": writeResult,
			"error":     writeErr,
		}
	} else {
		results["write"] = map[string]interface{}{
			"connected": false,
			"error":     "写库配置缺失",
		}
	}

	// 测试读库
	if c.pgReadConfig != nil {
		readResult, readErr := c.testSinglePostgreSQLConnection(c.pgReadConfig, "读库")
		results["read"] = map[string]interface{}{
			"connected": readResult,
			"error":     readErr,
		}
	} else if c.pgWriteConfig != nil {
		results["read"] = map[string]interface{}{
			"connected": true,
			"error":     "使用写库作为读库",
		}
	} else if c.pgConfig != nil {
		results["read"] = map[string]interface{}{
			"connected": true,
			"error":     "使用默认库作为读库",
		}
	} else {
		results["read"] = map[string]interface{}{
			"connected": false,
			"error":     "读库配置缺失",
		}
	}

	return results
}

// TestClickHouseConnection 测试ClickHouse连接
func (c *DBTestClient) TestClickHouseConnection() (bool, error) {
	// 优先使用读写分离配置
	if c.chWriteConfig != nil {
		return c.TestClickHouseWriteConnection()
	}

	// 兼容旧配置
	if c.chConfig == nil {
		return false, fmt.Errorf("ClickHouse配置为空")
	}

	c.log.Info("测试ClickHouse连接(兼容模式)...")
	return c.testSingleClickHouseConnection(c.chConfig, "默认")
}

// TestClickHouseWriteConnection 测试ClickHouse写连接
func (c *DBTestClient) TestClickHouseWriteConnection() (bool, error) {
	if c.chWriteConfig == nil {
		return false, fmt.Errorf("ClickHouse写库配置为空")
	}

	c.log.Info("测试ClickHouse写库连接...")
	return c.testSingleClickHouseConnection(c.chWriteConfig, "写库")
}

// TestClickHouseReadConnection 测试ClickHouse读连接
func (c *DBTestClient) TestClickHouseReadConnection() (bool, error) {
	if c.chReadConfig == nil {
		// 如果没有配置读库，则使用写库
		if c.chWriteConfig != nil {
			c.log.Info("ClickHouse读库配置为空，使用写库配置...")
			return c.testSingleClickHouseConnection(c.chWriteConfig, "读库(使用写库)")
		}
		return false, fmt.Errorf("ClickHouse读库配置为空")
	}

	c.log.Info("测试ClickHouse读库连接...")
	return c.testSingleClickHouseConnection(c.chReadConfig, "读库")
}

// TestClickHouseReadWriteConnections 测试ClickHouse读写分离连接
func (c *DBTestClient) TestClickHouseReadWriteConnections() map[string]interface{} {
	results := make(map[string]interface{})

	// 测试写库
	if c.chWriteConfig != nil {
		writeResult, writeErr := c.testSingleClickHouseConnection(c.chWriteConfig, "写库")
		results["write"] = map[string]interface{}{
			"connected": writeResult,
			"error":     writeErr,
		}
	} else if c.chConfig != nil {
		// 使用旧配置
		writeResult, writeErr := c.testSingleClickHouseConnection(c.chConfig, "默认")
		results["write"] = map[string]interface{}{
			"connected": writeResult,
			"error":     writeErr,
		}
	} else {
		results["write"] = map[string]interface{}{
			"connected": false,
			"error":     "写库配置缺失",
		}
	}

	// 测试读库
	if c.chReadConfig != nil {
		readResult, readErr := c.testSingleClickHouseConnection(c.chReadConfig, "读库")
		results["read"] = map[string]interface{}{
			"connected": readResult,
			"error":     readErr,
		}
	} else if c.chWriteConfig != nil {
		results["read"] = map[string]interface{}{
			"connected": true,
			"error":     "使用写库作为读库",
		}
	} else if c.chConfig != nil {
		results["read"] = map[string]interface{}{
			"connected": true,
			"error":     "使用默认库作为读库",
		}
	} else {
		results["read"] = map[string]interface{}{
			"connected": false,
			"error":     "读库配置缺失",
		}
	}

	return results
}

// testSingleClickHouseConnection 测试单个ClickHouse连接
func (c *DBTestClient) testSingleClickHouseConnection(config *conf.Clickhouse, dbType string) (bool, error) {
	if config == nil {
		return false, fmt.Errorf("ClickHouse %s 配置为空", dbType)
	}

	c.log.Infof("测试ClickHouse %s 连接...", dbType)

	// 设置超时时间
	timeout := time.Duration(config.DialTimeout) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	// 尝试使用HTTP方式连接ClickHouse
	_, version, err := c.testClickHouseHTTPWithConfig(config, dbType)
	if err == nil {
		c.log.Infof("ClickHouse %s 通过HTTP连接成功, 版本: %s", dbType, version)
		return true, nil
	}

	c.log.Warnf("ClickHouse %s HTTP连接失败, 尝试使用原生TCP协议连接: %v", dbType, err)

	// 如果HTTP失败，尝试TCP连接
	addr := config.Address
	c.log.Infof("使用TCP连接ClickHouse %s: %v", dbType, addr)

	// 连接ClickHouse
	clickhouseClient := clickhouse.OpenDB(&clickhouse.Options{
		Addr: addr,
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.Username,
			Password: config.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": config.MaxExecutionTime,
		},
		DialTimeout: timeout,
	})

	// 设置连接参数
	clickhouseClient.SetMaxOpenConns(10)
	clickhouseClient.SetMaxIdleConns(5)
	clickhouseClient.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := clickhouseClient.Ping(); err != nil {
		c.log.Errorf("ClickHouse %s TCP ping失败: %v", dbType, err)
		return false, err
	}

	// 执行简单查询
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := clickhouseClient.QueryRowContext(ctx, "SELECT version()")
	var version2 string
	if err := row.Scan(&version2); err != nil {
		c.log.Errorf("ClickHouse %s 查询版本失败: %v", dbType, err)
		return false, err
	}

	c.log.Infof("ClickHouse %s TCP连接成功, 版本: %s", dbType, version2)
	clickhouseClient.Close()
	return true, nil
}

// testClickHouseHTTPWithConfig 使用指定配置通过HTTP方式测试ClickHouse连接
func (c *DBTestClient) testClickHouseHTTPWithConfig(config *conf.Clickhouse, dbType string) (bool, string, error) {
	// 获取地址
	if len(config.Address) == 0 {
		return false, "", fmt.Errorf("ClickHouse %s 地址为空", dbType)
	}

	// 构建HTTP URL
	baseAddr := config.Address[0]
	if !strings.HasPrefix(baseAddr, "http") {
		baseAddr = "http://" + baseAddr
	}

	// 添加查询参数
	url := fmt.Sprintf("%s?query=%s", baseAddr, "SELECT%20version()")

	// 使用HTTP客户端直接请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, "", fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置Basic认证
	if config.Username != "" {
		req.SetBasicAuth(config.Username, config.Password)
	}

	// 发送请求
	httpClient := &http.Client{
		Timeout: time.Duration(config.DialTimeout) * time.Second,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("HTTP状态码错误: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 返回版本
	version := strings.TrimSpace(string(body))
	return true, version, nil
}

// 保留现有的testClickHouseHTTP方法，但委托给testClickHouseHTTPWithConfig
func (c *DBTestClient) testClickHouseHTTP() (bool, string, error) {
	if c.chConfig == nil {
		return false, "", fmt.Errorf("ClickHouse配置为空")
	}
	return c.testClickHouseHTTPWithConfig(c.chConfig, "默认")
}

// TestAllDBConnections 测试所有数据库连接
func (c *DBTestClient) TestAllDBConnections() map[string]interface{} {
	results := make(map[string]interface{})

	// 测试PostgreSQL
	pgResults := c.TestPostgreSQLReadWriteConnections()
	results["postgresql"] = pgResults

	// 测试ClickHouse
	chResults := c.TestClickHouseReadWriteConnections()
	results["clickhouse"] = chResults

	return results
}

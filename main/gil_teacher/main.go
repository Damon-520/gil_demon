package main

import (
	"context"
	"fmt"
	"gil_teacher/app/core/envx"
	"gil_teacher/common"

	"github.com/gin-gonic/gin"

	"gil_teacher/app/conf"
	"gil_teacher/app/third_party/elasticsearch"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Version is the version of the compiled software.
	Version string
)

func initNacos(nacos *conf.Nacos) config_client.IConfigClient {
	// ServerConfig
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      nacos.Host,
			Port:        uint64(nacos.Port),
			ContextPath: "/nacos",
			Scheme:      "http",
		},
	}
	// ClientConfig
	clientConfig := constant.ClientConfig{
		NamespaceId:         "public", // If you need to specify the namespace, fill in the ID of the namespace here.
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "nacos/log",
		CacheDir:            "nacos/cache",
		LogLevel:            "debug",
		Username:            nacos.Username,
		Password:            nacos.Password,
	}

	// Create config client
	configClient, err := clients.CreateConfigClient(map[string]any{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})

	if err != nil {
		log.Fatalf("create config client failed: %v", err)
	}
	fmt.Println("Nacos client initialized successfully!")
	return configClient
}

func main() {
	// 初始化基础配置
	c, logger_, cmdParams := common.InitBase(true)

	// 根据环境设置gin模式
	if cmdParams.Env == envx.ENV_PROD {
		gin.SetMode(gin.ReleaseMode)
	}
	//
	//c := config.New(
	//	config.WithSource(
	//		file.NewSource(cmdParams.FlagConf),
	//	),
	//)
	//defer c.Close()
	//
	//if err := c.Load(); err != nil {
	//	panic(err)
	//}
	//
	//// 使用baseConf而不是创建新的bc
	//if err := c.Scan(baseConf); err != nil {
	//	panic(err)
	//}
	//
	//baseConf.App.Mode = cmdParams.Mode
	//
	//// 初始化Nacos客户端
	//configClient := initNacos(c.Data.Nacos)

	//// 获取配置
	//content, err := configClient.GetConfig(vo.ConfigParam{
	//	DataId: "example-data-id",
	//	Group:  "example-group",
	//})
	//
	//if err != nil {
	//	// log.Fatalf("get config failed: %v", err)
	//}
	//
	//fmt.Printf("config content: %s\n", content)

	// 初始化 Elasticsearch 连接
	esClient, err := elasticsearch.InitES(c.Elasticsearch)
	if err != nil {
		// log.Fatal("Elasticsearch connection failed: ", err)
	}
	_ = esClient

	// 如果索引不存在则创建索引
	if err := elasticsearch.CreateIndex("test_index"); err != nil {
		// log.Fatal("创建索引失败: ", err)
	}

	// 使用 wire 生成 app
	app, cleanup, err := wireApp(context.Background(), c.Server, c, c.Data, c.Config, logger_)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}

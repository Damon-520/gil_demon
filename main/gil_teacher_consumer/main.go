package main

import (
	"context"
	"gil_teacher/app/core/envx"
	"gil_teacher/app/core/healthx"
	"gil_teacher/common"
	"github.com/gin-gonic/gin"

	clogger "gil_teacher/app/core/logger"
	"gil_teacher/app/domain/behavior"
	// "github.com/segmentio/kafka-go"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Version is the version of the compiled software.
	Version string
	// flagConf is the config flag.
	flagConf string
	mode     int
)

//
//func init() {
//	flag.StringVar(&flagConf, "conf", "./configs/local/api/", "config path, eg: -conf config.yaml")
//	flag.IntVar(&mode, "mode", 0, "start in HTTP or gRPC mode. Default value 0 starts both, 1 starts HTTP only, 2 starts gRPC only.")
//}

func main() {

	// 初始化基础配置
	bc, logger_, cmdParams := common.InitBase(true)

	// 根据环境设置gin模式
	if cmdParams.Env == envx.ENV_PROD {
		gin.SetMode(gin.ReleaseMode)
	}
	// 定义 Kafka broker 地址
	// brokers := []string{"localhost:9092"}

	// // 定义主题与对应的处理函数
	// topics := map[string]func(context.Context, kafka.Message){
	// 	"topic1": handleTopic1,
	// 	"topic2": handleTopic2,
	// 	// 可以根据需要添加更多主题及其处理函数
	// }

	//flag.Parse()
	//c := config.New(
	//	config.WithSource(
	//		file.NewSource(flagConf),
	//	),
	//)
	//defer c.Close()
	//
	//if err := c.Load(); err != nil {
	//	panic(err)
	//}
	//
	//var bc conf.Conf
	//if err := c.Scan(&bc); err != nil {
	//	panic(err)
	//}
	//
	//bc.App.Mode = mode

	//logger := log.With(
	//	logger.NewLogger(logger.Config{
	//		Path:         bc.Log.Path,
	//		Level:        bc.Log.Level,
	//		RotationTime: time.ParseDuration(bc.Log.RotationTime),
	//		MaxAge:       time.ParseDuration(bc.Log.MaxAge),
	//	}),
	//	"service_name", bc.App.Name,
	//)

	healthServer := healthx.NewHealthServer(cmdParams.Env, cmdParams.ScriptHealthzPort)
	healthServer.Run()

	behaviorHandler, cleanup, err := wireApp(bc.Server, bc, bc.Data, bc.Config, logger_)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	go func() {
		behaviorConsumer := behavior.NewBehaviorConsumer(bc.Data.Kafka, clogger.NewContextLogger(logger_))
		behaviorConsumer.Consume(context.Background(), behaviorHandler)
	}()

	// 阻塞主线程，防止程序退出
	select {}
}

// // consumeTopic 启动一个 Kafka 消费者，订阅指定的主题，并使用提供的处理函数处理消息
// func consumeTopic(brokers []string, topic, groupID string, handler func(context.Context, kafka.Message)) {
// 	r := kafka.NewReader(kafka.ReaderConfig{
// 		Brokers:  brokers,
// 		Topic:    topic,
// 		GroupID:  groupID,
// 		MinBytes: 10e3, // 10KB
// 		MaxBytes: 10e6, // 10MB
// 	})
// 	defer r.Close()

// 	ctx := context.Background()
// 	for {
// 		msg, err := r.ReadMessage(ctx)
// 		if err != nil {
// 			log.Printf("读取消息失败: %v", err)
// 			continue
// 		}
// 		handler(ctx, msg)
// 	}
// }

// // handleTopic1 处理来自 topic1 的消息
// func handleTopic1(ctx context.Context, msg kafka.Message) {
// 	fmt.Printf("处理 topic1 的消息：%s\n", string(msg.Value))
// 	// 在此添加具体的消息处理逻辑
// }

// // handleTopic2 处理来自 topic2 的消息
// func handleTopic2(ctx context.Context, msg kafka.Message) {
// 	fmt.Printf("处理 topic2 的消息：%s\n", string(msg.Value))
// 	// 在此添加具体的消息处理逻辑
// }

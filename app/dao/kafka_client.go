package dao

// import (
// 	"context"

// 	"gil_teacher/app/conf"
// 	"gil_teacher/app/core/kafka"
// 	clogger "gil_teacher/app/core/logger"

// 	"github.com/go-kratos/kratos/v2/log"
// )

// // 添加 kafka client 的提供者函数
// func ProvideKafkaClient(conf *conf.Data, logger log.Logger) kafka.ApiKafkaClient {
// 	return kafka.NewApiKafkaClient(context.Background(), conf, clogger.NewContextLogger(logger))
// }
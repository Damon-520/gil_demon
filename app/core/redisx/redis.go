package redisx

import (
	"context"

	"gil_teacher/app/conf"
	clogger "gil_teacher/app/core/logger"
	"gil_teacher/app/third_party/time"

	"github.com/go-redis/redis/v8"
)

func NewRedis(conf *conf.Redis, logger *clogger.ContextLogger) *redis.Client {
	options := redis.Options{
		Addr:         conf.Address,
		Username:     conf.Username,
		Password:     conf.Password,
		DB:           int(conf.Database),
		DialTimeout:  time.ParseDuration(conf.DialTimeout),
		WriteTimeout: time.ParseDuration(conf.WriteTimeout),
		ReadTimeout:  time.ParseDuration(conf.ReadTimeout),
	}
	rdb := redis.NewClient(&options)
	if rdb == nil {
		logger.Fatal(context.Background(), "failed opening connection to redisx")
		return nil
	}

	// ping
	if res := rdb.Ping(context.Background()).Err(); res != nil {
		logger.Fatal(context.Background(), "failed ping to redisx, err: %v", res)
		return nil
	}

	return rdb
}

package test

import (
	"gil_teacher/app/conf"
	"gil_teacher/app/core/logger"
	"gil_teacher/app/dao"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	chClients map[string]*dao.ClickHouseRWClient
	db        *gorm.DB
	Clog      *logger.ContextLogger
)

func init() {
	Clog = logger.NewContextLogger(log.With(
		logger.NewLogger(logger.Config{
			Path:         "./log/test.log",
			Level:        "debug",
			Rotationtime: 86400 * time.Second,
			Maxage:       259200 * time.Second,
		}),
		"service_name", "gil_teacher",
	))
	var err error
	chClients, _, err = dao.NewClickHouseRWClient(&conf.Data{
		ClickhouseWrite: &conf.Clickhouse{
			Address:          []string{"cc.local.xiaoluxue.cn:9000"},
			Databases:        []string{"db_gil_teacher", "db_gil_student"},
			Username:         "sunny",
			Password:         "sunny",
			MaxExecutionTime: 60,
			DialTimeout:      10,
			ReadTimeout:      10,
		},
		ClickhouseRead: &conf.Clickhouse{
			Address:          []string{"cc.local.xiaoluxue.cn:9000"},
			Database:         "db_gil_teacher",
			Username:         "sunny",
			Password:         "sunny",
			MaxExecutionTime: 60,
			DialTimeout:      10,
			ReadTimeout:      10,
		},
	}, Clog)
	if err != nil {
		panic(err)
	}
	db, err = gorm.Open(postgres.Open("host=pg.local.xiaoluxue.cn user=gil_admin password=vt0wq1QH8&NX^WouUb dbname=db_teacher port=5432 sslmode=disable TimeZone=Asia/Shanghai"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
}

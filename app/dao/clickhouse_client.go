package dao

import (
	"database/sql"
	"time"

	"gil_teacher/app/conf"

	"github.com/ClickHouse/clickhouse-go/v2"
)

func NewClickhouseClient(c *conf.Clickhouse) *sql.DB {
	clickhouseClient := clickhouse.OpenDB(&clickhouse.Options{
		Addr: c.Address,
		Auth: clickhouse.Auth{
			Database: c.Database,
			Username: c.Username,
			Password: c.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": c.MaxExecutionTime,
		},
		DialTimeout: time.Duration(5) * time.Second,
		//Debug:       true,
	})

	return clickhouseClient
}

package dao

/*

import (
	"gil_teacher/app/conf"
	"fmt"
	"testing"
	"time"
)

func TestNewClickhouseClient(t *testing.T) {
	clickhouseConf := &conf.Clickhouse{
		Address:  []string{"localhost:19000"},
		Database: "gil_db_name",
		Username: "default",
		Password: "changeme",
	}

	clickhouseClient := NewClickhouseClient(clickhouseConf)
	defer func() { _ = clickhouseClient.Close() }()

	rows, err := clickhouseClient.Query("select user_id, event_type, event_time from user_activity")
	if err != nil {
		fmt.Printf("query failed. err:%+v \r\n", err)
		return
	}

	for rows.Next() {
		var (
			userId    int32
			eventType string
			eventTime time.Time
		)

		err = rows.Scan(&userId, &eventType, &eventTime)
		if err != nil {
			fmt.Printf("query failed. err:%+v \r\n", err)
			return
		}

		fmt.Printf("userId:%d eventType:%s eventTime:%+v\r\n", userId, eventType, eventTime)

	}

}
*/

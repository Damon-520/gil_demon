package demoTask1

import (
	"context"
	"fmt"
	"time"

	"gil_teacher/script/common"

	"github.com/xxl-job/xxl-job-executor-go"
)

const (
	Pattern = "task1"
)

func Task1(ctx context.Context, param *xxl.RunReq) string {
	fmt.Println("Task1 will start......")
	time.Sleep(15 * time.Second)
	fmt.Println("Task1 done!!")
	// send notify to feishu
	common.SendFeishuWebhook("Task1 done success")
	return "success"
}

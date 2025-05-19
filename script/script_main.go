package main

import (
	"context"
	"fmt"
	"log"

	"github.com/xxl-job/xxl-job-executor-go"

	"gil_teacher/script/demoTask1"
	"gil_teacher/script/demoTask2"
)

const (
	XXL_ADMIN         = "http://127.0.0.1:8080/xxl-job-admin" //todo 从nacos 获取admin地址
	XXL_REGISTER_PORT = "9777"                                //默认9999, 建议使用 9777 - 9888
	XXL_TOKEN         = "test_token"                          //请求令牌,与admin保持一致
	XXL_REGISTER_KEY  = "golang-jobs"                         //执行器名称, todo 后续从nacos读取或者本地写死
)

func main() {
	// 动态获取xxl-job-admin地址
	exec := xxl.NewExecutor(
		xxl.ServerAddr(XXL_ADMIN),
		xxl.AccessToken(XXL_TOKEN),
		xxl.ExecutorPort(XXL_REGISTER_PORT),
		xxl.RegistryKey(XXL_REGISTER_KEY), //执行器名称 , todo 后续从nacos读取或者本地写死
		xxl.SetLogger(&logger{}),          //自定义日志
	)
	exec.Init()
	exec.Use(customMiddleware)
	//设置日志查看handler
	exec.LogHandler(customLogHandle)
	//注册任务handler
	exec.RegTask(demoTask1.Pattern, demoTask1.Task1)
	exec.RegTask(demoTask2.Pattern, demoTask2.Task1)
	log.Fatal(exec.Run())
}

// 自定义日志处理器
func customLogHandle(req *xxl.LogReq) *xxl.LogRes {
	return &xxl.LogRes{Code: xxl.SuccessCode, Msg: "", Content: xxl.LogResContent{
		FromLineNum: req.FromLineNum,
		ToLineNum:   2,
		LogContent:  "这个是自定义日志handler",
		IsEnd:       true,
	}}
}

// xxl.Logger接口实现
type logger struct{}

func (l *logger) Info(format string, a ...interface{}) {
	fmt.Println(fmt.Sprintf("自定义日志 - "+format, a...))
}

func (l *logger) Error(format string, a ...interface{}) {
	log.Println(fmt.Sprintf("自定义日志 - "+format, a...))
}

// 自定义中间件
func customMiddleware(tf xxl.TaskFunc) xxl.TaskFunc {
	return func(cxt context.Context, param *xxl.RunReq) string {
		log.Println("I am a middleware start")
		res := tf(cxt, param)
		log.Println("I am a middleware end", res)
		return res
	}
}

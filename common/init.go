package common

import (
	"fmt"
	"gil_teacher/app/conf"
	"gil_teacher/app/core/logger"
	"gil_teacher/app/core/nacosx"
	"gil_teacher/app/utils/prometheus"
	"gil_teacher/app/utils/time"
	"net"
	"os"
	"path/filepath"
	rawTime "time"

	"github.com/go-kratos/kratos/v2/log"
)

func timeZoneInit() {
	// 设置程序全局时区（UTC+8）
	var cstZone = rawTime.FixedZone("Asia/Shanghai", 8*60*60) // 东八区
	rawTime.Local = cstZone
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no valid IPv4 address found")
}

func nacosInit(bc *conf.Conf) {
	ip, err := getLocalIP()
	if err != nil {
		panic(err)
	}
	cmdParams.NacosConf.ServiceIp = ip
	cmdParams.NacosConf.DataId = "teacher-config.yaml"
	nacosxClient, err := nacosx.NewNacosClientX(cmdParams.NacosConf)
	if err != nil {
		panic(err)
	}

	err = nacosxClient.LoadConfig(bc)
	if err != nil {
		panic(err)
	}

	// 添加配置加载后的调试日志
	fmt.Printf("Loaded config: %+v\n", bc)
	if bc.Log == nil {
		panic("Log configuration is nil after loading from Nacos")
	}
	fmt.Printf("Log config: %+v\n", bc.Log)

	err = nacosxClient.WatchConfig(bc)
	if err != nil {
		panic(err)
	}

	if cmdParams.ApiService {
		err = nacosxClient.ServiceRegister()
		if err != nil {
			panic(err)
		}
	}
}

func GetExecutableName() (string, error) {
	// 获取可执行文件完整路径
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	// 解析符号链接（如果有）
	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		realPath = exePath // 如果解析失败，使用原始路径
	}

	// 提取纯文件名（带扩展名）
	return filepath.Base(realPath), nil
}

func logInit(bc *conf.Conf) log.Logger {

	appName, err := GetExecutableName()
	if err != nil {
		appName = "teacher-api"
	}

	return log.With(
		logger.NewLogger(logger.Config{
			Path:         bc.Log.Path,
			Level:        bc.Log.Level,
			RotationTime: time.ParseDuration(bc.Log.RotationTime),
			MaxAge:       time.ParseDuration(bc.Log.MaxAge),
		}),
		"service_name", appName,
	)
}

func InitBase(apiService bool) (*conf.Conf, log.Logger, CmdParams) {
	var bc conf.Conf
	timeZoneInit()
	envInit()
	parseFlag()
	cmdParams.ApiService = apiService
	nacosInit(&bc)
	logger_ := logInit(&bc)
	prometheus.Init()
	return &bc, logger_, cmdParams
}

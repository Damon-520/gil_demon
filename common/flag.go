package common

import (
	"flag"
	"gil_teacher/app/core/envx"
	"gil_teacher/app/core/nacosx"
)

type CmdParams struct {
	Env               string
	Mode              int
	ScriptHealthzPort int64
	NacosConf         nacosx.NacosConf
	ApiService        bool
	FlagConf          string
}

var cmdParams = CmdParams{}

func envInit() {
	cmdParams.Env = envx.GetEnv("env")
	if cmdParams.Env == "" {
		cmdParams.Env = envx.ENV_LOCAL
	}
}

func parseFlag() {
	flag.StringVar(&cmdParams.FlagConf, "conf", "./configs/local/api/", "config path, eg: -conf config.yaml")
	flag.StringVar(&cmdParams.Env, "env", cmdParams.Env, "env, eg: --env local")
	flag.IntVar(&cmdParams.Mode, "mode", 1, "env, eg: --mode 1")
	flag.Int64Var(&cmdParams.ScriptHealthzPort, "healthcheck_port", 8080, "--healthcheck_port, eg: 8080")
	flag.StringVar(&cmdParams.NacosConf.ServerAddrList, "nacos_servers", "nacos.local.xiaoluxue.cn:8848", "nacos server addresses, eg: --nacos_servers 127.0.0.1:8848")
	flag.StringVar(&cmdParams.NacosConf.Username, "nacos_username", "nacos_register", "nacos username, eg: --nacos_username nacos")
	flag.StringVar(&cmdParams.NacosConf.Password, "nacos_password", "VsFVkown9afKg5P8", "nacos password, eg: --nacos_password 123456")
	flag.StringVar(&cmdParams.NacosConf.NamespaceId, "nacos_namespace_id", "local", "nacos namespace_id, eg: --nacos_namespace_id local")
	flag.StringVar(&cmdParams.NacosConf.Group, "nacos_group", "DEFAULT_GROUP", "nacos group, eg: --nacos_group DEFAULT_GROUP")
	flag.StringVar(&cmdParams.NacosConf.ServiceName, "nacos_service_name", "teacher-api", "nacos service name, eg: --nacos_service_name teacher-api")
	flag.Uint64Var(&cmdParams.NacosConf.ServicePort, "nacos_service_port", 8280, "nacos service port, eg: --nacos_service_port 8280")

	flag.Parse()
}

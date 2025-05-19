package nacosx

import (
	"fmt"
	"gil_teacher/app/conf"
	"strconv"
	"strings"
	"sync"

	nacosConfig "github.com/go-kratos/kratos/contrib/config/nacos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

type NacosConf struct {
	ServerAddrList string
	Username       string
	Password       string
	NamespaceId    string
	Group          string
	ServiceName    string
	ServiceIp      string
	ServicePort    uint64
	DataId         string
}

type NacosClientX struct {
	sync.RWMutex
	cfgClient      config_client.IConfigClient
	namingClient   naming_client.INamingClient
	serverConfigs  []constant.ServerConfig
	clientConfig   constant.ClientConfig
	registerConfig vo.RegisterInstanceParam
	cfg            config.Config
	nacosConf      NacosConf
}

func NewNacosClientX(nacosConf NacosConf) (*NacosClientX, error) {

	serverConfigs := make([]constant.ServerConfig, 0)
	nacosServerAddrList := strings.Split(nacosConf.ServerAddrList, ",")
	for _, nacosServerAddr := range nacosServerAddrList {
		hostPort := strings.Split(nacosServerAddr, ":")
		if len(hostPort) != 2 {
			log.Fatalf("invalid nacos server address: %s", nacosServerAddr)
		}

		nacosPort, err := strconv.ParseUint(hostPort[1], 10, 64)
		if err != nil {
			log.Fatalf("invalid nacos server address: %s", nacosServerAddr)
		}

		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr: hostPort[0],
			Port:   nacosPort,
		})
	}

	clientConfig := constant.ClientConfig{
		Username:            nacosConf.Username,
		Password:            nacosConf.Password,
		NamespaceId:         nacosConf.NamespaceId,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "nacos/log",
		CacheDir:            "nacos/cache",
		LogLevel:            "debug",
	}

	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})

	if err != nil {
		return nil, err
	}

	log.Info("Nacos config client initialized successfully!")

	discoveryClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)

	if err != nil {
		return nil, err
	}

	log.Info("Nacos discovery client initialized successfully!")

	nacosSource := nacosConfig.NewConfigSource(configClient,
		nacosConfig.WithGroup(nacosConf.Group),
		nacosConfig.WithDataID(nacosConf.DataId),
	)

	cfg := config.New(
		config.WithSource(
			nacosSource,
		),
	)

	registerConfig := vo.RegisterInstanceParam{
		Ip:          nacosConf.ServiceIp,
		Port:        nacosConf.ServicePort,
		ServiceName: nacosConf.ServiceName,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
	}

	return &NacosClientX{
		cfgClient:      configClient,
		namingClient:   discoveryClient,
		registerConfig: registerConfig,
		cfg:            cfg,
		nacosConf:      nacosConf,
	}, err
}

func (nx *NacosClientX) LoadConfig(bc *conf.Conf) error {
	defer nx.Unlock()
	nx.Lock()

	defer nx.cfg.Close()

	fmt.Printf("Loading config from Nacos...\n")
	fmt.Printf("DataId: %s, Group: %s\n", nx.nacosConf.DataId, nx.nacosConf.Group)

	content, err := nx.cfgClient.GetConfig(vo.ConfigParam{
		DataId: nx.nacosConf.DataId,
		Group:  nx.nacosConf.Group,
	})
	if err != nil || content == "" {
		fmt.Printf("Failed to get config from Nacos or content is empty, using local config: %v\n", err)
		// 使用本地配置文件
		localConfig := config.New(
			config.WithSource(
				file.NewSource("configs/local/api/config.yaml"),
			),
		)
		defer localConfig.Close()

		if err := localConfig.Load(); err != nil {
			log.Errorf("load local config error: %v", err)
			return err
		}

		if err := localConfig.Scan(bc); err != nil {
			log.Errorf("scan local config error: %v", err)
			return err
		}

		return nil
	}

	fmt.Printf("Raw config content from Nacos:\n%s\n", content)

	if err := nx.cfg.Load(); err != nil {
		log.Errorf("load config error: %v", err)
		return err
	}

	if err := nx.cfg.Scan(bc); err != nil {
		log.Errorf("scan config error: %v", err)
		return err
	}

	return nil
}

func (nx *NacosClientX) WatchConfig(bc *conf.Conf) error {
	err := nx.cfgClient.ListenConfig(vo.ConfigParam{
		DataId: nx.nacosConf.DataId,
		Group:  nx.nacosConf.Group,
		Type:   vo.YAML,
		OnChange: func(namespace, group, dataId, data string) {
			fmt.Printf("changed, namespace: %s, group: %s, dataId: %s, data: %s\n", namespace, group, dataId, data)
			err := nx.LoadConfig(bc)
			if err != nil {
				log.Fatalf("nacos watch config load err: %v", err)
				return
			}
		},
	})
	return err
}

func (nx *NacosClientX) ServiceRegister() error {

	success, err := nx.namingClient.RegisterInstance(nx.registerConfig)

	if err != nil {
		return fmt.Errorf("register instance failed: %v", err)
	}

	if !success {
		return fmt.Errorf("register instance failed")
	}

	return nil
}

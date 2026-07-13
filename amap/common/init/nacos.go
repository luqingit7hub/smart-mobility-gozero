package init

import (
	"common/config"
	"fmt"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
)

type NaCos struct {
	Host        string
	Port        int
	NamespaceId string
	User        string
	Password    string
	DataId      string
	Group       string
}

func NacosInit() {
	viper.SetConfigFile("../common/yaml/nacos.yaml")
	//viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
		return
	}
	var NaCosConfig NaCos
	err = viper.UnmarshalKey("NaCos", &NaCosConfig)
	if err != nil {
		panic(err)
		return
	}
	// Nacos服务器地址
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: NaCosConfig.Host,
			Port:   uint64(NaCosConfig.Port),
		},
	}
	// 客户端配置
	clientConfig := constant.ClientConfig{
		NamespaceId:         NaCosConfig.NamespaceId, // 如果不需要命名空间，可以留空
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "debug",
	}

	// 创建配置客户端
	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	yaml, err := configClient.GetConfig(vo.ConfigParam{
		DataId: NaCosConfig.DataId,
		Group:  NaCosConfig.Group,
	})

	viper.Reset()
	viper.SetConfigType("yaml")
	err = viper.ReadConfig(strings.NewReader(yaml))
	if err != nil {
		panic(err)
		return
	}
	if yaml==""{
		fmt.Println("获取的yaml文件为空,异常情况")
		panic(err)
		return
	}
	err = viper.Unmarshal(&config.DataConfig)
	if err != nil {
		panic(err)
		return
	}
	fmt.Println("NaCos连接配置成功")
}

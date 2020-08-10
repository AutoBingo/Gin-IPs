package configure

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type Oid string

const (
	OidHost   Oid = "HOST"
	OidSwitch Oid = "SWITCH"
)

var OidArray = []Oid{OidHost, OidSwitch}

const (
	InstanceCollection = "instances" // 存放实例的集合
)

// 通用 Config 接口
type Config interface {
	InitError(msg string) error
}

// 根据yaml文件初始化通用配置, 无需输出日志
func InitYaml(filename string, config Config) error {
	fp, err := os.Open(filename)
	if err != nil {
		msg := fmt.Sprintf("configure file [ %s ] not found", filename)
		return config.InitError(msg)
	}
	defer func() {
		_ = fp.Close()
	}()
	if err := yaml.NewDecoder(fp).Decode(config); err != nil {
		msg := fmt.Sprintf("configure file [ %s ] initialed failed", filename)
		return config.InitError(msg)
	}
	return nil
}

// 定义 Gin 配置变量，可以仿照这个多拆分几个变量
type GinConfig struct {
	ApiServer struct {
		Env  string `yaml:"env"` // dev/prod
		Host string `yaml:"host"`
		Port int64  `yaml:"port"`
	} `yaml:"api_server"`
	Mgo struct {
		Uri      string `yaml:"uri"`
		Database string `yaml:"database"`
		PoolSize uint64 `yaml:"pool_size"`
	} `yaml:"mgo"`
	Log struct {
		Name  string `yaml:"name"`
		Path  string `yaml:"path"`
		Level string `yaml:"level"`
		Count uint   `yaml:"count"`
	} `yaml:"log"`
}

func (*GinConfig) InitError(msg string) error {
	return errors.New(msg)
}

//  初始化配置文件变量
var GinConfigValue = new(GinConfig)

// main 调用
func InitConfigValue() error {
	if err := InitYaml("conf/gin_ips.yaml", GinConfigValue); err != nil {
		return err
	}
	return nil
}

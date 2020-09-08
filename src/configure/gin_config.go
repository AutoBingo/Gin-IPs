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

// 路由定义  /cmdb/$version/$DevPkg/$ApiURI
type DevPkg string

const (
	ObjectPkg DevPkg = "object" // 标准模型搜索
	SdkPkg    DevPkg = "sdk"    // 封装场景 sdk
)

const (
	InstanceCollection = "instances" // 存放实例的集合
	SecretCollection   = "secret"    // 存放密钥的集合
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
	Redis struct { // 无密码Redis
		Host    string `yaml:"host"`
		Port    int64  `yaml:"port"`
		ErrorDb int    `yaml:"error_db"` // 存放错误信息的数据库 recovery
	} `yaml:"redis"`
	DaoLog struct {
		Name  string `yaml:"dao_name"`
		Path  string `yaml:"dao_path"`
		Level string `yaml:"dao_level"`
		Count uint   `yaml:"count"`
	} `yaml:"log"`
	ServiceLog struct {
		Name  string `yaml:"service_name"`
		Path  string `yaml:"service_path"`
		Level string `yaml:"service_level"`
		Count uint   `yaml:"count"`
	} `yaml:"log"`
	AccessLog struct {
		Name  string `yaml:"access_name"`
		Path  string `yaml:"access_path"`
		Level string `yaml:"access_level"`
		Count uint   `yaml:"count"`
	} `yaml:"access_log"`
	DetailLog struct {
		Name  string `yaml:"detail_name"`
		Path  string `yaml:"detail_path"`
		Level string `yaml:"detail_level"`
		Count uint   `yaml:"count"`
	} `yaml:"detail_log"`
	ErrorLog struct {
		Name  string `yaml:"error_name"`
		Path  string `yaml:"error_path"`
		Level string `yaml:"error_level"`
		Count uint   `yaml:"count"`
	} `yaml:"error_log"`
	Expires int64 `yaml:"expires"` // 请求过期时间
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

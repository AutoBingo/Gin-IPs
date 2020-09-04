package test

import (
	"Gin-IPs/src/configure"
	"Gin-IPs/src/dao"
	"Gin-IPs/src/models"
	"testing"
)

//func init()  {
//	if err := configure.InitConfigValue(); err != nil { // 初始化配置
//		fmt.Println(err)
//		os.Exit(1)
//	}
//	if err := dao.Init(); err != nil { // 初始化配置
//		fmt.Println(err)
//		os.Exit(1)
//	}
//}

/*
secret 集合数据

{
    "access_key": "A00001",
    "secret_key": "SECRET-A00001",
    "user": "xiaoming",
    "state": "valid",
    "ctime": "2020-08-01 12:00:00"
}

*/

// 插入数据库数据
func TestInsertData(t *testing.T) {
	hmArr := []models.HostModel{
		{
			Oid:      configure.OidHost,
			Id:       "H001",
			Ip:       "10.1.162.18",
			Hostname: "10-1-162-18",
			MemSize:  1024000,
			DiskSize: 102400000000,
			Class:    "物理机",
			Owner:    []string{"小林"},
		},
		{
			Oid:      configure.OidHost,
			Id:       "H002",
			Ip:       "10.1.162.19",
			Hostname: "10-1-162-19",
			MemSize:  1024000,
			DiskSize: 102400000000,
			Class:    "虚拟机",
			Owner:    []string{"小黄"},
		},
	}
	_ = dao.InsertHost(hmArr)

	swArr := []models.SwitchModel{
		{
			Oid:           configure.OidSwitch,
			Id:            "S001",
			Name:          "上海集群交换机",
			Ip:            "10.2.32.11",
			Vip:           []string{"10.2.20.1", "10.2.20.13", "10.1.162.18"},
			ConsoleIp:     "10.3.32.11",
			Manufacturers: "华为",
			Owner:         []string{"老马", "老曹"},
		},
		{
			Oid:           configure.OidSwitch,
			Id:            "S002",
			Name:          "广州集群交换机",
			Ip:            "10.2.32.13",
			Vip:           []string{"10.2.21.5", "10.2.21.23", "10.2.21.40"},
			ConsoleIp:     "10.3.32.13",
			Manufacturers: "思科",
			Owner:         []string{"老马", "老曹"},
		},
	}
	_ = dao.InsertSwitch(swArr)
}

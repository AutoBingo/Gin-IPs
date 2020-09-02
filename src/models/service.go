package models

import "Gin-IPs/src/configure"

//|ID|主机名|IP|内存大小|磁盘大小|类型|负责人|

type HostModel struct {
	Oid      configure.Oid `json:"oid"` // 考虑所有实例存放在同一个集合中，需要一个字段来区分
	Id       string        `json:"id"`
	Ip       string        `json:"ip"`
	Hostname string        `json:"hostname"`
	MemSize  int64         `json:"mem_size"`
	DiskSize int64         `json:"disk_size"`
	Class    string        `json:"class"` // 主机类型
	Owner    []string      `json:"owner"`
}

//|ID|设备名|管理IP|虚IP|带外IP|厂家|负责人|

type SwitchModel struct {
	Oid           configure.Oid `json:"oid"` // 考虑所有实例存放在同一个集合中，需要一个字段来区分
	Id            string        `json:"id"`
	Name          string        `json:"name"`
	Ip            string        `json:"ip"`
	Vip           []string      `json:"vip"`
	ConsoleIp     string        `json:"console_ip"`
	Manufacturers string        `json:"manufacturers"` // 厂家
	Owner         []string      `json:"owner"`
}

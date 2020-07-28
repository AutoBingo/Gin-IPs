package service

type Host struct {
	Ip       string   `json:"ip"`
	Hostname string   `json:"hostname"`
	MemSize  int64    `json:"mem_size"`  // kB
	DiskSize int64    `json:"disk_size"` // KB
	Class    string   `json:"class"`     // 物理机/虚拟机
	Owner    []string `json:"owner"`     // 负责人，可能有多个
}

type Switch struct {
	Name         string   `json:"name"` // 设备名
	Ip           string   `json:"ip"`
	Vip          []string `json:"vip"`          // 虚IP，可能有多个
	ConsoleIp    string   `json:"console_ip"`   // 控制口IP
	Manufacturer string   `json:"manufacturer"` // 厂家
	Owner        []string `json:"owner"`        // 负责人数组
}

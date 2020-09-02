package dao

import (
	"Gin-IPs/src/configure"
	"Gin-IPs/src/models"
	"Gin-IPs/src/utils/database/mongodb"
	"encoding/json"
)

func insertInstances(instArr []interface{}) error {
	insMgo := mongodb.NewConnection(ModelClient.MgoPool, ModelClient.MgoDb, configure.InstanceCollection)
	defer insMgo.Close()
	if len(instArr) > 0 {
		_, err := insMgo.InsertMany(instArr)
		return err
	}
	return nil
}

func InsertHost(hostArr []models.HostModel) error {
	var documents []interface{}
	for _, host := range hostArr {
		if hostBytes, err := json.Marshal(host); err != nil {
			return err
		} else {
			var hm models.HostModel
			if err := json.Unmarshal(hostBytes, &hm); err != nil {
				return err
			} else {
				documents = append(documents, hm)
			}
		}
	}
	return insertInstances(documents)
}

func InsertSwitch(switchArr []models.SwitchModel) error {
	var documents []interface{}
	for _, sw := range switchArr {
		if hostBytes, err := json.Marshal(sw); err != nil {
			return err
		} else {
			var sm models.SwitchModel
			if err := json.Unmarshal(hostBytes, &sm); err != nil {
				return err
			} else {
				documents = append(documents, sm)
			}
		}
	}
	return insertInstances(documents)
}

func MockTest() {
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
	_ = InsertHost(hmArr)

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
	_ = InsertSwitch(swArr)
}

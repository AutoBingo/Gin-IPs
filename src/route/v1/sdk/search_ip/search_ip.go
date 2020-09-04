package v1_sdk_search_ip

import (
	"Gin-IPs/src/configure"
	"Gin-IPs/src/dao"
	"go.mongodb.org/mongo-driver/bson"
)

func SearchIp(ipArr []string, oid configure.Oid) ([]interface{}, error) {
	condition := bson.M{"$in": ipArr}
	// 这批ip都需要建索引,最后一个 name 是宿主机
	ipAttrs := []string{"ip", "console_ip", "ip2", "eth.ip", "network.ip", "mgmt", "vip", "vcenter_ip", "out_band_ip", "name"}
	var attrFilter = make([]bson.M, len(ipAttrs))
	for i, attr := range ipAttrs {
		attrFilter[i] = bson.M{attr: condition}
		//projection[attr] = true
	}
	var oidArr []configure.Oid
	if oid == "" {
		oidArr = []configure.Oid{configure.OidHost, configure.OidSwitch}
	} else {
		oidArr = []configure.Oid{oid}
	}
	filter := bson.D{
		{"oid", bson.M{"$in": oidArr}},
		{"$or", attrFilter},
	}
	var ipRes []interface{}
	result, err := dao.FetchAnyIns(filter, nil)
	if err != nil {
		return ipRes, err
	} else {
		for _, each := range result {
			ipRes = append(ipRes, each)
		}
	}
	return ipRes, nil
}

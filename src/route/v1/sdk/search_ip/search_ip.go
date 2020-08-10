package v1_sdk_search_ip

import (
	"Gin-IPs/src/configure"
	"Gin-IPs/src/dao"
	"Gin-IPs/src/route/request"
	"Gin-IPs/src/route/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strings"
)

var SearchIpHandlerWithGet = func(c *gin.Context) {
	response := route_response.Response{
		Code: configure.RequestSuccess,
		Data: route_response.ResponseData{List: []interface{}{}},
	}
	var params route_request.ReqGetParaSearchIp
	if err := c.ShouldBindQuery(&params); err != nil {
		code, msg := params.ParseError(err)
		response.Code, response.Message = code, msg
		c.JSON(http.StatusOK, response)
		return
	}
	ipArr := strings.Split(params.Ip, ",")
	if err := route_request.CheckIp(ipArr, 1, 10); err != nil {
		response.Code, response.Message = configure.RequestParameterRangeError, err.Error()
		c.JSON(http.StatusOK, response)
		return
	}
	if ipRes, err := SearchIp(ipArr, params.Oid); err != nil {
		response.Code, response.Message = configure.RequestOtherError, err.Error()
		c.JSON(http.StatusOK, response)
		return
	} else {
		size := int64(len(ipRes))
		response.Data = route_response.ResponseData{
			Page:     1,
			PageSize: size,
			Size:     size,
			Total:    size,
			List:     ipRes,
		}
		c.JSON(http.StatusOK, response)
	}
	return
}

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

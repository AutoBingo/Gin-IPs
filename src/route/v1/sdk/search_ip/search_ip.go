package v1_sdk

import (
	"Gin-IPs/src/configure"
	"Gin-IPs/src/route/request"
	"Gin-IPs/src/route/response"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

var SearchIpHandlerWithGet = func(c *gin.Context) {
	response := route_response.Response{
		Code:configure.RequestSuccess,
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
	hostInfo := map[string]interface{}{
		"10.1.162.18": map[string]string{
			"model": "主机", "IP": "10.1.162.18",
		},
	}
	response.Data = route_response.ResponseData{
		Page:     1,
		PageSize: 1,
		Size:     1,
		Total:    1,
		List:     []interface{}{hostInfo, },
	}
	c.JSON(http.StatusOK, response)
	return
}





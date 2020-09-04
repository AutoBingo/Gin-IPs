package v1_sdk

import (
	"Gin-IPs/src/configure"
	"Gin-IPs/src/route/request"
	"Gin-IPs/src/route/response"
	"Gin-IPs/src/route/v1/sdk/search_ip"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// Get /v1/sdk/search_ip?ip=''
var SearchIpFunc = func() (string, func(c *gin.Context)) {
	uri := fmt.Sprintf("/search_ip") // 对外 URI
	return uri, func(c *gin.Context) {
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
		if ipRes, err := v1_sdk_search_ip.SearchIp(ipArr, params.Oid); err != nil {
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
}

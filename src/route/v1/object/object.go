package v1_object

import (
	"Gin-IPs/src/configure"
	"Gin-IPs/src/route/response"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// POST  /v1/object/$oid
var AllInstancesPostFunc = func(oid configure.Oid) (string, func(c *gin.Context)) {
	uri := fmt.Sprintf("/%s", oid) // 对外 URI
	return uri, func(c *gin.Context) {
		response := route_response.Response{Code: configure.RequestSuccess, Data: route_response.ResponseData{List: []interface{}{}}}

		// ... 获取 switch 或者 host 的所有实例

		c.JSON(http.StatusOK, response) // 无论结果，返回OK
		return
	}
}

// Get /v1/object/$oid/$id
var SingleInstanceGetFunc = func(oid configure.Oid) (string, func(c *gin.Context)) {
	uri := fmt.Sprintf("/%s/:id", oid) // 对外 URI
	return uri, func(c *gin.Context) {
		response := route_response.Response{Code: configure.RequestSuccess, Data: route_response.ResponseData{List: []interface{}{}}}
		insId := c.Param("id")
		if insId == "" {
			response.Code, response.Message = configure.RequestParameterMiss, "请求参数缺少实例的instanceId"
			c.JSON(http.StatusOK, response)
			return
		}

		// ... 获取 switch 或者 host 的某一个实例

		c.JSON(http.StatusOK, response)
		return
	}
}

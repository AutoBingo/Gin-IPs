package route

import (
	"Gin-IPs/src/configure"
	"Gin-IPs/src/route/v1/object"
	"Gin-IPs/src/route/v1/sdk"
	"github.com/gin-gonic/gin"
)

type Route map[string]func(c *gin.Context) // key:uri路径, value: 中间件函数
type Method int

const (
	GET Method = iota
	POST
	DELETE
	PUT
)

// devPkg 对应的路由
type DevPkgGroup struct {
	Name   configure.DevPkg
	Routes []map[Method]Route
}

// 版本对应的路由
type Group struct {
	Version string
	PkgList []DevPkgGroup
}

var RGroups []Group

func InitRouteGroups() {
	RGroups = []Group{
		{"v1", // RGroups[0] 表示 V1， RGroups[1] 表示 V2
			[]DevPkgGroup{},
		},
	}

	/*---------- 更新 V1 路由 ----------*/

	// Object 路由，根据oid遍历多个
	var objectRoutes []map[Method]Route
	for _, oid := range configure.OidArray {
		uri, postFunc := v1_object.AllInstancesPostFunc(oid) // POST  /v1/object/$oid
		objectRoutes = append(objectRoutes, map[Method]Route{POST: {uri: postFunc}})

		uri, getFunc := v1_object.SingleInstanceGetFunc(oid) // GET /v1/object/$oid/$id
		objectRoutes = append(objectRoutes, map[Method]Route{GET: {uri: getFunc}})
	}
	RGroups[0].PkgList = append(RGroups[0].PkgList, DevPkgGroup{configure.ObjectPkg, objectRoutes})

	// Sdk 路由
	var sdkRoutes []map[Method]Route
	// Sdk Get 路由
	sdkGetFuncArr := []func() (string, func(c *gin.Context)){
		v1_sdk.SearchIpFunc, // Get /v1/sdk/search_ip?ip='xxx'
	}

	for _, sdkGetFunc := range sdkGetFuncArr {
		sdkGetUri, sdkGetFunc := sdkGetFunc()
		sdkRoutes = append(sdkRoutes, map[Method]Route{GET: {sdkGetUri: sdkGetFunc}})
	}
	RGroups[0].PkgList = append(RGroups[0].PkgList, DevPkgGroup{configure.SdkPkg, sdkRoutes})
}

func methodMapper(group *gin.RouterGroup, method Method) func(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	if method == GET {
		return group.GET
	}
	if method == POST {
		return group.POST
	}
	if method == DELETE {
		return group.DELETE
	}
	if method == PUT {
		return group.PUT
	}
	return group.Any
}

// 路由解析
func AddRoute(app *gin.Engine) {
	cmdbGroup := app.Group("/")
	for _, group := range RGroups {
		versionGroup := cmdbGroup.Group(group.Version)
		for _, sdk := range group.PkgList {
			sdkGroup := versionGroup.Group(string(sdk.Name))
			for _, mapper := range sdk.Routes {
				for method, route := range mapper {
					for uri, handler := range route {
						methodMapper(sdkGroup, method)(uri, handler)
					}
				}
			}
		}
	}
}

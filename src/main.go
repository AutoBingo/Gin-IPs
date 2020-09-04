package main

import (
	"Gin-IPs/src/configure"
	"Gin-IPs/src/dao"
	"Gin-IPs/src/route"
	route_middleware "Gin-IPs/src/route/middleware"
	"Gin-IPs/src/route/request"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := configure.InitConfigValue(); err != nil { // 初始化配置
		fmt.Println(err)
		os.Exit(1)
	}
	if err := dao.Init(); err != nil { // 初始化配置
		fmt.Println(err)
		os.Exit(1)
	}
	dao.Start()

	if configure.GinConfigValue.ApiServer.Env == "prod" {
		gin.SetMode(gin.ReleaseMode) // 生产模式
	}

	r := gin.New() // 不用默认的日志中间件
	route_request.InitValidator()

	r.Use(route_middleware.Logger())
	r.Use(route_middleware.Validate())
	// route.Use(gin.Recovery())  // 内置的 recovery
	r.Use(route_middleware.Recovery())  // 放在鉴权后面可以得到 user

	route.InitRouteGroups()
	route.AddRoute(r)  // 动态路由实现

	addr := fmt.Sprintf("%s:%d", configure.GinConfigValue.ApiServer.Host,
		configure.GinConfigValue.ApiServer.Port)
	if err := r.Run(addr); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

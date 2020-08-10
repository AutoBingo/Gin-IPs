package main

import (
	"Gin-IPs/src/configure"
	"Gin-IPs/src/dao"
	"Gin-IPs/src/route/request"
	"Gin-IPs/src/route/v1/sdk/search_ip"
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
	route := gin.Default()
	route_request.InitValidator()
	route.GET("/", v1_sdk_search_ip.SearchIpHandlerWithGet)
	if err := route.Run("127.0.0.1:8080"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

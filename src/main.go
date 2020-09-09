package main

import (
	"Gin-IPs/src/configure"
	"Gin-IPs/src/dao"
	"Gin-IPs/src/service"
	"Gin-IPs/src/utils/daemon"
	"flag"
	"fmt"
	"os"
	"runtime"
)

// 是否平滑启动， 该参数应该只在 reload中生效，第一次不应该使用
// kill -USR2 $pid 可以平滑重启
var RunGraceful = flag.Bool("graceful", false, "reload graceful")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	if err := configure.InitConfigValue(); err != nil { // 初始化配置
		fmt.Println(err)
		os.Exit(1)
	}
	daemon.InitProcess()               // 启动守护进程
	if err := dao.Init(); err != nil { // 初始化配置
		fmt.Println(err)
		os.Exit(1)
	}
	dao.Start()

	if server, err := service.NewServer(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		if err := server.Start(*RunGraceful); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

}

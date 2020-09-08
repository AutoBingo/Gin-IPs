package service

import (
	"Gin-IPs/src/configure"
	"Gin-IPs/src/route"
	"Gin-IPs/src/route/middleware"
	"Gin-IPs/src/route/request"
	"Gin-IPs/src/utils/log"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net"
)

type Server struct {
	Host     string
	Port     int64
	Router   *gin.Engine
	Logger   *logrus.Logger
	Listener net.Listener
}

func NewServer() (*Server, error) {
	server := new(Server)
	if logger, err := mylog.New(
		configure.GinConfigValue.ServiceLog.Path, configure.GinConfigValue.ServiceLog.Name,
		configure.GinConfigValue.ServiceLog.Level, nil, configure.GinConfigValue.ServiceLog.Count); err != nil {
		return server, err
	} else {
		server.Logger = logger
	}
	server.Host = configure.GinConfigValue.ApiServer.Host
	server.Port = configure.GinConfigValue.ApiServer.Port
	if configure.GinConfigValue.ApiServer.Env == "prod" {
		gin.SetMode(gin.ReleaseMode) // 生产模式
	}
	server.Router = gin.New() // 不用gin.Default避免使用默认的日志打印和异常捕捉
	route_request.InitValidator()
	return server, nil
}

// 监听请求、监听信号
func (server *Server) Start(graceful bool) error {
	server.Router.Use(route_middleware.Logger())
	server.Router.Use(route_middleware.Validate())
	server.Router.Use(route_middleware.Recovery()) // 放在鉴权后面可以得到 user
	//server.Router.NoMethod(route_middleware.MethodNotAllow())
	//server.Router.NoRoute(route_middleware.NotFound())
	route.InitRouteGroups()       // 初始化路由组
	route.AddRoute(server.Router) // 动态路由实现

	if err := server.Listen(graceful); err != nil {
		server.Logger.Error(err)
		return err
	}
	return nil
}

# 项目介绍
本项目用`golang`创建了一个API服务，代码中包含完整的程序和配置文件。如果有需要使用`Gin`开发后台API服务的，可以直接`Clone`，也欢迎各路大佬一起完善代码。

# 编译运行
* 编译
```bash
cd Gin-IPs
go build src/main.go
```
* 更新配置
> 根据`conf/gin_ips.yaml`的`mgo`配置和`redis`配置新建mongodb数据库和redis，并确认服务正常运行

* 生成测试数据
```bash
cp -r conf test/conf 
cd test
go test -v mock_test.go 
```
* 运行
```bash
cd ..
./main
```
* 测试
```bash
cd test
go test -v request_test.go gin-api.go
```
* 平滑重启
```bash
# ps -ef |grep main
root      9142     1  1 13:26 ?        00:00:03 ./main
root      9667 21410  0 13:29 pts/0    00:00:00 grep main

# kill -USR2 9142

# ps -ef |grep main
root      9668     1 99 13:29 ?        00:00:02 ./main -graceful
root      9682 21410  0 13:29 pts/0    00:00:00 grep main

# go test -v request_test.go gin-api.go 
```
# Gin-API系列文章
* [【Gin-API系列】需求设计和功能规划（一）](https://www.cnblogs.com/lxmhhy/p/13385475.html)
* [【Gin-API系列】请求和响应参数的检查绑定（二）](https://www.cnblogs.com/lxmhhy/p/13385482.html)
* [【Gin-API系列】配置文件和数据库操作（三）](https://www.cnblogs.com/lxmhhy/p/13471256.html)
* [【Gin-API系列】Gin中间件之日志模块（四）](https://www.cnblogs.com/lxmhhy/p/13518211.html)
* [【Gin-API系列】Gin中间件之鉴权访问（五）](https://www.cnblogs.com/lxmhhy/p/13603330.html)
* [【Gin-API系列】Gin中间件之异常处理（六）](https://www.cnblogs.com/lxmhhy/p/13608517.html)
* [【Gin-API系列】实现动态路由分组（七）](https://www.cnblogs.com/lxmhhy/p/13614097.html)
* [【Gin-API系列】守护进程和平滑重启（八）](https://www.cnblogs.com/lxmhhy/p/13633581.html)

package route_middleware

import (
	"Gin-IPs/src/configure"
	"Gin-IPs/src/dao"
	"Gin-IPs/src/route/response"
	"Gin-IPs/src/utils/log"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime/debug"
	"time"
)

// 日志打印没必要异步处理，一般crash比较少
func Recovery() gin.HandlerFunc {
	log, _ := mylog.New(
		configure.GinConfigValue.ErrorLog.Path, configure.GinConfigValue.ErrorLog.Name,
		configure.GinConfigValue.ErrorLog.Level, nil, configure.GinConfigValue.ErrorLog.Count)
	log.Info("Test Panic")
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				response := route_response.Response{}
				response.Data.List = []interface{}{} // 初始化为空切片，而不是空引用
				traceId := c.Writer.Header().Get("X-Request-Trace-Id")
				stackMsg := string(debug.Stack())
				logField := map[string]interface{}{
					"trace_id":    traceId, //  鉴权之后可以得到唯一跟踪ID和用户名
					"user":        c.Writer.Header().Get("X-Request-User"),
					"uri":         c.Request.URL.Path,
					"remote_addr": c.ClientIP(),
					"stack":       stackMsg, // 打印堆栈信息
				}
				c.Abort()
				response.Code, response.Message = configure.ApiInnerResponseError, fmt.Sprintf("Api内部报错，请联系管理员(id=%s", traceId)
				log.WithFields(logField).Error(err) // 输出panic 信息
				redisField := make(map[string]interface{})
				for k, v := range logField {
					redisField[k] = v
				}
				redisField["time"] = time.Now().Format("2006-01-02 15:04:05")
				redisField["error"] = err
				dao.ModelClient.RedisClient.HMSet(traceId, redisField) // 上报redis
				c.JSON(http.StatusUnauthorized, response)
				return
			}
		}()

		c.Next()
	}
}

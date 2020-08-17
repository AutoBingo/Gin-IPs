package route_middleware

import (
	"Gin-IPs/src/configure"
	"Gin-IPs/src/route/response"
	"Gin-IPs/src/utils/log"
	"Gin-IPs/src/utils/uuid"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"time"
)

var logChannel = make(chan map[string]interface{}, 300)

func logHandlerFunc() {
	accessLog, _ := mylog.New(
		configure.GinConfigValue.AccessLog.Path, configure.GinConfigValue.AccessLog.Name,
		configure.GinConfigValue.AccessLog.Level, nil, configure.GinConfigValue.AccessLog.Count)
	detailLog, _ := mylog.New(
		configure.GinConfigValue.DetailLog.Path, configure.GinConfigValue.DetailLog.Name,
		configure.GinConfigValue.DetailLog.Level, nil, configure.GinConfigValue.DetailLog.Count)
	for logField := range logChannel {
		var (
			msgStr    string
			levelStr  string
			detailStr string
		)
		if msg, ok := logField["msg"]; ok {
			msgStr = msg.(string)
			delete(logField, "msg")
		}
		if level, ok := logField["level"]; ok {
			levelStr = level.(string)
			delete(logField, "level")
		}
		if detail, ok := logField["detail"]; ok {
			detailStr = detail.(string)
			delete(logField, "detail")
		}
		accessLog.WithFields(logField).Info("Request Finished")
		if "info" == levelStr {
			detailLog.WithFields(logField).Info(detailStr)
			detailLog.WithFields(logField).Info(msgStr)
		} else {
			detailLog.WithFields(logField).Error(detailStr)
			detailLog.WithFields(logField).Error(msgStr)
		}
	}
}

type BodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w BodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
func (w BodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

var SnowWorker, _ = uuid.NewSnowWorker(100) // 随机生成一个uuid，100是节点的值（随便给一个）

// 打印日志
func Logger() gin.HandlerFunc {
	go logHandlerFunc()
	return func(c *gin.Context) {
		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		requestBody, _ := ioutil.ReadAll(tee)
		c.Request.Body = ioutil.NopCloser(&buf)

		user := c.Writer.Header().Get("X-Request-User")
		bodyLogWriter := &BodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bodyLogWriter

		start := time.Now()

		c.Next()

		responseBody := bodyLogWriter.body.Bytes()
		response := route_response.Response{}
		if len(responseBody) > 0 {
			_ = json.Unmarshal(responseBody, &response)
		}
		end := time.Now()
		responseTime := float64(end.Sub(start).Nanoseconds()) / 1000000.0 // 纳秒转毫秒才能保留小数
		logField := map[string]interface{}{
			"user":            user,
			"uri":             c.Request.URL.Path,
			"raw_query":       c.Request.URL.RawQuery,
			"start_timestamp": start.Format("2006-01-02 15:04:05"),
			"end_timestamp":   end.Format("2006-01-02 15:04:05"),
			"server_name":     c.Request.Host,
			"server_addr": fmt.Sprintf("%s:%d", configure.GinConfigValue.ApiServer.Host,
				configure.GinConfigValue.ApiServer.Port), // 无法动态读取
			"remote_addr":    c.ClientIP(),
			"proto":          c.Request.Proto,
			"referer":        c.Request.Referer(),
			"request_method": c.Request.Method,
			"response_time":  fmt.Sprintf("%.3f", responseTime), // 毫秒
			"content_type":   c.Request.Header.Get("Content-Type"),
			"status":         c.Writer.Status(),
			"user_agent":     c.Request.UserAgent(),
			"trace_id":       SnowWorker.GetId(),
		}
		logField["detail"] = string(requestBody)

		if response.Code != configure.RequestSuccess {
			logField["msg"], logField["level"] = fmt.Sprintf("code=%d, message=%s", response.Code, response.Message), "error"
		} else {
			logField["msg"], logField["level"] = fmt.Sprintf("total=%d, page_size=%d, page=%d, size=%d",
				response.Data.Total, response.Data.PageSize, response.Data.Page, response.Data.Size), "info"
		}
		logChannel <- logField
	}
}

package mylog

import (
	"errors"
	"io"
	"os"
	"path"
	"time"

	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

var LevelMap = map[string]logrus.Level{
	"DEBUG": logrus.DebugLevel,
	"ERROR": logrus.ErrorLevel,
	"WARN":  logrus.WarnLevel,
	"INFO":  logrus.InfoLevel,
}

// 创建 @filePth: 如果路径不存在会创建 @fileName: 如果存在会被覆盖  @std: os.stdout/stderr 标准输出和错误输出
func New(filePath string, fileName string, level string, std io.Writer, count uint) (*logrus.Logger, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.MkdirAll(filePath, 755); err != nil {
			return nil, err
		}
	}
	fn := path.Join(filePath, fileName)

	logger := logrus.New()
	//timeFormatter := &logrus.TextFormatter{
	//	FullTimestamp:   true,
	//	TimestampFormat: "2006-01-02 15:04:05.999999999",
	//}
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05.999999999",
	}) // 设置日志格式为json格式

	if logLevel, ok := LevelMap[level]; !ok {
		return nil, errors.New("log level not found")
	} else {
		logger.SetLevel(logLevel)
	}

	//logger.SetFormatter(timeFormatter)

	/*  根据文件大小分割日志
	// import "gopkg.in/natefinch/lumberjack.v2"
	logger := &lumberjack.Logger{
		// 日志输出文件路径
		Filename:   "D:\\test_go.log",
		// 日志文件最大 size, 单位是 MB
		MaxSize:    500, // megabytes
		// 最大过期日志保留的个数
		MaxBackups: 3,
		// 保留过期文件的最大时间间隔,单位是天
		MaxAge:     28,   //days
		// 是否需要压缩滚动日志, 使用的 gzip 压缩
		Compress:   true, // disabled by default
	}
	*/
	if 0 == count {
		count = 90 // 0的话则是默认保留90天
	}
	logFd, err := rotatelogs.New(
		fn+".%Y-%m-%d",
		// rotatelogs.WithLinkName(fn),
		//rotatelogs.WithMaxAge(time.Duration(24*count)*time.Hour),
		rotatelogs.WithRotationTime(time.Duration(24)*time.Hour),
		rotatelogs.WithRotationCount(count),
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = logFd.Close() // don't need handle error
	}()

	if nil != std {
		logger.SetOutput(io.MultiWriter(logFd, std)) // 设置日志输出
	} else {
		logger.SetOutput(logFd) // 设置日志输出
	}
	// logger.SetReportCaller(true)   // 测试环境可以开启，生产环境不能开，会增加很大开销
	return logger, nil
}

package dao

import (
	"Gin-IPs/src/configure"
	"Gin-IPs/src/utils/database/mongodb"
	"Gin-IPs/src/utils/log"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

type Models struct {
	Logger      *logrus.Logger
	MgoPool     *mongo.Client
	MgoDb       string
	RedisClient *redis.Client
}

var ModelClient = new(Models)

func Init() error {
	if logger, err := mylog.New(
		configure.GinConfigValue.DaoLog.Path, configure.GinConfigValue.DaoLog.Name,
		configure.GinConfigValue.DaoLog.Level, nil, configure.GinConfigValue.DaoLog.Count); err != nil {
		return err
	} else {
		ModelClient.Logger = logger
	}
	var err error
	ModelClient.MgoPool, err = mongodb.CreatePool(configure.GinConfigValue.Mgo.Uri, configure.GinConfigValue.Mgo.PoolSize)
	if err != nil {
		ModelClient.Logger.Errorf("Collection Client Pool With Uri %s Create Failed: %s", configure.GinConfigValue.Mgo.Uri, err)
		return err
	}
	ModelClient.MgoDb = configure.GinConfigValue.Mgo.Database
	ModelClient.Logger.Infof("Collection Client Pool Created successful With Uri %s", configure.GinConfigValue.Mgo.Uri)

	redisAddr := fmt.Sprintf("%s:%d", configure.GinConfigValue.Redis.Host, configure.GinConfigValue.Redis.Port)
	ModelClient.RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       configure.GinConfigValue.Redis.ErrorDb,
	})
	_, err = ModelClient.RedisClient.Ping().Result()
	if err != nil {
		ModelClient.Logger.Errorf("Redis Client With Addr %s Create Failed: %s", redisAddr, err)
		return err
	}
	ModelClient.Logger.Errorf("Redis Client With Addr %s Create Successful", redisAddr)

	ModelClient.Logger.Infof("Models Created Success")
	return nil
}

func (m *Models) LogMongo() {
	for log := range mongodb.MongoLogChannel {
		logField := map[string]interface{}{
			"Database":   log.Database,
			"Collection": log.Collection,
			"Action":     log.Action,
			"Result":     log.Result,
		}
		switch log.Documents.(type) {
		case map[string]interface{}:
			docBytes, _ := json.Marshal(log.Documents)
			logField["Documents"] = string(docBytes)
		default:
			logField["Documents"] = log.Documents
		}
		if log.Ok {
			m.Logger.WithFields(logField).Info("")
		} else {
			m.Logger.WithFields(logField).Error(log.ErrMsg)
		}
	}
}

func Start() {
	go ModelClient.LogMongo()
	// MockTest()  // 插入初始化数据
}

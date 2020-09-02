package dao

import (
	"Gin-IPs/src/configure"
	"Gin-IPs/src/models"
	"Gin-IPs/src/utils/database/mongodb"
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
)

// 返回的是Mongo所有匹配的数据
func FetchAnyIns(filter bson.D, projection bson.M) ([]map[string]interface{}, error) {
	insMgo := mongodb.NewConnection(ModelClient.MgoPool, ModelClient.MgoDb, configure.InstanceCollection)
	defer insMgo.Close()
	result, err := insMgo.FindAll(filter, nil, projection)
	var anything []map[string]interface{}
	if err != nil {
		return anything, err
	}
	for _, each := range result {
		delete(each, "_id") // 删除_id 字段
		anything = append(anything, each)
	}
	return anything, nil
}

func FetchSecret(accessKey string) (models.Secret, error) {
	secretMgo := mongodb.NewConnection(ModelClient.MgoPool, ModelClient.MgoDb, configure.SecretCollection)
	defer secretMgo.Close()
	filter := bson.D{{"access_key", accessKey}}
	result, err := secretMgo.FindOne(filter, nil, nil)
	secret := models.Secret{}
	if err != nil {
		return secret, errors.New("获取不到Secret Key")
	}
	secretBytes, _ := json.Marshal(result)
	_ = json.Unmarshal(secretBytes, &secret)
	return secret, nil
}

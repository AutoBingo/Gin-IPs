package dao

import (
	"Gin-IPs/src/configure"
	"Gin-IPs/src/models"
	"Gin-IPs/src/utils/database/mongodb"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
)

func insertInstances(instArr []interface{}) error {
	insMgo := mongodb.NewConnection(ModelClient.MgoPool, ModelClient.MgoDb, configure.InstanceCollection)
	defer insMgo.Close()
	if len(instArr) > 0 {
		_, err := insMgo.InsertMany(instArr)
		return err
	}
	return nil
}

func InsertHost(hostArr []models.HostModel) error {
	var documents []interface{}
	for _, host := range hostArr {
		if hostBytes, err := json.Marshal(host); err != nil {
			return err
		} else {
			var hm bson.M
			if err := json.Unmarshal(hostBytes, &hm); err != nil {
				return err
			} else {
				documents = append(documents, hm)
			}
		}
	}
	return insertInstances(documents)
}

func InsertSwitch(switchArr []models.SwitchModel) error {
	var documents []interface{}
	for _, sw := range switchArr {
		if hostBytes, err := json.Marshal(sw); err != nil {
			return err
		} else {
			var sm bson.M
			if err := json.Unmarshal(hostBytes, &sm); err != nil {
				return err
			} else {
				documents = append(documents, sm)
			}
		}
	}
	return insertInstances(documents)
}

func InsertSecret(secret models.Secret) error {
	mgo := mongodb.NewConnection(ModelClient.MgoPool, ModelClient.MgoDb, configure.SecretCollection)
	defer mgo.Close()
	if secretBytes, err := json.Marshal(secret); err != nil {
		return err
	} else {
		var sm bson.M
		if err := json.Unmarshal(secretBytes, &sm); err != nil {
			return err
		} else {
			_, err := mgo.InsertOne(sm)
			return err
		}
	}
}

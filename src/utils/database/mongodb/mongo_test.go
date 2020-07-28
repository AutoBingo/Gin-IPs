package mongodb

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
	"time"
)

var pool *mongo.Client

func init() {
	uri := "mongodb://cmdb:cmdb123456@10.2.145.89:28010"
	var err error
	if pool, err = CreatePool(uri, 10); err != nil {
		panic(err)
	}
}

//type Entity struct {
//	Id string `bson:"_id,omitempty"`
//	A  string `bson:"a"`
//	B  string `bson:"b"`
//}
type Entity struct {
	Id primitive.ObjectID `bson:"_id,omitempty"`
	A  interface{}
	B  string
}

func TestMongo_InsertOne(t *testing.T) {
	mgo := NewConnection(pool, "cmdbtest", "test1")
	defer mgo.Close()
	t.Log(mgo.InsertOne(bson.M{"a": "世界", "b": "你好!!!非常好"}))
	t.Log(mgo.InsertOne(bson.M{"a": 2345, "b": "你好非常好"}))
}

func TestMongo_InsertMany(t *testing.T) {
	mgo := NewConnection(pool, "cmdbtest", "test1")
	defer mgo.Close()
	documents := []bson.M{{"a": "新世界", "b": "你们非常好"}, {"a": "旧世界", "b": "你们很厉害"}}
	t.Log(mgo.InsertMany(documents))
}

func TestMongo_FindOne(t *testing.T) {
	mgo := NewConnection(pool, "cmdbtest", "test1")
	defer mgo.Close()
	var elem Entity
	filter := bson.D{{"a", map[string]bool{"$exists": true}}}
	if res, err := mgo.FindOne(filter, -1); err != nil {
		t.Fatal(err)
	} else {
		bsonString, _ := bson.Marshal(res)
		_ = bson.Unmarshal(bsonString, &elem)
		t.Log(elem.Id.Hex())
		t.Log(elem.A)
		t.Log(elem.B)
	}
}

func TestMongo_FindLimit(t *testing.T) {
	mgo := NewConnection(pool, "cmdbtest", "test1")
	defer mgo.Close()
	// filter := bson.D{{"a", "世界"}}
	filter := bson.D{{"a", map[string]bool{"$exists": true}}}
	if bsonArray, err := mgo.FindLimit(filter, 3, 1, -1, bson.M{}); err != nil {
		t.Fatal(err)
	} else {
		for _, each := range bsonArray {
			var elem Entity
			bsonString, _ := bson.Marshal(each)
			_ = bson.Unmarshal(bsonString, &elem)
			t.Log(elem.Id.Hex(), elem.A, elem.B)
		}
	}
}

func TestMongo_FindAll(t *testing.T) {
	mgo := NewConnection(pool, "cmdbtest", "attrs")
	defer mgo.Close()
	// filter := bson.D{{"a", "世界"}}
	// filter := bson.D{{"a", map[string]bool{"$exists": true}}}
	filter := bson.D{{"relation.HOST", bson.M{"$exists": true}}}
	if bsonArray, err := mgo.FindAll(filter, -1, nil); err != nil {
		t.Fatal(err)
	} else {
		for _, each := range bsonArray {
			//var elem Entity
			t.Log(each)
			// bsonString, _ := bson.Marshal(each)
			// t.Log(string(bsonString))
			//_ = bson.Unmarshal(bsonString, &elem)
			//t.Log(elem.Id.Hex(), elem.A, elem.B)
		}
	}
}

func TestMongo_CollectionCount(t *testing.T) {
	mgo := NewConnection(pool, "cmdbtest", "test1")
	defer mgo.Close()
	t.Log(mgo.CollectionCount())
}

func TestMongo_DeleteOne(t *testing.T) {
	// mgo := NewConnection(pool, "cmdbtest", "test1")
	mgo := NewConnection(pool, "cmdbtest", "instances")
	defer mgo.Close()
	// filter := bson.D{{"a", "世界"}}
	var elem Entity
	//filter := bson.D{{"a", "中国"}}
	// filter := bson.D{{"a", map[string]bool{"$exists": true}}}
	filter := bson.D{{"HOST", bson.M{"$elemMatch": bson.M{"instanceId": "5d6f8b47e4768"}}}}
	if res, err := mgo.DeleteAll(filter); err != nil {
		t.Fatal(err)
	} else {
		bsonString, _ := bson.Marshal(res)
		_ = bson.Unmarshal(bsonString, &elem)
		t.Log(elem.Id.Hex())
		t.Log(elem.A)
		t.Log(elem.B)
	}
}

func TestMongo_ParseId(t *testing.T) {
	mgo := NewConnection(pool, "cmdbtest", "test1")
	defer mgo.Close()
	var elem Entity
	filter := bson.D{{"a", map[string]bool{"$exists": true}}}
	if res, err := mgo.FindOne(filter, -1); err != nil {
		t.Fatal(err)
	} else {
		bsonString, _ := bson.Marshal(res)
		_ = bson.Unmarshal(bsonString, &elem)
		idStr := elem.Id.Hex()
		t.Log(idStr)
		t.Log(mgo.ParseId(idStr))
		loc, _ := time.LoadLocation("Asia/Shanghai")
		t.Log(elem.Id.Timestamp().Format("2006-01-02 15:04:05"))
		t.Log(elem.Id.Timestamp().In(loc)) // 上海时区
		t.Log(elem.Id.String())
		t.Log(elem.Id.IsZero())
	}
}

func TestMongo_Update(t *testing.T) {
	mgo := NewConnection(pool, "cmdbtest", "instances")
	defer mgo.Close()
	//filter := bson.D{{"a", map[string]bool{"$exists": true}}}
	//fmt.Println(mgo.FindAll(filter, -1))
	////fmt.Println(mgo.UpdateOne(filter, bson.M{"$set": bson.M{"a": "新的a4", "b": "新的b4"}}, true))
	//// fmt.Println(mgo.UpdateMany(filter, bson.M{"$set": bson.M{"a": "新的a8", "b": "新的b8"}}, true))
	//fmt.Println(mgo.Replace(filter, bson.M{"a": "新的a28", "b": "新的b28"}, true))
	instanceId := "5c1c93a484311"
	filter := bson.D{{"_owner_HOST", bson.M{"$elemMatch": bson.M{"instanceId": instanceId}}}}
	//res, _ := mgo.FindAll(filter, -1)
	//for _, each := range res {
	//}
	action := bson.M{"$pull": bson.M{"_owner_HOST": bson.M{"instanceId": instanceId}}}
	t.Log(mgo.UpdateMany(filter, action, false))
}

/*
Mongo连接池使用
primitive.ObjectIDFromHex(objectIdString)  将字符串Id转成 ObjectId

type Entity struct {
	Id primitive.ObjectID `bson:"_id,omitempty"`
	A  string
	B  string
}

输出Id字符串： entity.Id.Hex()
sort 只能是 1 或者 -1

*/
package mongodb

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"time"
)

var MongoLogChannel = make(chan MongoLog, 3000) // 最好能订阅消费

type MongoLog struct {
	Database   string      // 数据库
	Collection string      // 集合
	Action     string      // 动作
	Documents  interface{} // 文档集合
	Result     string      // 结果拼装
	Ok         bool        // 是否报错
	ErrMsg     string      // 错误信息
}

func Log(ml MongoLog) {
	select {
	case MongoLogChannel <- ml:
	case <-time.After(1 * time.Millisecond):
		return
	}
	return
}

type Collection struct {
	client     *mongo.Collection
	database   string // 数据库
	collection string // 集合
}

// 连接池创建
func CreatePool(uri string, size uint64) (pool *mongo.Client, e error) {
	defer func() {
		if err := recover(); err != nil {
			e = errors.New(fmt.Sprintf("%v", err))
			Log(MongoLog{Database: "", Collection: "", Action: "Create Pool", ErrMsg: e.Error(),
				Result: fmt.Sprintf("URI=%s, SIZE=%d", uri, size), Ok: false})
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 10s超时
	defer cancel()

	var err error
	pool, err = mongo.Connect(ctx, options.Client().ApplyURI(uri).SetMinPoolSize(size)) // 连接池
	if err != nil {
		Log(MongoLog{Database: "", Collection: "", Action: "Create Pool", ErrMsg: err.Error(),
			Result: fmt.Sprintf("URI=%s, SIZE=%d", uri, size), Ok: false})
		return pool, err
	}
	err = pool.Ping(context.Background(), nil) // 检查连接
	if err != nil {
		Log(MongoLog{Database: "", Collection: "", Action: "Create Pool Ping Check", ErrMsg: err.Error(),
			Result: fmt.Sprintf("URI=%s, SIZE=%d", uri, size), Ok: false})
		return pool, err
	}
	Log(MongoLog{Database: "", Collection: "", Action: "Create Pool", ErrMsg: "",
		Result: fmt.Sprintf("URI=%s, SIZE=%d", uri, size), Ok: true})
	return pool, nil
}

// 连接池销毁
func DestroyPool(client *mongo.Client) error {
	err := client.Disconnect(context.Background())
	if err != nil {
		Log(MongoLog{Database: "", Collection: "", Action: "Destroy Pool", ErrMsg: err.Error(), Result: "", Ok: false})
		return err
	}
	return nil
}

// 初始化db连接
func NewConnection(pool *mongo.Client, database, collection string) *Collection {
	mgo := new(Collection)
	mgo.database = database
	mgo.collection = collection
	mgo.client = pool.Database(mgo.database).Collection(mgo.collection)
	return mgo
}

// 关闭连接
func (m *Collection) Close() {
	// m.client.Drop(nil)  // drop 后连接池就变小
	m.client = nil
	m.database = ""
	m.collection = ""
}

//插入单个
func (m *Collection) InsertOne(document interface{}) (primitive.ObjectID, error) {
	insertResult, err := m.client.InsertOne(context.Background(), document)
	var objectId primitive.ObjectID
	if err != nil {
		Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Insert One",
			Documents: document, ErrMsg: err.Error(),
			Result: fmt.Sprintf("%v", objectId), Ok: false})
		return objectId, err
	}
	objectId = insertResult.InsertedID.(primitive.ObjectID)
	Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Insert One",
		Documents: document, ErrMsg: "",
		Result: fmt.Sprintf("%v", objectId), Ok: true})
	return objectId, nil
}

//插入多个文档
func (m *Collection) InsertMany(documents []interface{}) ([]primitive.ObjectID, error) {
	var insertDocs []interface{}
	for _, doc := range documents {
		insertDocs = append(insertDocs, doc)
	}
	insertResult, err := m.client.InsertMany(context.Background(), insertDocs)
	var objectIds []primitive.ObjectID
	if err != nil {
		Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Insert Many",
			Documents: fmt.Sprintf("length=%d", len(documents)), ErrMsg: err.Error(),
			Result: fmt.Sprintf("%v", objectIds), Ok: false})
		return objectIds, err
	}
	for _, oid := range insertResult.InsertedIDs {
		objectIds = append(objectIds, oid.(primitive.ObjectID))
	}
	Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Insert Many",
		Documents: fmt.Sprintf("length=%d", len(documents)), ErrMsg: "",
		Result: fmt.Sprintf("%v", objectIds), Ok: true})
	return objectIds, nil
}

// 查找单个文档, sort 等于1表示 返回最旧的，sort 等于-1 表示返回最新的
/*
BSON(二进制编码的JSON) D家族  bson.D
D：一个BSON文档。这种类型应该在顺序重要的情况下使用，比如MongoDB命令。
M：一张无序的map。它和D是一样的，只是它不保持顺序。
A：一个BSON数组。
E：D里面的一个元素。
*/
func (m *Collection) FindOne(filter bson.D, sort, projection bson.M) (bson.M, error) {
	findOptions := options.FindOne().SetProjection(projection)
	if sort != nil {
		findOptions = findOptions.SetSort(sort)
	}
	singleResult := m.client.FindOne(context.Background(), filter, findOptions)
	var result bson.M
	if err := singleResult.Decode(&result); err != nil {
		Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Find One",
			Documents: map[string]interface{}{"filter": filter, "sort": sort}, ErrMsg: err.Error(),
			Result: "", Ok: false})
		return result, err
	}
	if result == nil {
		Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Find One",
			Documents: map[string]interface{}{"filter": filter, "sort": sort}, ErrMsg: "",
			Result: "empty", Ok: true})
	} else {
		Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Find One",
			Documents: map[string]interface{}{"filter": filter, "sort": sort}, ErrMsg: "",
			Result: "one result", Ok: true})
	}
	return result, nil
}

/*
查询多个 sort 等于1表示 返回最旧的，sort 等于-1 表示返回最新的
每次只返回1页 page size 大小的数据
project 不能混合 True 和 False
*/
func (m *Collection) FindLimit(filter bson.D, page, pageSize uint64, sort, projection bson.M) ([]bson.M, error) {
	var resultArray []bson.M
	if page == 0 || pageSize == 0 {
		return resultArray, errors.New("page or page size can't be 0")
	}
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)
	if projection == nil {
		projection = bson.M{}
	}
	findOptions := options.Find().SetProjection(projection).SetSkip(skip).SetLimit(limit)
	if sort != nil {
		findOptions = findOptions.SetSort(sort)
	}
	cur, err := m.client.Find(context.Background(), filter, findOptions)
	if err != nil {
		Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Find Limit",
			Documents: map[string]interface{}{"filter": filter, "sort": sort, "page": page, "pageSize": pageSize},
			ErrMsg:    err.Error(), Result: fmt.Sprintf("length=%d", len(resultArray)), Ok: false})
		return resultArray, err
	}
	defer func() {
		_ = cur.Close(context.Background())
	}()
	for cur.Next(context.Background()) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Find Limit",
				Documents: map[string]interface{}{"filter": filter, "sort": sort, "page": page, "pageSize": pageSize},
				ErrMsg:    err.Error(), Result: fmt.Sprintf("length=%d", len(resultArray)), Ok: false})
			return resultArray, err
		}
		resultArray = append(resultArray, result)
	}
	//err = cur.All(context.Background(), &resultArray)
	if err := cur.Err(); err != nil {
		Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Find Limit",
			Documents: map[string]interface{}{"filter": filter, "sort": sort, "page": page, "pageSize": pageSize},
			ErrMsg:    err.Error(), Result: fmt.Sprintf("length=%d", len(resultArray)), Ok: false})
		return resultArray, err
	}
	Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Find Limit",
		Documents: map[string]interface{}{"filter": filter, "sort": sort, "page": page, "pageSize": pageSize},
		ErrMsg:    "", Result: fmt.Sprintf("length=%d", len(resultArray)), Ok: true})
	return resultArray, nil
}

// 返回查找条件的全部文档记录
// project 不能混合 True 和 False
func (m *Collection) FindAll(filter bson.D, sort, projection bson.M) ([]bson.M, error) {
	var resultArray []bson.M
	if projection == nil {
		projection = bson.M{}
	}
	findOptions := options.Find().SetProjection(projection)
	if sort != nil {
		findOptions = findOptions.SetSort(sort)
	}
	cur, err := m.client.Find(context.Background(), filter, findOptions)
	if err != nil {
		Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Find All",
			Documents: map[string]interface{}{"filter": filter, "sort": sort},
			ErrMsg:    err.Error(), Result: fmt.Sprintf("length=%d", len(resultArray)), Ok: false})
		return resultArray, err
	}
	defer func() {
		_ = cur.Close(context.Background())
	}()
	for cur.Next(context.Background()) {
		// fmt.Println(cur.Current)
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			return resultArray, err
		}
		resultArray = append(resultArray, result)
	}

	if err := cur.Err(); err != nil {
		Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Find All",
			Documents: map[string]interface{}{"filter": filter, "sort": sort},
			ErrMsg:    err.Error(), Result: fmt.Sprintf("length=%d", len(resultArray)), Ok: false})
		return resultArray, err
	}
	Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Find All",
		Documents: map[string]interface{}{"filter": filter, "sort": sort},
		ErrMsg:    "", Result: fmt.Sprintf("length=%d", len(resultArray)), Ok: true})
	return resultArray, nil
}

// 快速统计集合中所有文档的数量
func (m *Collection) CollectionCount() (int64, error) {
	size, _ := m.client.EstimatedDocumentCount(context.Background())
	Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Quick Count Collection",
		Documents: "", ErrMsg: "", Result: fmt.Sprintf("length=%d", size), Ok: true})
	return size, nil
}

// 根据条件统计集合中文档的数量
func (m *Collection) Count(filter bson.D) (int64, error) {
	size, _ := m.client.CountDocuments(context.Background(), filter)
	Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Count Collection",
		Documents: map[string]interface{}{"filter": filter}, ErrMsg: "",
		Result: fmt.Sprintf("length=%d", size), Ok: true})
	return size, nil
}

/*
查找并删除一个  sort 等于1表示 删除最旧的，sort 等于-1 表示删除最新的
一般根据 id 查找就会保证删除正确
*/
func (m *Collection) DeleteOne(filter bson.D, sort bson.M) (bson.M, error) {
	findOptions := options.FindOneAndDelete()
	if sort != nil {
		findOptions = findOptions.SetSort(sort)
	}
	singleResult := m.client.FindOneAndDelete(context.Background(), filter, findOptions)
	var result bson.M
	if err := singleResult.Decode(&result); err != nil {
		Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Delete One",
			Documents: map[string]interface{}{"filter": filter, "sort": sort},
			ErrMsg:    err.Error(), Result: "", Ok: false})
		return result, err
	}
	Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Delete One",
		Documents: map[string]interface{}{"filter": filter, "sort": sort},
		ErrMsg:    "", Result: "", Ok: true})
	return result, nil
}

/*
根据条件删除全部
*/
func (m *Collection) DeleteAll(filter bson.D) (int64, error) {
	count, err := m.client.DeleteMany(context.Background(), filter)
	if err != nil {
		Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Delete All",
			Documents: map[string]interface{}{"filter": filter},
			ErrMsg:    err.Error(), Result: "0", Ok: false})
		return 0, err
	}
	Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Delete All",
		Documents: map[string]interface{}{"filter": filter},
		ErrMsg:    "", Result: fmt.Sprintf("%d", count.DeletedCount), Ok: true})
	return count.DeletedCount, nil
}

/*
将id转换成时间
*/
func (m *Collection) ParseId(ObjectIdStr string) (time.Time, uint64) {
	timestamp, _ := strconv.ParseInt(ObjectIdStr[:8], 16, 64) // 4个字节8位
	dateTime := time.Unix(timestamp, 0)
	count, _ := strconv.ParseUint(ObjectIdStr[18:], 16, 64) // 随机码
	return dateTime, count
}

/*
更新 filter 返回的第一条记录
如果匹配到了（matchCount >= 1) 并且 objectId.IsZero()= false 则是新插入， objectId.IsZero()=true则是更新（更新没有获取到id）
ObjectID("000000000000000000000000")
document 修改为 interface 表示支持多个字段更新，使用bson.D ，即 []bson.E
*/
func (m *Collection) UpdateOne(filter bson.D, document interface{}, insert bool) (int64, primitive.ObjectID, error) {
	updateOption := options.Update().SetUpsert(insert)
	updateResult, err := m.client.UpdateOne(context.Background(), filter, document, updateOption)
	var objectId primitive.ObjectID
	if err != nil {
		Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Update One",
			Documents: map[string]interface{}{"filter": filter},
			ErrMsg:    err.Error(), Result: "0", Ok: false})
		return 0, objectId, err
	}
	if updateResult.UpsertedID != nil {
		objectId = updateResult.UpsertedID.(primitive.ObjectID)
	}
	// fmt.Println(objectId.IsZero())
	Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Update One",
		Documents: map[string]interface{}{"filter": filter},
		ErrMsg:    "", Result: fmt.Sprintf("%d", updateResult.MatchedCount), Ok: true})
	return updateResult.MatchedCount, objectId, nil
}

/*
更新 filter 返回的所有记录，返回的匹配是指本次查询匹配到的所有数量，也就是最后更新后等于新的值的数量
如果匹配到了（matchCount >= 1) 并且 objectId.IsZero()= false 则是新插入， objectId.IsZero()=true则是更新（更新没有获取到id）
ObjectID("000000000000000000000000")
docAction 修改为 interface 表示支持多个字段更新，使用bson.D ，即 []bson.E
*/
func (m *Collection) UpdateMany(filter bson.D, docAction interface{}, insert bool) (int64, primitive.ObjectID, error) {
	updateOption := options.Update().SetUpsert(insert)
	updateResult, err := m.client.UpdateMany(context.Background(), filter, docAction, updateOption)
	var objectId primitive.ObjectID
	if err != nil {
		Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Update Many",
			Documents: map[string]interface{}{"filter": filter},
			ErrMsg:    err.Error(), Result: "0", Ok: false})
		return 0, objectId, err
	}
	if updateResult.UpsertedID != nil {
		objectId = updateResult.UpsertedID.(primitive.ObjectID)
	}
	// fmt.Println(objectId.IsZero())
	Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Update Many",
		Documents: map[string]interface{}{"filter": filter},
		ErrMsg:    "", Result: fmt.Sprintf("%d", updateResult.MatchedCount), Ok: true})
	return updateResult.MatchedCount, objectId, nil
}

/*
替换 filter 返回的1条记录（最旧的）
如果匹配到了（matchCount >= 1) 并且 objectId.IsZero()= false 则是新插入， objectId.IsZero()=true则是更新（更新没有获取到id）
ObjectID("000000000000000000000000")
采用 FindOneAndReplace 在查找不到但正确插入新的数据会有"mongo: no documents in result" 的错误
*/
func (m *Collection) Replace(filter bson.D, document interface{}, insert bool) (int64, primitive.ObjectID, error) {
	//sortOpt := bson.D{{"_id", sort}}
	//option := options.FindOneAndReplace().SetSort(sortOpt).SetUpsert(insert)
	//replaceResult := m.client.FindOneAndReplace(context.Background(), filter, document, option)
	option := options.Replace().SetUpsert(insert)
	replaceResult, err := m.client.ReplaceOne(context.Background(), filter, document, option)
	var objectId primitive.ObjectID
	if err != nil {
		Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Replace",
			Documents: map[string]interface{}{"filter": filter},
			ErrMsg:    err.Error(), Result: "0", Ok: false})
		return 0, objectId, err
	}
	if replaceResult.UpsertedID != nil {
		objectId = replaceResult.UpsertedID.(primitive.ObjectID)
	}
	// fmt.Println(objectId.IsZero())
	Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Replace",
		Documents: map[string]interface{}{"filter": filter},
		ErrMsg:    "", Result: fmt.Sprintf("%d", replaceResult.MatchedCount), Ok: true})
	return replaceResult.MatchedCount, objectId, nil
	//var result bson.M
	//if err := replaceResult.Decode(&result); err != nil {
	//	if err.Error() == "mongo: no documents in result" {  //特殊处理
	//		return result, nil
	//	}
	//	return result, err
	//}
	//return result, nil
}

// 创建索引，重复创建不会报错
func (m *Collection) CreateIndex(index string, unique bool) (string, error) {
	indexModel := mongo.IndexModel{Keys: bson.M{index: 1}, Options: options.Index().SetUnique(unique)}
	name, err := m.client.Indexes().CreateOne(context.Background(), indexModel)
	Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Create Index",
		Documents: map[string]interface{}{"index": index, "unique": unique},
		ErrMsg:    "", Result: "", Ok: true})
	return name, err
}

// TODO 未测试
func (m *Collection) DeleteIndex(name string) error {
	_, err := m.client.Indexes().DropOne(context.Background(), name)
	Log(MongoLog{Database: m.database, Collection: m.collection, Action: "Delete Index",
		Documents: map[string]interface{}{"name": name},
		ErrMsg:    "", Result: "", Ok: true})
	return err
}

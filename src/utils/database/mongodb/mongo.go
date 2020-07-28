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

type Mongo struct {
	Client         *mongo.Client
	DatabaseName   string
	CollectionName string
	Collection     *mongo.Collection
}

// 连接池创建
func CreatePool(uri string, size uint64) (pool *mongo.Client, e error) {
	defer func() {
		if err := recover(); err != nil {
			e = errors.New(fmt.Sprintf("%v", err))
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 10s超时
	defer cancel()

	var err error
	pool, err = mongo.Connect(ctx, options.Client().ApplyURI(uri).SetMinPoolSize(size)) // 连接池
	if err != nil {
		return pool, err
	}
	err = pool.Ping(context.Background(), nil) // 检查连接
	if err != nil {
		return pool, err
	}
	return pool, nil
}

// 连接池销毁
func DestroyPool(client *mongo.Client) error {
	err := client.Disconnect(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// 初始化db连接
func NewConnection(pool *mongo.Client, database, collection string) *Mongo {
	mgo := new(Mongo)
	mgo.Client = pool
	mgo.DatabaseName = database
	mgo.CollectionName = collection
	mgo.Collection = pool.Database(mgo.DatabaseName).Collection(mgo.CollectionName)
	return mgo
}

// 关闭连接
func (m *Mongo) Close() {
	m.Client = nil
	m.DatabaseName = ""
	m.CollectionName = ""
	m.Collection = nil
}

//插入单个
func (m *Mongo) InsertOne(document interface{}) (primitive.ObjectID, error) {
	insertResult, err := m.Collection.InsertOne(context.Background(), document)
	var objectId primitive.ObjectID
	if err != nil {
		return objectId, err
	}
	objectId = insertResult.InsertedID.(primitive.ObjectID)
	return objectId, nil
}

//插入多个文档
func (m *Mongo) InsertMany(documents []interface{}) ([]primitive.ObjectID, error) {
	var insertDocs []interface{}
	for _, doc := range documents {
		insertDocs = append(insertDocs, doc)
	}
	insertResult, err := m.Collection.InsertMany(context.Background(), insertDocs)
	var objectIds []primitive.ObjectID
	if err != nil {
		return objectIds, err
	}
	for _, oid := range insertResult.InsertedIDs {
		objectIds = append(objectIds, oid.(primitive.ObjectID))
	}

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
func (m *Mongo) FindOne(filter bson.D, sort, projection bson.M) (bson.M, error) {
	findOptions := options.FindOne().SetProjection(projection)
	if sort != nil {
		findOptions = findOptions.SetSort(sort)
	}
	singleResult := m.Collection.FindOne(context.Background(), filter, findOptions)
	var result bson.M
	if err := singleResult.Decode(&result); err != nil {
		return result, err
	}
	return result, nil
}

/*
查询多个 sort 等于1表示 返回最旧的，sort 等于-1 表示返回最新的
每次只返回1页 page size 大小的数据
project 不能混合 True 和 False
*/
func (m *Mongo) FindLimit(filter bson.D, page, pageSize uint64, sort, projection bson.M) ([]bson.M, error) {
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
	cur, err := m.Collection.Find(context.Background(), filter, findOptions)
	if err != nil {
		return resultArray, err
	}
	defer func() {
		_ = cur.Close(context.Background())
	}()
	for cur.Next(context.Background()) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			return resultArray, err
		}
		resultArray = append(resultArray, result)
	}
	//err = cur.All(context.Background(), &resultArray)
	if err := cur.Err(); err != nil {
		return resultArray, err
	}
	return resultArray, nil
}

// 返回查找条件的全部文档记录
// project 不能混合 True 和 False
func (m *Mongo) FindAll(filter bson.D, sort, projection bson.M) ([]bson.M, error) {
	var resultArray []bson.M
	if projection == nil {
		projection = bson.M{}
	}
	findOptions := options.Find().SetProjection(projection)
	if sort != nil {
		findOptions = findOptions.SetSort(sort)
	}
	cur, err := m.Collection.Find(context.Background(), filter, findOptions)
	if err != nil {
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
		return resultArray, err
	}
	return resultArray, nil
}

// 快速统计集合中所有文档的数量
func (m *Mongo) CollectionCount() (int64, error) {
	size, _ := m.Collection.EstimatedDocumentCount(context.Background())
	return size, nil
}

// 根据条件统计集合中文档的数量
func (m *Mongo) Count(filter bson.D) (int64, error) {
	size, _ := m.Collection.CountDocuments(context.Background(), filter)
	return size, nil
}

/*
查找并删除一个  sort 等于1表示 删除最旧的，sort 等于-1 表示删除最新的
一般根据 id 查找就会保证删除正确
*/
func (m *Mongo) DeleteOne(filter bson.D, sort bson.M) (bson.M, error) {
	findOptions := options.FindOneAndDelete()
	if sort != nil {
		findOptions = findOptions.SetSort(sort)
	}
	singleResult := m.Collection.FindOneAndDelete(context.Background(), filter, findOptions)
	var result bson.M
	if err := singleResult.Decode(&result); err != nil {
		return result, err
	}
	return result, nil
}

/*
根据条件删除全部
*/
func (m *Mongo) DeleteAll(filter bson.D) (int64, error) {
	count, err := m.Collection.DeleteMany(context.Background(), filter)
	if err != nil {
		return 0, err
	}
	return count.DeletedCount, nil
}

/*
将id转换成时间
*/
func (m *Mongo) ParseId(ObjectIdStr string) (time.Time, uint64) {
	timestamp, _ := strconv.ParseInt(ObjectIdStr[:8], 16, 64) // 4个字节8位
	dateTime := time.Unix(timestamp, 0)
	count, _ := strconv.ParseUint(ObjectIdStr[18:], 16, 64) // 随机码
	return dateTime, count
}

/*
更新 filter 返回的第一条记录
如果匹配到了（matchCount >= 1) 并且 objectId.IsZero()= false 则是新插入， objectId.IsZero()=true则是更新（更新没有获取到id）
ObjectID("000000000000000000000000")
docAction 修改为 interface 表示支持多个字段更新，使用bson.D ，即 []bson.E
*/
func (m *Mongo) UpdateOne(filter bson.D, docAction interface{}, insert bool) (int64, primitive.ObjectID, error) {
	updateOption := options.Update().SetUpsert(insert)
	updateResult, err := m.Collection.UpdateOne(context.Background(), filter, docAction, updateOption)
	var objectId primitive.ObjectID
	if err != nil {
		return 0, objectId, err
	}
	if updateResult.UpsertedID != nil {
		objectId = updateResult.UpsertedID.(primitive.ObjectID)
	}
	// fmt.Println(objectId.IsZero())
	return updateResult.MatchedCount, objectId, nil
}

/*
更新 filter 返回的所有记录，返回的匹配是指本次查询匹配到的所有数量，也就是最后更新后等于新的值的数量
如果匹配到了（matchCount >= 1) 并且 objectId.IsZero()= false 则是新插入， objectId.IsZero()=true则是更新（更新没有获取到id）
ObjectID("000000000000000000000000")
docAction 修改为 interface 表示支持多个字段更新，使用bson.D ，即 []bson.E
*/
func (m *Mongo) UpdateMany(filter bson.D, docAction interface{}, insert bool) (int64, primitive.ObjectID, error) {
	updateOption := options.Update().SetUpsert(insert)
	updateResult, err := m.Collection.UpdateMany(context.Background(), filter, docAction, updateOption)
	var objectId primitive.ObjectID
	if err != nil {
		return 0, objectId, err
	}
	if updateResult.UpsertedID != nil {
		objectId = updateResult.UpsertedID.(primitive.ObjectID)
	}
	// fmt.Println(objectId.IsZero())
	return updateResult.MatchedCount, objectId, nil
}

/*
替换 filter 返回的1条记录（最旧的）
如果匹配到了（matchCount >= 1) 并且 objectId.IsZero()= false 则是新插入， objectId.IsZero()=true则是更新（更新没有获取到id）
ObjectID("000000000000000000000000")
采用 FindOneAndReplace 在查找不到但正确插入新的数据会有"mongo: no documents in result" 的错误
*/
func (m *Mongo) Replace(filter bson.D, document interface{}, insert bool) (int64, primitive.ObjectID, error) {
	//sortOpt := bson.D{{"_id", sort}}
	//option := options.FindOneAndReplace().SetSort(sortOpt).SetUpsert(insert)
	//replaceResult := m.Collection.FindOneAndReplace(context.Background(), filter, document, option)
	option := options.Replace().SetUpsert(insert)
	replaceResult, err := m.Collection.ReplaceOne(context.Background(), filter, document, option)
	var objectId primitive.ObjectID
	if err != nil {
		return 0, objectId, err
	}
	if replaceResult.UpsertedID != nil {
		objectId = replaceResult.UpsertedID.(primitive.ObjectID)
	}
	// fmt.Println(objectId.IsZero())
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
func (m *Mongo) CreateIndex(index string, unique bool) (string, error) {
	indexModel := mongo.IndexModel{Keys: bson.M{index: 1}, Options: options.Index().SetUnique(unique)}
	name, err := m.Collection.Indexes().CreateOne(context.Background(), indexModel)
	return name, err
}

// TODO 未测试
func (m *Mongo) DeleteIndex(name string) error {
	_, err := m.Collection.Indexes().DropOne(context.Background(), name)
	return err
}

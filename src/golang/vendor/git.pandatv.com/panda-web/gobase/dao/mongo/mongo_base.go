package mongokit

//
// Copyright (c) 2015-2025 - shenguanpu <shenguanpu@panda.tv>
//
// All rights reserved.
//
// 第二个返回值均为状态码
/*
*
*  MongoBaseDao  通用的crud list方法，不需要每个 dao实现
 */

import (
	// "golang/config"
	"git.pandatv.com/panda-web/gobase/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//需要运行即初始化
var GlobalMgoSession *mgo.Session
var globalMgoDbName = "panda_bikini"
var MgoBaseDao = &MongoBaseDao{}

func NewMgoSession(host string) *mgo.Session {
	logkit.Logger.Info("NewMongoSession ..." + host)

	//GlobalMgoSession, err := mgo.Dial(config.Get("env[MONGO_HOST]"))
	GlobalMgoSession, err := mgo.Dial(host)
	if err != nil {
		logkit.Logger.Error("NewMgoSession error: " + err.Error())
		panic(err)
	}
	return GlobalMgoSession
}

func CloseMgo() {
	GlobalMgoSession.Close()
}

/*interface 目前看无用途*/
type BaseDao interface {
	//just insert object
	Create(tablename string, doc interface{})
	//insert map and return id
	CreateAndGetIdByMap(tablename string, params map[string]interface{})
	//insert object and return id
	CreateAndGetIdByEntity(tablename string, entity *BaseEntity)
	//get ojbect by id
	Get(tablename string, id string, result interface{})
	//update object by id
	Update(tablename string, id string, doc interface{})
	//get lists by cond
	GetList(tablename string, cond interface{}, pageno int, pagenum int, result interface{})
}
type MongoBaseDao struct {
}

func (m *MongoBaseDao) Create(tablename string, doc interface{}) bool {
	session := GlobalMgoSession.Clone()
	defer session.Close()

	collection := session.DB(globalMgoDbName).C(tablename)

	err := collection.Insert(doc)

	if err != nil {
		logkit.Logger.Error("mongo_base method:Create " + err.Error())
		return false
	}
	return true
}
func (m *MongoBaseDao) CreateAndGetIdByMap(tablename string, params map[string]interface{}) (interface{}, bool) {
	session := GlobalMgoSession.Clone()
	defer session.Close()

	collection := session.DB(globalMgoDbName).C(tablename)

	id := bson.NewObjectId()
	params["_id"] = id
	err := collection.Insert(params)

	if err != nil {
		logkit.Logger.Error("mongo_base method:CreateAndGetIdByMap " + err.Error())
		return "0", false
	}
	return id, true
}

// entity.Gift => entity.BaseEntity 的向下转型golang 不支持，目前实现不了
//func (m *MongoBaseDao) CreateAndGetIdByEntity(tablename string, entity *entity.Gift) (interface{}, bool) {
func (m *MongoBaseDao) CreateAndGetIdByEntity(tablename string, entity *BaseEntity) (interface{}, bool) {
	session := GlobalMgoSession.Clone()
	defer session.Close()

	collection := session.DB(globalMgoDbName).C(tablename)

	id := bson.NewObjectId()
	entity.Id = id
	err := collection.Insert(entity)

	if err != nil {
		logkit.Logger.Error("mongo_base method:CreateAndGetIdByEntity " + err.Error())
		return "0", false
	}
	return id, true
}

func (m *MongoBaseDao) Get(tablename string, id string, result interface{}) interface{} {
	session := GlobalMgoSession.Clone()
	defer session.Close()
	if len(id) != 24 {
		return result
	}

	collection := session.DB(globalMgoDbName).C(tablename)
	// 两种方法都可行
	//err = c.Find(bson.D{{"_id", id}}).One(&result)
	//err = c.FindId(id).One(&result)

	// 晚上突然两种都不可用  what the fuck ????
	//2015.12.24 0:41

	err := collection.FindId(bson.ObjectIdHex(id)).One(result)
	if err != nil {
		logkit.Logger.Error("mongo_base method:Get " + err.Error())
	}

	return result
}

func (m *MongoBaseDao) GetByCond(tablename string, cond interface{}, result interface{}, sort string) interface{} {
	session := GlobalMgoSession.Clone()
	defer session.Close()

	collection := session.DB(globalMgoDbName).C(tablename)

	if sort == "" {
		err := collection.Find(cond).One(result)
		if err != nil {
			logkit.Logger.Error("mongo_base method:GetByCond " + err.Error())
		}
	} else {
		err := collection.Find(cond).Sort(sort).One(result)
		if err != nil {
			logkit.Logger.Error(err.Error())
			logkit.Logger.Error("mongo_base method:GetByCond sort " + sort + err.Error())
		}
	}

	return result
}

func (m *MongoBaseDao) Update(tablename string, id string, doc interface{}) bool {
	session := GlobalMgoSession.Clone()
	defer session.Close()

	collection := session.DB(globalMgoDbName).C(tablename)

	err := collection.UpdateId(bson.ObjectIdHex(id), doc)
	if err != nil {
		logkit.Logger.Error("mongo_base method:Update " + err.Error())
		return false
	}
	return true
}
func (m *MongoBaseDao) GetList(tablename string, cond interface{}, pageno int, pagenum int, result interface{}, sort string) int {
	if pageno <= 0 {
		pageno = 1
	}
	if pagenum <= 0 || pagenum > 100 {
		pagenum = 30
	}
	session := GlobalMgoSession.Clone()
	defer session.Close()

	collection := session.DB(globalMgoDbName).C(tablename)

	start := (pageno - 1) * pagenum
	err := collection.Find(cond).Sort(sort).Skip(start).Limit(pagenum).All(result)
	if err != nil {
		logkit.Logger.Error("mongo_base method:GetList " + err.Error())
	}
	count, err := collection.Find(cond).Count()
	if err != nil {
		logkit.Logger.Error("mongo_base method:GetList count " + err.Error())
	}

	return count
}

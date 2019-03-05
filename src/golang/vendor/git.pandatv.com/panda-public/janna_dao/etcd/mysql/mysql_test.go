package mysql

import (
	"testing"
	"time"
)

func TestNewCustom(t *testing.T) {
	var masterConfig = NewConfig()
	masterConfig.User = "p_dev"
	masterConfig.Passwd = "r2evKK0QJj1oHQe1"
	masterConfig.DBName = "gobase_mysql_demo"

	var slaveConfig = NewConfig()
	slaveConfig.User = "p_dev"
	slaveConfig.Passwd = "r2evKK0QJj1oHQe1"
	slaveConfig.DBName = "gobase_mysql_demo"

	MysqlClient, err := NewMysqlBaseDaoCustom("etcd.demo.pdtv.io:2379", "test", "default", masterConfig, slaveConfig)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 5; i++ {
		sql := "Insert INTO `test` (`ver`, `createtime`, `updatetime`, `name`) VALUES (?, ?, ?, ?)"
		now := time.Now().Format(time.RFC3339)
		var args = make([]interface{}, 0)
		args = append(args, 1)
		args = append(args, now)
		args = append(args, now)
		args = append(args, "haha")
		id, err := MysqlClient.Insert(sql, args...)
		if err != nil {
			t.Error(err)
		}
		t.Log(id)
	}

	var args = make([]interface{}, 0)
	args = append(args, "haha")
	res, err := MysqlClient.FetchRows("SELECT * FROM `test` WHERE `name` = ?", args...)
	if err != nil {
		t.Error(err)
	}
	for _, v := range res {
		t.Log(v)
	}
}

func TestNewCustomEnc(t *testing.T) {
	var masterConfig = NewConfig()
	masterConfig.DBName = "gobase_mysql_demo"

	var slaveConfig = NewConfig()
	slaveConfig.DBName = "gobase_mysql_demo"

	MysqlClient, err := NewMysqlBaseDaoCustomEnc("etcd.demo.pdtv.io:2379", "test", "userPwd", "key12345", masterConfig, slaveConfig)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 5; i++ {
		sql := "Insert INTO `test` (`ver`, `createtime`, `updatetime`, `name`) VALUES (?, ?, ?, ?)"
		now := time.Now().Format(time.RFC3339)
		var args = make([]interface{}, 0)
		args = append(args, 1)
		args = append(args, now)
		args = append(args, now)
		args = append(args, "haha")
		id, err := MysqlClient.Insert(sql, args...)
		if err != nil {
			t.Error(err)
		}
		t.Log(id)
	}

	var args = make([]interface{}, 0)
	args = append(args, "haha")
	res, err := MysqlClient.FetchRows("SELECT * FROM `test` WHERE `name` = ?", args...)
	if err != nil {
		t.Error(err)
	}
	for _, v := range res {
		t.Log(v)
	}
}

func TestNew(t *testing.T) {
	MysqlClient, err := NewMysqlBaseDao("etcd.demo.pdtv.io:2379", "test", "default", "p_dev", "r2evKK0QJj1oHQe1", "gobase_mysql_demo", "p_dev", "r2evKK0QJj1oHQe1", "gobase_mysql_demo")
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 5; i++ {
		sql := "Insert INTO `test` (`ver`, `createtime`, `updatetime`, `name`) VALUES (?, ?, ?, ?)"
		now := time.Now().Format(time.RFC3339)
		var args = make([]interface{}, 0)
		args = append(args, 1)
		args = append(args, now)
		args = append(args, now)
		args = append(args, "haha")
		id, err := MysqlClient.Insert(sql, args...)
		if err != nil {
			t.Error(err)
		}
		t.Log(id)
	}

	var args = make([]interface{}, 0)
	args = append(args, "haha")
	res, err := MysqlClient.FetchRows("SELECT * FROM `test` WHERE `name` = ?", args...)
	if err != nil {
		t.Error(err)
	}
	for _, v := range res {
		t.Log(v)
	}
}

func TestNewEnc(t *testing.T) {
	MysqlClient, err := NewMysqlBaseDaoEnc("etcd.demo.pdtv.io:2379", "test", "userPwd", "key12345", "gobase_mysql_demo", "gobase_mysql_demo")
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 5; i++ {
		sql := "Insert INTO `test` (`ver`, `createtime`, `updatetime`, `name`) VALUES (?, ?, ?, ?)"
		now := time.Now().Format(time.RFC3339)
		var args = make([]interface{}, 0)
		args = append(args, 1)
		args = append(args, now)
		args = append(args, now)
		args = append(args, "haha")
		id, err := MysqlClient.Insert(sql, args...)
		if err != nil {
			t.Error(err)
		}
		t.Log(id)
	}

	var args = make([]interface{}, 0)
	args = append(args, "haha")
	res, err := MysqlClient.FetchRows("SELECT * FROM `test` WHERE `name` = ?", args...)
	if err != nil {
		t.Error(err)
	}
	for _, v := range res {
		t.Log(v)
	}
}

package mysql

import (
	"testing"
	"time"

	"git.pandatv.com/panda-web/gobase/utils"
)

func initTestMysqlShardConfig() *ShardConfig {
	conf := NewShardConfig()
	conf.User = "p_dev"
	conf.Addr = "10.20.1.11:3306"
	conf.Passwd = "r2evKK0QJj1oHQe1"
	conf.DBName = "gobase_mysql_demo"
	return conf
}

func initTestMysqlShardEncConfig() *ShardConfig {
	conf := NewShardConfig()
	conf.Addr = "10.20.1.11:3306"
	conf.DBName = "gobase_mysql_demo"
	return conf
}

func TestMysqlShard(t *testing.T) {
	masterConfig := make([]*ShardConfig, 0)
	conf := initTestMysqlShardConfig()
	conf.ShardStart = 0
	conf.ShardEnd = 1
	conf.InstanceTag = "userPwd"
	masterConfig = append(masterConfig, conf)
	conf = initTestMysqlShardConfig()
	conf.ShardStart = 2
	conf.ShardEnd = 3
	conf.InstanceTag = "target1"
	masterConfig = append(masterConfig, conf)

	slaveConfig := make([]*ShardConfig, 0)
	conf = initTestMysqlShardConfig()
	conf.ShardStart = 0
	conf.ShardEnd = 1
	conf.InstanceTag = "userPwd"
	slaveConfig = append(slaveConfig, conf)
	conf = initTestMysqlShardConfig()
	conf.ShardStart = 2
	conf.ShardEnd = 3
	conf.InstanceTag = "target1"
	slaveConfig = append(slaveConfig, conf)

	MysqlClient, err := NewMysqlShardDao("etcd.demo.pdtv.io:2379", "test", masterConfig, slaveConfig, 4, utils.AlgCrc32)
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
		id, err := MysqlClient.Insert(uint64(time.Now().UnixNano()), sql, args...)
		if err != nil {
			t.Error(err)
		}
		t.Log(id)
	}

	for i := 0; i < 5; i++ {
		var args = make([]interface{}, 0)
		args = append(args, "haha")
		now := uint64(time.Now().UnixNano())
		res, err := MysqlClient.FetchRows(now, "SELECT * FROM `test` WHERE `name` = ?", args...)
		if err != nil {
			t.Error(err)
		}
		for _, v := range res {
			t.Log(v)
		}
	}

	MysqlClient.Close()
}

func TestMysqlShardEnc(t *testing.T) {
	masterConfig := make([]*ShardConfig, 0)
	conf := initTestMysqlShardEncConfig()
	conf.ShardStart = 0
	conf.ShardEnd = 1
	conf.InstanceTag = "userPwd"
	masterConfig = append(masterConfig, conf)
	conf = initTestMysqlShardEncConfig()
	conf.ShardStart = 2
	conf.ShardEnd = 3
	conf.InstanceTag = "userPwd1"
	masterConfig = append(masterConfig, conf)

	slaveConfig := make([]*ShardConfig, 0)
	conf = initTestMysqlShardEncConfig()
	conf.ShardStart = 0
	conf.ShardEnd = 1
	conf.InstanceTag = "userPwd"
	slaveConfig = append(slaveConfig, conf)
	conf = initTestMysqlShardEncConfig()
	conf.ShardStart = 2
	conf.ShardEnd = 3
	conf.InstanceTag = "userPwd1"
	slaveConfig = append(slaveConfig, conf)

	MysqlClient, err := NewMysqlShardDaoEnc("etcd.demo.pdtv.io:2379", "test", "key12345", masterConfig, slaveConfig, 4, utils.AlgCrc32)
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
		id, err := MysqlClient.Insert(uint64(time.Now().UnixNano()), sql, args...)
		if err != nil {
			t.Error(err)
		}
		t.Log(id)
	}

	for i := 0; i < 5; i++ {
		var args = make([]interface{}, 0)
		args = append(args, "haha")
		now := uint64(time.Now().UnixNano())
		res, _ := MysqlClient.FetchRows(now, "SELECT * FROM `test` WHERE `name` = ?", args...)
		if err != nil {
			t.Error(err)
		}
		for _, v := range res {
			t.Log(v)
		}
	}

	MysqlClient.Close()
}

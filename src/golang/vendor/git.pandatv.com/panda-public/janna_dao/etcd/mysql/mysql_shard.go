package mysql

import (
	"strings"

	"git.pandatv.com/panda-public/janna/etcd"
	"git.pandatv.com/panda-web/gobase/dao/mysql"
	"git.pandatv.com/panda-web/gobase/log"
)

type MysqlShardDao struct {
	*mysql.MysqlShardDao
}

type ShardConfig mysql.ShardConfig

func NewShardConfig() *ShardConfig {
	return (*ShardConfig)(mysql.NewShardConfig())
}

// 传入密钥
func NewMysqlShardDaoEnc(endPoints, callName, encryCode string, masterConfig []*ShardConfig, slaveConfig []*ShardConfig, shardCount uint64, shardAlg int) (*MysqlShardDao, error) {
	endPointsSlice := strings.Split(endPoints, ",")
	etcdClient, err := etcd.NewWithDes(etcd.Config{
		Endpoints:   endPointsSlice,
		DialTimeout: etcd.ConnectTimeout,
		Username:    defaultUser,
		Password:    defaultPwd,
	}, []byte(encryCode))

	if err != nil {
		logkit.Errorf("init etcd error:%s", err)
		return nil, err
	}

	masterConfigTemp := make([]*mysql.ShardConfig, len(masterConfig))
	slaveConfigTemp := make([]*mysql.ShardConfig, len(slaveConfig))

	for i, conf := range masterConfig {
		masterConfigTemp[i] = (*mysql.ShardConfig)(conf)
	}

	for i, conf := range slaveConfig {
		slaveConfigTemp[i] = (*mysql.ShardConfig)(conf)
	}

	cli, err := mysql.NewMysqlShardDaoJannaEnc(etcdClient, callName, masterConfigTemp, slaveConfigTemp, shardCount, shardAlg)
	if cli == nil {
		return nil, err
	}

	return &MysqlShardDao{cli}, nil
}

// 传入用户密码
func NewMysqlShardDao(endPoints, callName string, masterConfig []*ShardConfig, slaveConfig []*ShardConfig, shardCount uint64, shardAlg int) (*MysqlShardDao, error) {
	endPointsSlice := strings.Split(endPoints, ",")
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   endPointsSlice,
		DialTimeout: etcd.ConnectTimeout,
		Username:    defaultUser,
		Password:    defaultPwd,
	})

	if err != nil {
		logkit.Errorf("init etcd error:%s", err)
		return nil, err
	}

	masterConfigTemp := make([]*mysql.ShardConfig, len(masterConfig))
	slaveConfigTemp := make([]*mysql.ShardConfig, len(slaveConfig))

	for i, conf := range masterConfig {
		masterConfigTemp[i] = (*mysql.ShardConfig)(conf)
	}

	for i, conf := range slaveConfig {
		slaveConfigTemp[i] = (*mysql.ShardConfig)(conf)
	}

	cli, err := mysql.NewMysqlShardDaoJanna(etcdClient, callName, masterConfigTemp, slaveConfigTemp, shardCount, shardAlg)
	if cli == nil {
		return nil, err
	}

	return &MysqlShardDao{cli}, nil
}

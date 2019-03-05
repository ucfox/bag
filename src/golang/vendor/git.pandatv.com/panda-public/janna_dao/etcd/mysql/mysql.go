package mysql

import (
	"strings"

	"git.pandatv.com/panda-public/janna/etcd"
	"git.pandatv.com/panda-web/gobase/dao/mysql"
	"git.pandatv.com/panda-web/gobase/log"
)

const (
	defaultUser = "janna"
	defaultPwd  = "RDX19t0dNBoAsaTO"
)

type MysqlBaseDao struct {
	*mysql.MysqlBaseDao
}

type Config mysql.Config

func NewConfig() *Config {
	return (*Config)(mysql.NewConfig())
}

// 传入密钥
func NewMysqlBaseDaoEnc(endPoints, callName, targetTag, encryCode, masterDatabase, slaveDatabase string) (*MysqlBaseDao, error) {
	masterConfig := NewConfig()
	masterConfig.DBName = masterDatabase

	slaveConfig := NewConfig()
	slaveConfig.DBName = slaveDatabase

	return NewMysqlBaseDaoCustomEnc(endPoints, callName, targetTag, encryCode, masterConfig, slaveConfig)
}

func NewMysqlBaseDaoCustomEnc(endPoints, callName, targetTag, encryCode string, masterConfig *Config, slaveConfig *Config) (*MysqlBaseDao, error) {
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

	cli, err := mysql.NewMysqlBaseDaoJannaCustomEnc(etcdClient, callName, targetTag, (*mysql.Config)(masterConfig), (*mysql.Config)(slaveConfig))
	if cli == nil {
		return nil, err
	}

	return &MysqlBaseDao{cli}, nil
}

// 传入用户密码
func NewMysqlBaseDao(endPoints, callName, targetTag, masterUser, masterPwd, masterDatabase, slaveUser, slavePwd, slaveDatabase string) (*MysqlBaseDao, error) {
	masterConfig := NewConfig()
	masterConfig.User = masterUser
	masterConfig.Passwd = masterPwd
	masterConfig.DBName = masterDatabase

	slaveConfig := NewConfig()
	slaveConfig.User = slaveUser
	slaveConfig.Passwd = slavePwd
	slaveConfig.DBName = slaveDatabase

	return NewMysqlBaseDaoCustom(endPoints, callName, targetTag, masterConfig, slaveConfig)
}

func NewMysqlBaseDaoCustom(endPoints, callName, targetTag string, masterConfig *Config, slaveConfig *Config) (*MysqlBaseDao, error) {
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

	cli, err := mysql.NewMysqlBaseDaoJannaCustom(etcdClient, callName, targetTag, (*mysql.Config)(masterConfig), (*mysql.Config)(slaveConfig))
	if cli == nil {
		return nil, err
	}

	return &MysqlBaseDao{cli}, nil
}

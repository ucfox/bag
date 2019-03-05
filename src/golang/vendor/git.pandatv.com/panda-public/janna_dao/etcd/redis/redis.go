package redis

import (
	"strings"
	"time"

	"git.pandatv.com/panda-public/janna/etcd"
	"git.pandatv.com/panda-web/gobase/dao/redis"
	"git.pandatv.com/panda-web/gobase/log"
)

const (
	defaultUser = "janna"
	defaultPwd  = "RDX19t0dNBoAsaTO"
)

type RedisBaseDao struct {
	*redis.RedisBaseDao
}

type Config redis.Config

func NewConfig() *Config {
	return (*Config)(redis.NewConfig())
}

// 传入密钥
func NewRedisClientEnc(endPoints, callName, targetTag, encryCode string) *RedisBaseDao {
	masterConfig := NewConfig()
	slaveConfig := NewConfig()

	return NewRedisClientCustomEnc(endPoints, callName, targetTag, encryCode, masterConfig, slaveConfig)
}

func NewRedisClientCustomEnc(endPoints, callName, targetTag, encryCode string, masterConfig *Config, slaveConfig *Config) *RedisBaseDao {
	endPointsSlice := strings.Split(endPoints, ",")
	etcdClient, err := etcd.NewWithDes(etcd.Config{
		Endpoints:   endPointsSlice,
		DialTimeout: etcd.ConnectTimeout,
		Username:    defaultUser,
		Password:    defaultPwd,
	}, []byte(encryCode))

	if err != nil {
		logkit.Errorf("init etcd error:%s", err)
		return nil
	}

	cli := redis.NewRedisClientJannaCustomEnc(etcdClient, callName, targetTag, (*redis.Config)(masterConfig), (*redis.Config)(slaveConfig))
	if cli == nil {
		return nil
	}

	return &RedisBaseDao{cli}
}

// 传入密码
func NewRedisClient(endPoints, callName, targetTag, masterPwd, slavePwd string) *RedisBaseDao {
	return NewRedisClientCustom(endPoints, callName, targetTag, masterPwd, slavePwd, redis.Default_connect_timeout, redis.Default_read_timeout, redis.Default_write_timeout, redis.Default_idle_timeout, redis.Default_max_active, redis.Default_max_idle, redis.Default_wait)
}

func NewRedisClientCustom(endPoints, callName, targetTag, masterPwd, slavePwd string, connectTimeout, readTimeout, writeTimeout, idleTimeout time.Duration, maxActive, maxIdle int, waitConn bool) *RedisBaseDao {
	endPointsSlice := strings.Split(endPoints, ",")
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   endPointsSlice,
		DialTimeout: etcd.ConnectTimeout,
		Username:    defaultUser,
		Password:    defaultPwd,
	})

	if err != nil {
		logkit.Errorf("init etcd error:%s", err)
		return nil
	}

	cli := redis.NewRedisClientJannaCustom(etcdClient, callName, targetTag, masterPwd, slavePwd, connectTimeout, readTimeout, writeTimeout, idleTimeout, maxActive, maxIdle, waitConn)
	if cli == nil {
		return nil
	}

	return &RedisBaseDao{cli}
}

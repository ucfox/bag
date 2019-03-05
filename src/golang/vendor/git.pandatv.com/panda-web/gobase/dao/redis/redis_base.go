package redis

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	janna "git.pandatv.com/panda-public/janna/client"
	"git.pandatv.com/panda-web/gobase/dao/watcher"
	"git.pandatv.com/panda-web/gobase/log"
	"github.com/garyburd/redigo/redis"
)

const (
	Default_connect_timeout = time.Second
	Default_read_timeout    = 500 * time.Millisecond
	Default_write_timeout   = 500 * time.Millisecond
	Default_idle_timeout    = 60 * time.Second
	Default_max_active      = 50
	Default_max_idle        = 25
	Default_wait            = true

	masterTag        = "master"
	slaveTag         = "slave"
	serviceNameRedis = "redis"
)

var (
	errNoUseablePool = errors.New("redis_base: No useable read pool!")
	ErrNil           = redis.ErrNil
)

type connPool struct {
	pool   *redis.Pool
	addr   string
	id     string
	active bool
}

type Config struct {
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxActive      int
	MaxIdle        int
	Wait           bool
}

//client instance struct
type RedisBaseDao struct {
	//主从连接池
	masterPool  []*connPool
	slavePool   []*connPool
	masterIndex *indexInt32
	slaveIndex  *indexInt32
	//主配置
	masterConfig *Config
	//从配置
	slaveConfig *Config
	//用户传入的密码
	masterPwd string
	slavePwd  string
	//true，使用janna中的密码，false，使用用户传入的密码
	encFlag   bool
	masterMux *sync.RWMutex
	slaveMux  *sync.RWMutex
	//连接池是否可用
	useable bool
	//true， 读从库，false，读主库
	readStale bool
	//是否过滤失效连接池
	filterFailConn bool
	//回收go routine
	closeCh chan struct{}
	//janna watch
	watcher watcher.IWatcher
}

func NewConfig() *Config {
	c := new(Config)
	c.ConnectTimeout = Default_connect_timeout
	c.ReadTimeout = Default_read_timeout
	c.WriteTimeout = Default_write_timeout
	c.IdleTimeout = Default_idle_timeout
	c.MaxActive = Default_max_active
	c.MaxIdle = Default_max_idle
	c.Wait = Default_wait

	return c
}

func newRedisPoolCustom(server string, password string, conf *Config) *redis.Pool {
	logkit.Debugf("init redis server: %s", server)
	return &redis.Pool{
		MaxIdle:     conf.MaxIdle,
		MaxActive:   conf.MaxActive,
		IdleTimeout: conf.IdleTimeout,
		Wait:        conf.Wait,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialTimeout("tcp", server, conf.ConnectTimeout, conf.ReadTimeout, conf.WriteTimeout)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			//if time.Since(t) < conf.IdleTimeout/5 {
			//	return nil
			//}
			_, err := c.Do("PING")
			return err
		},
	}
}

func initRedisBaseDao() *RedisBaseDao {
	RedisClient := new(RedisBaseDao)
	RedisClient.masterPool = make([]*connPool, 0)
	RedisClient.slavePool = make([]*connPool, 0)
	RedisClient.masterIndex = new(indexInt32)
	RedisClient.slaveIndex = new(indexInt32)
	RedisClient.masterMux = &sync.RWMutex{}
	RedisClient.slaveMux = &sync.RWMutex{}
	RedisClient.useable = true
	RedisClient.readStale = true
	RedisClient.filterFailConn = true
	RedisClient.closeCh = make(chan struct{})
	return RedisClient
}

func initJannaInstance(servConf janna.ServiceClient, callName, targetTag string, masterConf, slaveConf *Config, masterPwd, slavePwd string, encFlag bool) *RedisBaseDao {
	if masterConf == nil || slaveConf == nil {
		logkit.Errorf("redis conf is nil")
		return nil
	}
	w, err := watcher.NewWatcher(servConf)
	if err != nil {
		logkit.Errorf("init watcher error:%s", err)
		return nil
	}
	redisInfos, err := w.GetAllInstance(callName, serviceNameRedis, targetTag)
	if err != nil {
		logkit.Errorf("get all redis info from janna error:%s", err)
		return nil
	}

	RedisClient := initRedisBaseDao()
	RedisClient.masterConfig = masterConf
	RedisClient.slaveConfig = slaveConf
	RedisClient.masterPwd = masterPwd
	RedisClient.slavePwd = slavePwd
	RedisClient.encFlag = encFlag
	RedisClient.watcher = w

	for _, redisInfo := range redisInfos {
		addressPort := fmt.Sprintf("%s:%d", redisInfo.Address, redisInfo.Port)
		var pwd string
		var rdsConf *Config
		var serverPool *[]*connPool
		if redisInfo.Master {
			serverPool = &RedisClient.masterPool
			rdsConf = RedisClient.masterConfig
			pwd = RedisClient.masterPwd
		} else {
			serverPool = &RedisClient.slavePool
			rdsConf = RedisClient.slaveConfig
			pwd = RedisClient.slavePwd
		}
		if RedisClient.encFlag {
			pwd = redisInfo.Password
		}
		tempPool := newRedisPoolCustom(addressPort, pwd, rdsConf)
		logkit.Infof("add master:%t id:%s addr:%s", redisInfo.Master, redisInfo.Id, addressPort)
		*serverPool = append(*serverPool, &connPool{
			pool:   tempPool,
			addr:   addressPort,
			id:     redisInfo.Id,
			active: true,
		})
	}

	go RedisClient.checkBadPool(true)
	go RedisClient.checkBadPool(false)
	go RedisClient.jannaWatch(callName, targetTag)

	return RedisClient
}

//only capture the instance has the target tag
func NewRedisClientJannaEnc(servConf janna.ServiceClient, callName, targetTag string) *RedisBaseDao {
	masterConfig := NewConfig()
	slaveConfig := NewConfig()
	return NewRedisClientJannaCustomEnc(servConf, callName, targetTag, masterConfig, slaveConfig)
}

func NewRedisClientJannaCustomEnc(servConf janna.ServiceClient, callName, targetTag string, masterConf, slaveConf *Config) *RedisBaseDao {
	return initJannaInstance(servConf, callName, targetTag, masterConf, slaveConf, "", "", true)
}

func NewRedisClient(master string, masterPwd string, slave string, slavePwd string) *RedisBaseDao {
	return NewRedisClientCustom(master, masterPwd, slave, slavePwd, Default_connect_timeout, Default_read_timeout, Default_write_timeout, Default_idle_timeout, Default_max_active, Default_max_idle, Default_wait)
}

func NewRedisClientCustom(master string, masterPwd string, slave string, slavePwd string, connectTimeout, readTimeout, writeTimeout, idleTimeout time.Duration, maxActive, maxIdle int, waitConn bool) *RedisBaseDao {
	RedisClient := initRedisBaseDao()
	RedisClient.masterPwd = masterPwd
	RedisClient.slavePwd = slavePwd
	masterConf := new(Config)
	masterConf.ConnectTimeout = connectTimeout
	masterConf.ReadTimeout = readTimeout
	masterConf.WriteTimeout = writeTimeout
	masterConf.IdleTimeout = idleTimeout
	masterConf.MaxActive = maxActive
	masterConf.MaxIdle = maxIdle
	masterConf.Wait = waitConn
	RedisClient.masterConfig = masterConf
	RedisClient.slaveConfig = masterConf

	tempPool := newRedisPoolCustom(master, RedisClient.masterPwd, RedisClient.masterConfig)
	RedisClient.masterPool = append(RedisClient.masterPool, &connPool{
		pool:   tempPool,
		addr:   master,
		id:     "",
		active: true,
	})

	sl_slice := strings.Split(slave, ",")
	for _, sl := range sl_slice {
		tempPool := newRedisPoolCustom(sl, RedisClient.slavePwd, RedisClient.slaveConfig)
		RedisClient.slavePool = append(RedisClient.slavePool, &connPool{
			pool:   tempPool,
			addr:   sl,
			id:     "",
			active: true,
		})
	}

	go RedisClient.checkBadPool(true)
	go RedisClient.checkBadPool(false)

	return RedisClient
}

//only capture the instance has the target tag
func NewRedisClientJanna(servConf janna.ServiceClient, callName, targetTag, masterPwd, slavePwd string) *RedisBaseDao {
	return NewRedisClientJannaCustom(servConf, callName, targetTag, masterPwd, slavePwd, Default_connect_timeout, Default_read_timeout, Default_write_timeout, Default_idle_timeout, Default_max_active, Default_max_idle, Default_wait)
}

func NewRedisClientJannaCustom(servConf janna.ServiceClient, callName, targetTag, masterPwd, slavePwd string, connectTimeout, readTimeout, writeTimeout, idleTimeout time.Duration, maxActive, maxIdle int, waitConn bool) *RedisBaseDao {
	masterConf := new(Config)
	masterConf.ConnectTimeout = connectTimeout
	masterConf.ReadTimeout = readTimeout
	masterConf.WriteTimeout = writeTimeout
	masterConf.IdleTimeout = idleTimeout
	masterConf.MaxActive = maxActive
	masterConf.MaxIdle = maxIdle
	masterConf.Wait = waitConn

	return initJannaInstance(servConf, callName, targetTag, masterConf, masterConf, masterPwd, slavePwd, false)
}

func (dao *RedisBaseDao) jannaWatch(callName, targetTag string) {
	if dao.watcher == nil {
		logkit.Errorf("redis: nil watcher")
		return
	}

	ch, err := dao.watcher.WatchInstance(callName, serviceNameRedis, targetTag)
	if err != nil {
		logkit.Errorf("get janna watch error:%s", err)
		return
	}

	for redisEvent := range ch {
		logkit.Infof("watch event %+v", redisEvent)
		var pwd string
		var serverPool *[]*connPool
		var mux *sync.RWMutex
		var rdsConf *Config
		if redisEvent.Master {
			pwd = dao.masterPwd
			serverPool = &dao.masterPool
			mux = dao.masterMux
			rdsConf = dao.masterConfig
		} else {
			pwd = dao.slavePwd
			serverPool = &dao.slavePool
			mux = dao.slaveMux
			rdsConf = dao.slaveConfig
		}
		if dao.encFlag {
			pwd = redisEvent.Password
		}

		mux.Lock()

		for i, pool := range *serverPool {
			if pool.id == redisEvent.Id {
				logkit.Infof("rm master:%t id:%s addr:%s", redisEvent.Master, pool.id, pool.addr)
				pool.pool.Close()
				*serverPool = append((*serverPool)[0:i], (*serverPool)[i+1:]...)
				break
			}
		}

		if redisEvent.Opt == janna.OptPut {
			addressPort := fmt.Sprintf("%s:%d", redisEvent.Address, redisEvent.Port)
			tempPool := newRedisPoolCustom(addressPort, pwd, rdsConf)
			logkit.Infof("add master:%t id:%s addr:%s", redisEvent.Master, redisEvent.Id, addressPort)
			*serverPool = append(*serverPool, &connPool{
				pool:   tempPool,
				addr:   addressPort,
				id:     redisEvent.Id,
				active: true,
			})
		}

		mux.Unlock()
	}
}

func (dao *RedisBaseDao) CloseRedis() {
	close(dao.closeCh)

	dao.masterMux.Lock()
	dao.slaveMux.Lock()
	dao.useable = false
	dao.slaveMux.Unlock()
	dao.masterMux.Unlock()

	if dao.watcher != nil {
		err := dao.watcher.Close()
		if err != nil {
			logkit.Errorf("close redis janna watch error:%s", err)
		}
	}

	for _, serverPool := range [][]*connPool{dao.slavePool, dao.masterPool} {
		for _, pool := range serverPool {
			pool.pool.Close()
		}
	}
}

func (dao *RedisBaseDao) checkBadPool(write bool) {
	var serverPool *[]*connPool
	var mux *sync.RWMutex
	var index int
	if write {
		serverPool = &dao.masterPool
		mux = dao.masterMux
	} else {
		serverPool = &dao.slavePool
		mux = dao.slaveMux
	}

	t := time.Tick(time.Second * 2)
	for {
		select {
		case <-t:
			mux.Lock()

			for i := index; i < index+len(*serverPool); i++ {
				tempPos := i % len(*serverPool)
				if !(*serverPool)[tempPos].active {
					logkit.Debugf("check master:%t pool:%s status", write, (*serverPool)[tempPos].addr)
					conn := (*serverPool)[tempPos].pool.Get()
					if conn.Err() == nil {
						(*serverPool)[tempPos].active = true
					}
					index = tempPos + 1
					break
				}
			}

			mux.Unlock()
		case <-dao.closeCh:
			return
		}
	}
}

func (dao *RedisBaseDao) getConn(write bool) (redis.Conn, error) {
	var conn redis.Conn
	var serverPool []*connPool
	var index *indexInt32
	if write {
		serverPool = dao.masterPool
		index = dao.masterIndex
	} else {
		serverPool = dao.slavePool
		index = dao.slaveIndex
	}

	lenPool := len(serverPool)
	if lenPool == 0 {
		return nil, errNoUseablePool
	}

	var pool *connPool
	pos := index.IncrAndMod(lenPool)
	for i := pos; i < pos+lenPool; i++ {
		tempPos := i % lenPool
		if serverPool[tempPos].active {
			pool = serverPool[tempPos]
		}
	}

	if pool == nil {
		return nil, errNoUseablePool
	}

	conn = pool.pool.Get()
	if !dao.filterFailConn || conn.Err() == nil {
		logkit.Debugf("choose master:%t addr:%s\n", write, pool.addr)
		return conn, nil
	}

	logkit.Errorf("[dao|redis]redis master:%t addr:%s has error:%s", write, pool.addr, conn.Err())
	pool.active = false

	return conn, nil

}

func (dao *RedisBaseDao) doWrite(cmd string, args ...interface{}) (reply interface{}, err error) {
	conn, err := dao.getWrite()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return conn.Do(cmd, args...)
}

func (dao *RedisBaseDao) getWrite() (redis.Conn, error) {
	dao.masterMux.RLock()
	defer dao.masterMux.RUnlock()
	if dao.useable {
		return dao.getConn(true)
	}

	return nil, errNoUseablePool
}

func (dao *RedisBaseDao) doRead(cmd string, args ...interface{}) (reply interface{}, err error) {
	conn, err := dao.getRead()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return conn.Do(cmd, args...)
}

func (dao *RedisBaseDao) getRead() (redis.Conn, error) {
	var conn redis.Conn
	var err error

	dao.slaveMux.RLock()
	if dao.useable && dao.readStale {
		conn, err = dao.getConn(false)
		dao.slaveMux.RUnlock()
		return conn, err
	}
	dao.slaveMux.RUnlock()

	if !dao.readStale {
		return dao.getWrite()
	}

	return nil, errNoUseablePool
}

func (dao *RedisBaseDao) SetReadStale(stale bool) {
	dao.readStale = stale
}

func (dao *RedisBaseDao) SetFilterFailConn(filter bool) {
	dao.filterFailConn = filter
}

package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	janna "git.pandatv.com/panda-public/janna/client"
	"git.pandatv.com/panda-web/gobase/dao/watcher"
	"git.pandatv.com/panda-web/gobase/log"
	"git.pandatv.com/panda-web/gobase/utils"
)

var (
	ErrShardType = errors.New("unkown shard type")
)

type MysqlShardDao struct {
	shardCount uint64
	shardAlg   int

	dbWrite     [][]connPool
	dbRead      [][]connPool
	writeConfig []*ShardConfig
	readConfig  []*ShardConfig
	writeMux    *sync.RWMutex
	readMux     *sync.RWMutex
	writePos    []*countInt32
	readPos     []*countInt32
	encFlag     bool

	watcher watcher.IWatcher

	releaseCh      chan connPool
	closed         bool
	closeCh        chan struct{}
	filterFailConn bool
}

type ShardConfig struct {
	*Config
	ShardStart  uint64
	ShardEnd    uint64
	InstanceTag string
}

func NewShardConfig() *ShardConfig {
	c := new(ShardConfig)
	c.Config = NewConfig()
	c.MaxOpenConns = 30

	return c
}

func initialMysqlShard() *MysqlShardDao {
	mysqlDao := new(MysqlShardDao)
	mysqlDao.writeMux = new(sync.RWMutex)
	mysqlDao.readMux = new(sync.RWMutex)

	mysqlDao.closed = false
	mysqlDao.closeCh = make(chan struct{})
	mysqlDao.releaseCh = make(chan connPool, 10)
	mysqlDao.filterFailConn = true

	return mysqlDao
}

func NewMysqlShardDao(masterConfig []*ShardConfig, slaveConfig []*ShardConfig, shardCount uint64, shardAlg int) (*MysqlShardDao, error) {
	var err error
	MysqlClient := initialMysqlShard()
	err = MysqlClient.initShard(shardCount, shardAlg)
	if err != nil {
		return nil, err
	}

	MysqlClient.writeConfig = masterConfig
	MysqlClient.readConfig = slaveConfig

	for _, conf := range masterConfig {
		for i := conf.ShardStart; i <= conf.ShardEnd; i++ {
			if i >= shardCount {
				break
			}
			dbName := conf.DBName
			conf.DBName = utils.GetDbName(conf.DBName, int(i))
			dbWrite, err := conn(conf.Config)
			conf.DBName = dbName
			if err != nil {
				return nil, err
			} else {
				if MysqlClient.dbWrite[i] == nil {
					MysqlClient.dbWrite[i] = make([]connPool, 0)
				}
				MysqlClient.dbWrite[i] = append(MysqlClient.dbWrite[i], connPool{
					pool: dbWrite,
					addr: conf.Addr,
					id:   "",
				})
			}
		}
	}

	for num, pools := range MysqlClient.dbWrite {
		for _, pool := range pools {
			logkit.Infof("init mysql master shard %d addr %s", num, pool.addr)
		}
	}

	for _, conf := range slaveConfig {
		for i := conf.ShardStart; i <= conf.ShardEnd; i++ {
			if i >= shardCount {
				break
			}
			dbName := conf.DBName
			conf.DBName = utils.GetDbName(conf.DBName, int(i))
			dbRead, err := conn(conf.Config)
			conf.DBName = dbName
			if err != nil {
				return nil, err
			} else {
				if MysqlClient.dbRead[i] == nil {
					MysqlClient.dbRead[i] = make([]connPool, 0)
				}
				MysqlClient.dbRead[i] = append(MysqlClient.dbRead[i], connPool{
					pool: dbRead,
					addr: conf.Addr,
					id:   "",
				})
			}
		}
	}

	for num, pools := range MysqlClient.dbRead {
		for _, pool := range pools {
			logkit.Infof("init mysql slave shard %d addr %s", num, pool.addr)
		}
	}

	return MysqlClient, nil
}

func (db *MysqlShardDao) initShard(shardCount uint64, shardAlg int) error {
	db.shardCount = shardCount
	db.shardAlg = shardAlg

	db.dbWrite = make([][]connPool, shardCount)
	db.dbRead = make([][]connPool, shardCount)
	db.writePos = make([]*countInt32, shardCount)
	db.readPos = make([]*countInt32, shardCount)
	for i, _ := range db.writePos {
		db.writePos[i] = new(countInt32)
	}
	for i, _ := range db.readPos {
		db.readPos[i] = new(countInt32)
	}

	return nil
}

func initJannaShardInstance(servConf janna.ServiceClient, callName string, masterConfig []*ShardConfig, slaveConfig []*ShardConfig, shardCount uint64, shardAlg int, encFlag bool) (*MysqlShardDao, error) {
	var err error
	MysqlClient := initialMysqlShard()
	err = MysqlClient.initShard(shardCount, shardAlg)
	if err != nil {
		return nil, err
	}

	MysqlClient.writeConfig = masterConfig
	MysqlClient.readConfig = slaveConfig
	MysqlClient.encFlag = encFlag

	watch, err := watcher.NewWatcher(servConf)
	if err != nil {
		return nil, err
	}
	MysqlClient.watcher = watch

	dbList, err := MysqlClient.watcher.GetAllInstance(callName, serviceNameMysql, "")
	if err != nil {
		return nil, err
	}

	for _, dbInfo := range dbList {
		addressPort := fmt.Sprintf("%s:%d", dbInfo.Address, dbInfo.Port)
		var conf *ShardConfig
		var dbPool [][]connPool
		if dbInfo.Master {
			dbPool = MysqlClient.dbWrite
			for _, writeConfig := range MysqlClient.writeConfig {
				for _, tag := range dbInfo.Tag {
					if writeConfig.InstanceTag == tag {
						conf = writeConfig
						break
					}
				}
				if conf != nil {
					break
				}
			}
		} else {
			dbPool = MysqlClient.dbRead
			for _, readConfig := range MysqlClient.readConfig {
				for _, tag := range dbInfo.Tag {
					if readConfig.InstanceTag == tag {
						conf = readConfig
						break
					}
				}
				if conf != nil {
					break
				}
			}
		}
		if conf == nil {
			continue
		}

		conf.Addr = addressPort
		if MysqlClient.encFlag {
			conf.User = dbInfo.User
			conf.Passwd = dbInfo.Password
		}
		for i := conf.ShardStart; i <= conf.ShardEnd; i++ {
			if i >= shardCount {
				break
			}
			dbName := conf.DBName
			conf.DBName = utils.GetDbName(conf.DBName, int(i))
			pool, err := conn(conf.Config)
			conf.DBName = dbName
			if err != nil {
				return nil, err
			} else {
				if dbPool[i] == nil {
					dbPool[i] = make([]connPool, 0)
				}
				dbPool[i] = append(dbPool[i], connPool{
					pool: pool,
					addr: conf.Addr,
					id:   dbInfo.Id,
				})
			}
		}

	}

	go MysqlClient.jannaWatch(callName, "")

	return MysqlClient, nil

}

func NewMysqlShardDaoJanna(servConf janna.ServiceClient, callName string, masterConfig []*ShardConfig, slaveConfig []*ShardConfig, shardCount uint64, shardAlg int) (*MysqlShardDao, error) {
	return initJannaShardInstance(servConf, callName, masterConfig, slaveConfig, shardCount, shardAlg, false)
}

func NewMysqlShardDaoJannaEnc(servConf janna.ServiceClient, callName string, masterConfig []*ShardConfig, slaveConfig []*ShardConfig, shardCount uint64, shardAlg int) (*MysqlShardDao, error) {
	return initJannaShardInstance(servConf, callName, masterConfig, slaveConfig, shardCount, shardAlg, true)
}

func (db *MysqlShardDao) jannaWatch(callName, targetTag string) {
	if db.watcher == nil {
		logkit.Errorf("mysql: nil watch")
		return
	}

	ch, err := db.watcher.WatchInstance(callName, serviceNameMysql, targetTag)
	if err != nil {
		logkit.Errorf("get janna watch error:%s", err)
		return
	}

	for mysqlEvent := range ch {
		logkit.Infof("watch event %+v", mysqlEvent)
		var conf *ShardConfig
		var dbPool [][]connPool
		var mux *sync.RWMutex
		if mysqlEvent.Master {
			dbPool = db.dbWrite
			mux = db.writeMux
			for _, writeConfig := range db.writeConfig {
				for _, tag := range mysqlEvent.Tag {
					if writeConfig.InstanceTag == tag {
						conf = writeConfig
						break
					}
				}
				if conf != nil {
					break
				}
			}
		} else {
			dbPool = db.dbRead
			mux = db.readMux
			for _, readConfig := range db.readConfig {
				for _, tag := range mysqlEvent.Tag {
					if readConfig.InstanceTag == tag {
						conf = readConfig
						break
					}
				}
				if conf != nil {
					break
				}
			}
		}

		if conf == nil {
			continue
		}

		if mysqlEvent.Opt == janna.OptPut {
			for i := conf.ShardStart; i <= conf.ShardEnd; i++ {
				for j, serverPool := range dbPool[i] {
					if serverPool.id == mysqlEvent.Id {
						mux.Lock()
						logkit.Infof("rm master:%t id:%s addr:%s", mysqlEvent.Master, serverPool.id, serverPool.addr)
						db.releaseCh <- serverPool
						dbPool[i] = append(dbPool[i][0:j], dbPool[i][j+1:]...)
						mux.Unlock()
						break
					}
				}
			}
			mux.Lock()
			addressPort := fmt.Sprintf("%s:%d", mysqlEvent.Address, mysqlEvent.Port)
			logkit.Infof("add master:%t id:%s addr:%s", mysqlEvent.Master, mysqlEvent.Id, addressPort)
			conf.Addr = addressPort
			if db.encFlag {
				conf.User = mysqlEvent.User
				conf.Passwd = mysqlEvent.Password
			}
			for i := conf.ShardStart; i <= conf.ShardEnd; i++ {
				if i >= db.shardCount {
					break
				}
				dbName := conf.DBName
				conf.DBName = utils.GetDbName(conf.DBName, int(i))
				pool, _ := conn(conf.Config)
				conf.DBName = dbName
				if dbPool[i] == nil {
					dbPool[i] = make([]connPool, 0)
				}
				dbPool[i] = append(dbPool[i], connPool{
					pool: pool,
					addr: conf.Addr,
					id:   mysqlEvent.Id,
				})
			}
			mux.Unlock()
		} else if mysqlEvent.Opt == janna.OptDelete {
			for i := conf.ShardStart; i <= conf.ShardEnd; i++ {
				for j, serverPool := range dbPool[i] {
					if serverPool.id == mysqlEvent.Id {
						mux.Lock()
						logkit.Infof("rm master:%t id:%s addr:%s", mysqlEvent.Master, serverPool.id, serverPool.addr)
						db.releaseCh <- serverPool
						dbPool[i] = append(dbPool[i][0:j], dbPool[i][j+1:]...)
						mux.Unlock()
						break
					}
				}
			}
		}
	}
}

func (db *MysqlShardDao) releaseDB() {
	t := time.Tick(5 * time.Second)
	dbs := make([]connPool, 0)
	for {
		select {
		case <-t:
			dbsTemp := make([]connPool, 0)
			for _, dbsql := range dbs {
				if dbsql.pool.Stats().OpenConnections == 0 {
					dbsql.pool.Close()
				} else {
					dbsTemp = append(dbsTemp, dbsql)
				}
			}
			dbs = dbsTemp
		case dbsql := <-db.releaseCh:
			dbs = append(dbs, dbsql)
		case <-db.closeCh:
			for _, dbsql := range dbs {
				dbsql.pool.Close()
			}
			return
		}
	}
}

func (db *MysqlShardDao) Close() {
	db.writeMux.Lock()
	db.readMux.Lock()
	db.closed = true
	if db.watcher != nil {
		db.watcher.Close()
	}
	db.readMux.Unlock()
	db.writeMux.Unlock()

	close(db.closeCh)

	for _, dbRead := range db.dbRead {
		for _, pool := range dbRead {
			pool.pool.Close()
		}
	}

	for _, dbWrite := range db.dbWrite {
		for _, pool := range dbWrite {
			pool.pool.Close()
		}
	}
}

// 设置最大连接数
func (db *MysqlShardDao) SetMaxOpenConns(maxOpenConns int) {
	db.writeMux.Lock()
	for _, dbWrite := range db.dbWrite {
		for _, pool := range dbWrite {
			pool.pool.SetMaxOpenConns(maxOpenConns)
		}
	}
	for _, wc := range db.writeConfig {
		wc.MaxOpenConns = maxOpenConns
	}
	db.writeMux.Unlock()

	db.readMux.Lock()
	for _, dbRead := range db.dbRead {
		for _, pool := range dbRead {
			pool.pool.SetMaxOpenConns(maxOpenConns)
		}
	}
	for _, wc := range db.readConfig {
		wc.MaxOpenConns = maxOpenConns
	}
	db.readMux.Unlock()
}

// 设置最大空闲连接数
func (db *MysqlShardDao) SetMaxIdleConns(maxIdleConns int) {
	db.writeMux.Lock()
	for _, dbWrite := range db.dbWrite {
		for _, pool := range dbWrite {
			pool.pool.SetMaxIdleConns(maxIdleConns)
		}
	}
	for _, wc := range db.writeConfig {
		wc.MaxIdleConns = maxIdleConns
	}
	db.writeMux.Unlock()

	db.readMux.Lock()
	for _, dbRead := range db.dbRead {
		for _, pool := range dbRead {
			pool.pool.SetMaxIdleConns(maxIdleConns)
		}
	}
	for _, wc := range db.readConfig {
		wc.MaxIdleConns = maxIdleConns
	}
	db.readMux.Unlock()
}

func (db *MysqlShardDao) GetWrite(key uint64) (*sql.DB, error) {
	var dbWrite *sql.DB

	db.writeMux.RLock()
	defer db.writeMux.RUnlock()

	if db.closed {
		return nil, ErrNoUseableDB
	}

	num := utils.GetShardDb(key, db.shardCount, db.shardAlg)
	poolSlice := db.dbWrite[num]
	length := len(poolSlice)
	if length > 0 {
		pos := db.writePos[num].Incr() % length
		for i := pos; i < pos+length; i++ {
			temppos := i % length
			dbWrite = poolSlice[temppos].pool
			if !db.filterFailConn {
				break
			}
			if dbWrite.Ping() == nil {
				logkit.Debugf("choose master:%s\n", poolSlice[temppos].addr)
				break
			} else {
				logkit.Errorf("[dao|mysql] mysql master:%s may be down", poolSlice[temppos].addr)
				dbWrite = nil
			}
		}
	}

	if dbWrite == nil {
		return nil, ErrNoUseableDB
	}

	return dbWrite, nil
}

// 插入数据
func (db *MysqlShardDao) Insert(key uint64, sqlstr string, args ...interface{}) (int64, error) {
	dbWrite, err := db.GetWrite(key)
	if err != nil {
		return 0, err
	}

	id, _, err := doWrite(dbWrite, sqlstr, args...)

	return id, err
}

// 更新数据
func (db *MysqlShardDao) Update(key uint64, sqlstr string, args ...interface{}) (int64, error) {
	dbWrite, err := db.GetWrite(key)
	if err != nil {
		return 0, err
	}

	_, num, err := doWrite(dbWrite, sqlstr, args...)

	return num, err
}

// 删除数据
func (db *MysqlShardDao) Delete(key uint64, sqlstr string, args ...interface{}) (int64, error) {
	dbWrite, err := db.GetWrite(key)
	if err != nil {
		return 0, err
	}

	_, num, err := doWrite(dbWrite, sqlstr, args...)

	return num, err
}

func (db *MysqlShardDao) GetRead(key uint64) (*sql.DB, error) {
	var dbRead *sql.DB

	db.readMux.RLock()
	defer db.readMux.RUnlock()

	if db.closed {
		return nil, ErrNoUseableDB
	}

	num := utils.GetShardDb(key, db.shardCount, db.shardAlg)
	poolSlice := db.dbRead[num]
	length := len(poolSlice)
	if length > 0 {
		pos := db.readPos[num].Incr() % length
		for i := pos; i < pos+length; i++ {
			temppos := i % length
			dbRead = poolSlice[temppos].pool
			if !db.filterFailConn {
				break
			}
			if dbRead.Ping() == nil {
				logkit.Debugf("choose slave:%s\n", poolSlice[temppos].addr)
				break
			} else {
				logkit.Errorf("[dao|mysql] mysql slave:%s may be down", poolSlice[temppos].addr)
				dbRead = nil
			}
		}
	}

	if dbRead == nil {
		return nil, ErrNoUseableDB
	}

	return dbRead, nil
}

// 取一行数据
func (db *MysqlShardDao) FetchRow(key uint64, sqlstr string, args ...interface{}) (map[string]string, error) {
	dbRead, err := db.GetRead(key)
	if err != nil {
		return nil, err
	}

	return readRow(dbRead, sqlstr, args...)
}

// 取多行数据
func (db *MysqlShardDao) FetchRows(key uint64, sqlstr string, args ...interface{}) ([]map[string]string, error) {
	dbRead, err := db.GetRead(key)
	if err != nil {
		return nil, err
	}

	return readRows(dbRead, sqlstr, args...)
}

// 从master取一行数据
func (db *MysqlShardDao) FetchRowForMaster(key uint64, sqlstr string, args ...interface{}) (map[string]string, error) {
	dbWrite, err := db.GetWrite(key)
	if err != nil {
		return nil, err
	}

	return readRow(dbWrite, sqlstr, args...)
}

// 从master取多行数据
func (db *MysqlShardDao) FetchRowsForMaster(key uint64, sqlstr string, args ...interface{}) ([]map[string]string, error) {
	dbWrite, err := db.GetWrite(key)
	if err != nil {
		return nil, err
	}

	return readRows(dbWrite, sqlstr, args...)
}

func (db *MysqlShardDao) SetFilterFailConn(filter bool) {
	db.writeMux.Lock()
	db.readMux.Lock()
	db.filterFailConn = filter
	db.readMux.Unlock()
	db.writeMux.Unlock()
}

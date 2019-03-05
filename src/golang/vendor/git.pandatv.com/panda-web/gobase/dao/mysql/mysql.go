package mysql

/*
 * 支持简单的增删改查，主从分别建立连接池
 * 如果主从是一个数据库请设置masterSlave为false,会只建立一个连接池
 * 调用示例见mysql_test.go
 */

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	janna "git.pandatv.com/panda-public/janna/client"
	"git.pandatv.com/panda-web/gobase/dao/watcher"
	"git.pandatv.com/panda-web/gobase/log"
	"github.com/go-sql-driver/mysql"
)

type connPool struct {
	pool *sql.DB
	addr string
	id   string
}

type MysqlBaseDao struct {
	dbWrite     []connPool
	dbRead      []connPool
	writeConfig *Config
	readConfig  *Config
	writeMux    *sync.RWMutex
	readMux     *sync.RWMutex
	writePos    *countInt32
	readPos     *countInt32

	encFlag bool

	watcher watcher.IWatcher

	releaseCh      chan connPool
	closed         bool
	closeCh        chan struct{}
	filterFailConn bool
}

type Config struct {
	*mysql.Config
	MaxOpenConns int
	MaxIdleConns int
	MaxLifeTime  time.Duration
}

const (
	DEFAULT_MAX_OPEN_CONNS = 50
	DEFAULT_MAX_IDLE_CONNS = 10
	DEFAULT_MAX_LIFE_TIME  = 0
	DEFAULT_MASTER_SLAVE   = true

	DEFAULT_TIMEOUT       = 2 * time.Second
	DEFAULT_READ_TIMEOUT  = 0
	DEFAULT_WRITE_TIMEOUT = 0
)

const (
	masterTag        = "master"
	slaveTag         = "slave"
	serviceNameMysql = "mysql"
)

var (
	ErrNoUseableDB = errors.New("no usealbe mysql")
)

func NewConfig() *Config {
	c := new(Config)
	c.Config = new(mysql.Config)

	c.Net = "tcp"
	c.Timeout = DEFAULT_TIMEOUT
	c.ReadTimeout = DEFAULT_READ_TIMEOUT
	c.WriteTimeout = DEFAULT_WRITE_TIMEOUT
	c.MaxOpenConns = DEFAULT_MAX_OPEN_CONNS
	c.MaxIdleConns = DEFAULT_MAX_IDLE_CONNS
	c.MaxLifeTime = DEFAULT_MAX_LIFE_TIME

	return c
}

func initialMysqlBase() *MysqlBaseDao {
	mysqlDao := new(MysqlBaseDao)

	mysqlDao.dbWrite = make([]connPool, 0)
	mysqlDao.dbRead = make([]connPool, 0)
	mysqlDao.writeMux = new(sync.RWMutex)
	mysqlDao.readMux = new(sync.RWMutex)
	mysqlDao.writePos = new(countInt32)
	mysqlDao.readPos = new(countInt32)

	mysqlDao.closed = false
	mysqlDao.closeCh = make(chan struct{})
	mysqlDao.releaseCh = make(chan connPool, 10)
	mysqlDao.filterFailConn = true

	return mysqlDao
}

func NewMysqlBaseDao(masterUser string, masterPwd string, masterServer string, masterPort string, masterDatabase string, masterMaxOpenConns int, masterMaxIdleConns int, slaveUser string, slavePwd string, slaveServer string, slavePort string, slaveDatabase string, slaveMaxOpenConns int, slaveMaxIdleConns int, masterSlave bool) (*MysqlBaseDao, error) {
	writeConfig := NewConfig()
	writeConfig.User = masterUser
	writeConfig.Passwd = masterPwd
	masterAddrPort := fmt.Sprintf("%s:%s", masterServer, masterPort)
	writeConfig.Addr = masterAddrPort
	writeConfig.DBName = masterDatabase
	writeConfig.MaxOpenConns = masterMaxOpenConns
	writeConfig.MaxIdleConns = masterMaxIdleConns

	readConfig := NewConfig()
	readConfig.User = slaveUser
	readConfig.Passwd = slavePwd
	slaveAddrPort := fmt.Sprintf("%s:%s", slaveServer, slavePort)
	readConfig.Addr = slaveAddrPort
	readConfig.DBName = slaveDatabase
	readConfig.MaxOpenConns = slaveMaxOpenConns
	readConfig.MaxIdleConns = slaveMaxIdleConns

	return NewMysqlBaseDaoCustom(writeConfig, readConfig, masterSlave)
}

// 创建连接
func conn(config *Config) (*sql.DB, error) {
	dsn := config.FormatDSN()

	dbsql, err := sql.Open("mysql", dsn)
	logkit.Debugf("[do] mysql_conn: dsn : %s, ", dsn)
	if err != nil {
		logkit.Errorf("init mysql %s error:%s", config.Addr, err)
		return nil, err
	}
	dbsql.SetMaxOpenConns(config.MaxOpenConns)
	dbsql.SetMaxIdleConns(config.MaxIdleConns)
	dbsql.SetConnMaxLifetime(config.MaxLifeTime)

	return dbsql, nil
}

// 自定义dns 链接，只有主库的话, slaveConfig传nil
func NewMysqlBaseDaoCustom(masterConfig *Config, slaveConfig *Config, masterSlave bool) (*MysqlBaseDao, error) {
	var err error
	MysqlClient := initialMysqlBase()
	MysqlClient.writeConfig = masterConfig
	if masterSlave {
		MysqlClient.readConfig = slaveConfig
	} else {
		MysqlClient.readConfig = masterConfig
	}

	mdbPool, err := conn(MysqlClient.writeConfig)
	if err != nil {
		logkit.Errorf("init mysql master %s error:%s", MysqlClient.writeConfig.Addr, err)
		return nil, err
	}
	MysqlClient.dbWrite = append(MysqlClient.dbWrite, connPool{
		pool: mdbPool,
		addr: MysqlClient.writeConfig.Addr,
		id:   "",
	})

	sdbPool, err := conn(MysqlClient.readConfig)
	if err != nil {
		logkit.Errorf("init mysql slave %s error:%s", MysqlClient.readConfig.Addr, err)
		return nil, err
	}
	MysqlClient.dbRead = append(MysqlClient.dbRead, connPool{
		pool: sdbPool,
		addr: MysqlClient.readConfig.Addr,
		id:   "",
	})

	return MysqlClient, err
}

func NewMysqlBaseDaoJannaCustom(servConf janna.ServiceClient, callName, targetTag string, masterConfig *Config, slaveConfig *Config) (*MysqlBaseDao, error) {
	return initJannaInstance(servConf, callName, targetTag, masterConfig, slaveConfig, false)
}

func NewMysqlBaseDaoJannaCustomEnc(servConf janna.ServiceClient, callName, targetTag string, masterConfig *Config, slaveConfig *Config) (*MysqlBaseDao, error) {
	return initJannaInstance(servConf, callName, targetTag, masterConfig, slaveConfig, true)
}

func initJannaInstance(servConf janna.ServiceClient, callName, targetTag string, masterConfig *Config, slaveConfig *Config, encFlag bool) (*MysqlBaseDao, error) {
	w, err := watcher.NewWatcher(servConf)
	if err != nil {
		logkit.Errorf("init watcher error:%s", err)
		return nil, err
	}
	mysqlInfos, err := w.GetAllInstance(callName, serviceNameMysql, targetTag)
	if err != nil {
		logkit.Errorf("get all mysql info from janna error:%s", err)
		return nil, err
	}

	MysqlClient := initialMysqlBase()
	MysqlClient.writeConfig = masterConfig
	MysqlClient.readConfig = slaveConfig
	MysqlClient.encFlag = encFlag
	MysqlClient.watcher = w

	go MysqlClient.releaseDB()

	for _, mysqlInfo := range mysqlInfos {
		addressPort := fmt.Sprintf("%s:%d", mysqlInfo.Address, mysqlInfo.Port)
		var conf *Config
		var serverPool *[]connPool
		if mysqlInfo.Master {
			MysqlClient.writeConfig.Addr = addressPort
			if MysqlClient.encFlag {
				MysqlClient.writeConfig.User = mysqlInfo.User
				MysqlClient.writeConfig.Passwd = mysqlInfo.Password
			}
			conf = MysqlClient.writeConfig
			serverPool = &(MysqlClient.dbWrite)
		} else {
			MysqlClient.readConfig.Addr = addressPort
			if MysqlClient.encFlag {
				MysqlClient.readConfig.User = mysqlInfo.User
				MysqlClient.readConfig.Passwd = mysqlInfo.Password
			}
			conf = MysqlClient.readConfig
			serverPool = &(MysqlClient.dbRead)
		}
		dbPool, err := conn(conf)
		if err != nil {
			logkit.Errorf("init mysql master:%t id:%s addr:%s error:%s", mysqlInfo.Master, mysqlInfo.Id, conf.Addr, err)
			return nil, err
		}
		logkit.Infof("add master:%t id:%s addr:%s", mysqlInfo.Master, mysqlInfo.Id, conf.Addr)
		*serverPool = append(*serverPool, connPool{
			pool: dbPool,
			addr: addressPort,
			id:   mysqlInfo.Id,
		})
	}

	go MysqlClient.jannaWatch(callName, targetTag)

	return MysqlClient, nil
}

func (db *MysqlBaseDao) jannaWatch(callName, targetTag string) {
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
		var conf *Config
		var serverPool *[]connPool
		var mux *sync.RWMutex
		if mysqlEvent.Master {
			conf = db.writeConfig
			serverPool = &(db.dbWrite)
			mux = db.writeMux
		} else {
			conf = db.readConfig
			serverPool = &(db.dbRead)
			mux = db.readMux
		}

		if mysqlEvent.Opt == janna.OptPut {
			addressPort := fmt.Sprintf("%s:%d", mysqlEvent.Address, mysqlEvent.Port)
			conf.Addr = addressPort
			if db.encFlag {
				conf.User = mysqlEvent.User
				conf.Passwd = mysqlEvent.Password
			}
			dbPool, _ := conn(conf)
			for i, pool := range *serverPool {
				if pool.id == mysqlEvent.Id {
					mux.Lock()
					logkit.Infof("rm master:%t id:%s addr:%s", mysqlEvent.Master, pool.id, pool.addr)
					db.releaseCh <- pool
					*serverPool = append((*serverPool)[0:i], (*serverPool)[i+1:]...)
					mux.Unlock()
					break
				}
			}
			mux.Lock()
			logkit.Infof("add master:%t id:%s addr:%s", mysqlEvent.Master, mysqlEvent.Id, addressPort)
			*serverPool = append(*serverPool, connPool{
				pool: dbPool,
				addr: addressPort,
				id:   mysqlEvent.Id,
			})
			mux.Unlock()
		} else if mysqlEvent.Opt == janna.OptDelete {
			for i, pool := range *serverPool {
				if pool.id == mysqlEvent.Id {
					mux.Lock()
					logkit.Infof("rm master:%t id:%s addr:%s", mysqlEvent.Master, pool.id, pool.addr)
					db.releaseCh <- pool
					*serverPool = append((*serverPool)[0:i], (*serverPool)[i+1:]...)
					mux.Unlock()
					break
				}
			}
		}
	}
}

func (db *MysqlBaseDao) releaseDB() {
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

func (db *MysqlBaseDao) Close() {
	db.writeMux.Lock()
	db.readMux.Lock()
	db.closed = true
	db.readMux.Unlock()
	db.writeMux.Unlock()

	if db.watcher != nil {
		err := db.watcher.Close()
		if err != nil {
			logkit.Errorf("close mysql janna watch error:%s", err)
		}
	}

	close(db.closeCh)

	for _, dbsql := range db.dbRead {
		dbsql.pool.Close()
	}

	for _, dbsql := range db.dbWrite {
		dbsql.pool.Close()
	}
}

// 设置最大连接数
func (db *MysqlBaseDao) SetMaxOpenConns(maxOpenConns int) {
	db.writeMux.Lock()
	db.writeConfig.MaxOpenConns = maxOpenConns
	for _, dbWrite := range db.dbWrite {
		dbWrite.pool.SetMaxOpenConns(maxOpenConns)
	}
	db.writeMux.Unlock()

	db.readMux.Lock()
	db.readConfig.MaxOpenConns = maxOpenConns
	for _, dbRead := range db.dbRead {
		dbRead.pool.SetMaxOpenConns(maxOpenConns)
	}
	db.readMux.Unlock()
}

// 设置最大空闲连接数
func (db *MysqlBaseDao) SetMaxIdleConns(maxIdleConns int) {
	db.writeMux.Lock()
	db.writeConfig.MaxIdleConns = maxIdleConns
	for _, dbWrite := range db.dbWrite {
		dbWrite.pool.SetMaxIdleConns(maxIdleConns)
	}
	db.writeMux.Unlock()

	db.readMux.Lock()
	db.readConfig.MaxIdleConns = maxIdleConns
	for _, dbRead := range db.dbRead {
		dbRead.pool.SetMaxIdleConns(maxIdleConns)
	}
	db.readMux.Unlock()
}

func (db *MysqlBaseDao) GetWrite() (*sql.DB, error) {
	var dbWrite *sql.DB

	db.writeMux.RLock()
	defer db.writeMux.RUnlock()

	length := len(db.dbWrite)
	if !db.closed && length > 0 {
		pos := db.writePos.Incr() % length
		for i := pos; i < pos+length; i++ {
			temppos := i % length
			dbWrite = db.dbWrite[temppos].pool
			if !db.filterFailConn {
				break
			}
			if dbWrite.Ping() == nil {
				logkit.Debugf("choose master:%s\n", db.dbWrite[temppos].addr)
				break
			} else {
				logkit.Errorf("[dao|mysql] mysql master:%s may be down", db.dbWrite[temppos].addr)
				dbWrite = nil
			}
		}
	}

	if dbWrite == nil {
		return nil, ErrNoUseableDB
	}

	return dbWrite, nil
}

func doWrite(dbWrite *sql.DB, sqlstr string, args ...interface{}) (int64, int64, error) {
	result, err := dbWrite.Exec(sqlstr, args...)
	if err != nil {
		return 0, 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, 0, err
	}

	num, err := result.RowsAffected()
	if err != nil {
		return 0, 0, err
	}

	return id, num, nil
}

// 插入数据
func (db *MysqlBaseDao) Insert(sqlstr string, args ...interface{}) (int64, error) {
	dbWrite, err := db.GetWrite()
	if err != nil {
		return 0, err
	}

	id, _, err := doWrite(dbWrite, sqlstr, args...)

	return id, err
}

// 更新数据
func (db *MysqlBaseDao) Update(sqlstr string, args ...interface{}) (int64, error) {
	dbWrite, err := db.GetWrite()
	if err != nil {
		return 0, err
	}

	_, num, err := doWrite(dbWrite, sqlstr, args...)

	return num, err
}

// 删除数据
func (db *MysqlBaseDao) Delete(sqlstr string, args ...interface{}) (int64, error) {
	dbWrite, err := db.GetWrite()
	if err != nil {
		return 0, err
	}

	_, num, err := doWrite(dbWrite, sqlstr, args...)

	return num, err
}

func (db *MysqlBaseDao) GetRead() (*sql.DB, error) {
	var dbRead *sql.DB

	db.readMux.RLock()
	defer db.readMux.RUnlock()

	length := len(db.dbRead)
	if !db.closed && length > 0 {
		pos := db.readPos.Incr() % length
		for i := pos; i < pos+length; i++ {
			temppos := i % length
			dbRead = db.dbRead[temppos].pool
			if !db.filterFailConn {
				break
			}
			if dbRead.Ping() == nil {
				logkit.Debugf("choose slave:%s\n", db.dbRead[temppos].addr)
				break
			} else {
				logkit.Errorf("[dao|mysql] mysql slave:%s may be down", db.dbRead[temppos].addr)
				dbRead = nil
			}
		}
	}

	if dbRead == nil {
		return nil, ErrNoUseableDB
	}

	return dbRead, nil

}

func readRow(dbsql *sql.DB, sqlstr string, args ...interface{}) (map[string]string, error) {
	rows, err := dbsql.Query(sqlstr, args...)

	if err != nil {
		return nil, err
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	ret := make(map[string]string)

	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		var value string
		for i, col := range values {
			if col == nil {
				value = "" //把数据表中所有为null的地方改成“”
			} else {
				value = string(col)
			}

			ret[columns[i]] = value
		}

		break
	}

	rows.Close()

	return ret, err
}

func readRows(dbsql *sql.DB, sqlstr string, args ...interface{}) ([]map[string]string, error) {
	rows, err := dbsql.Query(sqlstr, args...)

	if err != nil {
		return nil, err
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	var rets = make([]map[string]string, 0)

	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		var ret = make(map[string]string) //这里要注意(对语法的理解)

		var value string
		for i, col := range values {
			if col == nil {
				value = "" //把数据表中所有为null的地方改成“”
			} else {
				value = string(col)
			}

			ret[columns[i]] = value
		}

		rets = append(rets, ret)
	}

	return rets, err
}

// 取一行数据
func (db *MysqlBaseDao) FetchRow(sqlstr string, args ...interface{}) (map[string]string, error) {
	dbRead, err := db.GetRead()
	if err != nil {
		return nil, err
	}

	return readRow(dbRead, sqlstr, args...)
}

// 取多行数据
func (db *MysqlBaseDao) FetchRows(sqlstr string, args ...interface{}) ([]map[string]string, error) {
	dbRead, err := db.GetRead()
	if err != nil {
		return nil, err
	}

	return readRows(dbRead, sqlstr, args...)
}

// 从master取一行数据
func (db *MysqlBaseDao) FetchRowForMaster(sqlstr string, args ...interface{}) (map[string]string, error) {
	dbWrite, err := db.GetWrite()
	if err != nil {
		return nil, err
	}

	return readRow(dbWrite, sqlstr, args...)
}

// 从master取多行数据
func (db *MysqlBaseDao) FetchRowsForMaster(sqlstr string, args ...interface{}) ([]map[string]string, error) {
	dbWrite, err := db.GetWrite()
	if err != nil {
		return nil, err
	}

	return readRows(dbWrite, sqlstr, args...)
}

func (db *MysqlBaseDao) SetFilterFailConn(filter bool) {
	db.writeMux.Lock()
	db.readMux.Lock()
	db.filterFailConn = filter
	db.readMux.Unlock()
	db.writeMux.Unlock()
}

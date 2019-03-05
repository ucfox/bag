#建议通过janna_dao使用gobase [janna_dao](https://git.pandatv.com/panda-public/janna_dao)

#### 普通客户端使用说明

* 普通方式

```go
func NewMysqlBaseDao(masterUser string, masterPwd string, masterServer string, masterPort string, masterDatabase string, masterMaxOpenConns int, masterMaxIdleConns int, slaveUser string, slavePwd string, slaveServer string, slavePort string, slaveDatabase string, slaveMaxOpenConns int, slaveMaxIdleConns int, masterSlave bool) (*MysqlBaseDao, error)
```

1. `masterMaxOpenConns` 连接池拥有的最大连接数
2. `masterMaxIdleConns` 连接池的空闲连接
3. `masterSlave`        是否使用主从分离, false不分离


* 自定义方式

```go
func NewMysqlBaseDaoCustom(masterConfig *Config, slaveConfig *Config, masterSlave bool) (*MysqlBaseDao, error)
```

```go
定义默认参数
func NewConfig() *Config

type Config struct {
    *mysql.Config
	MaxOpenConns int
	MaxIdleConns int
	MaxLifeTime  time.Duration
}

type mysql.Config struct {
    User             string            // Username
    Passwd           string            // Password (requires User)
    Net              string            // Network type
    Addr             string            // Network address (requires Net)
    DBName           string            // Database name
    Params           map[string]string // Connection parameters
    Collation        string            // Connection collation
    Loc              *time.Location    // Location for time.Time values
    MaxAllowedPacket int               // Max packet size allowed
    TLSConfig        string            // TLS configuration name

    Timeout      time.Duration // Dial timeout
    ReadTimeout  time.Duration // I/O read timeout
    WriteTimeout time.Duration // I/O write timeout

    AllowAllFiles           bool // Allow all files to be used with LOAD DATA LOCAL INFILE
    AllowCleartextPasswords bool // Allows the cleartext client side plugin
    AllowNativePasswords    bool // Allows the native password authentication method
    AllowOldPasswords       bool // Allows the old insecure password method
    ClientFoundRows         bool // Return number of matching rows instead of rows changed
    ColumnsWithAlias        bool // Prepend table alias to column names
    InterpolateParams       bool // Interpolate placeholders into query string
    MultiStatements         bool // Allow multiple statements in one query
    ParseTime               bool // Parse time values to time.Time
    Strict                  bool // Return warnings as errors
}
```

* 使用janna

```go
func NewMysqlBaseDaoJannaCustom(servConf janna.ServiceClient, callName, targetTag string, masterConfig *Config, slaveConfig *Config) (*MysqlBaseDao, error)
```
```go
func NewMysqlBaseDaoJannaCustomEnc(servConf janna.ServiceClient, callName, targetTag string, masterConfig *Config, slaveConfig *Config) (*MysqlBaseDao, error)
```


1. `servConf` 实现了janna ServiceClient接口的实例
2. `callName` 使用方的名字，例如是业务test将使用mysql
3. `targetTag` 带有目标tag的实例会被获取，用来区分业务使用多套mysql的情况

#### 注：git.pandatv.com/panda-web/gobase/janna/etcd/mysql也可初始化客户端，集成了etcd的初始化及mysql的初始化，详情参考相关README

* 关闭客户端

```go
func (db *MysqlBaseDao) Close()
```

* 设置最大连接数

```go
func (db *MysqlBaseDao) SetMaxOpenConns(maxOpenConns int)
```

* 设置最大空闲连接数

```go
func (db *MysqlBaseDao) SetMaxIdleConns(maxIdleConns int)
```

* 除CRUD的操作，可以直接以下方法获取DB去实现

```go
func (db *MysqlBaseDao) GetWrite() (*sql.DB, error)
func (db *MysqlBaseDao) GetRead() (*sql.DB, error)
```

#### 常量

```go
const (
	DEFAULT_MAX_OPEN_CONNS = 50
	DEFAULT_MAX_IDLE_CONNS = 10
	DEFAULT_MAX_LIFE_TIME  = time.Second * 10
	DEFAULT_MASTER_SLAVE   = true

	DEFAULT_TIMEOUT       = 2 * time.Second
	DEFAULT_READ_TIMEOUT  = 0
	DEFAULT_WRITE_TIMEOUT = 0
)
```

### 调用示例CRUD

- 见mysql_test.go
- 支持链接超时，读写超时。用法参见测试用例Test_Custom

#### 注：

- 当日志级别为LevelDebug时，会在日志记录中记录产生的sql



#### Sharding客户端使用说明

* 普通方式

```go
func NewMysqlShardDao(masterConfig []*ShardConfig, slaveConfig []*ShardConfig, shardCount uint64, shardAlg int) (*MysqlShardDao, error)
```

1. `masterConfig` 若干主库的配置，包括分片规则
2. `slaveConfig` 若干从库的配置，包括分片规则
3. `shardCount` 总的分片数
4. `shardAlg` 分片算法，gobase/utils中定义了相应常量，AlgNone=0 AlgCrc32=1

```go
定义默认参数
func NewShardConfig() *ShardConfig

type ShardConfig struct {
    *Config             //上面有定义
    //区间左闭又闭
	ShardStart  uint64  //起始分片号
	ShardEnd    uint64  //终止分片号
    InstanceKey string  //janna初始化时使用该参数，该key是etcd中的key，指向一个实例，实例的地址将替代Config地址
}
```

#### 特别注意：sharding的初始化，最终的DBName是由传入的DBName及分片序号的组合，举例如下

```go
sc := NewShardConfig()
sc.DBName = "testdb"
sc.ShardStart = 0
sc.ShardEnd = 3
sc.Addr = "127.0.0.1:3306"


客户端初始化后，映射结果是
0 -> 127.0.0.1:3306/testdb_0
1 -> 127.0.0.1:3306/testdb_1
2 -> 127.0.0.1:3306/testdb_2
3 -> 127.0.0.1:3306/testdb_3
```

* 使用janna

```go
func NewMysqlShardDaoJanna(servConf janna.ServiceClient, callName, targetTag string, masterConfig []*ShardConfig, slaveConfig []*ShardConfig, shardCount uint64, shardAlg int) (*MysqlShardDao, error)
```
```go
func NewMysqlShardDaoJannaEnc(servConf janna.ServiceClient, callName, targetTag string, masterConfig []*ShardConfig, slaveConfig []*ShardConfig, shardCount uint64, shardAlg int) (*MysqlShardDao, error)
```


1. `servConf` 实现了janna ServiceClient接口的实例
2. `callName` 使用方的名字，例如是业务test将使用mysql
3. `targetTag` 带有目标tag的实例会被获取，用来区分业务使用多套mysql的情况

#### 注：git.pandatv.com/panda-web/gobase/janna/etcd/mysql也可初始化客户端，集成了etcd的初始化及mysql的初始化，详情参考相关README

* 获取主库

```go
func (db *MysqlShardDao) GetWrite(key uint64) (*sql.DB, error)
```

* 获取从库

```go
func (db *MysqlShardDao) GetRead(key uint64) (*sql.DB, error)
```

#### 注：CRUD等方法也需增加key参数，用以做hash，详细使用方法可参考单元测试

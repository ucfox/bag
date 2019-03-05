#### 使用说明
#### 普通客户端初始化
* 初始化

```go
func NewMysqlBaseDao(endPoints, callName, targetTag, masterUser, masterPwd, masterDatabase, slaveUser, slavePwd, slaveDatabase string) (*MysqlBaseDao, error)
```
```go
func NewMysqlBaseDaoEnc(endPoints, callName, targetTag, encryCode, masterDatabase, slaveDatabase string) (*MysqlBaseDao, error)
```
1. `endPoints` etcd的ip port，多个ip port用逗号隔开，如bjac.etcd1.beta.pdtv.io:2379,bjac.etcd2.beta.pdtv.io:2379,bjac.etcd3.beta.pdtv.io:2379
2. `callName` 使用方的名字，如villa
3. `targetTag` tag标签，如web,cache等，由用户自定义，需和etcd对应
4. `encryCode` 密钥(8位)，用户自定义


* 自定义连接参数，初始化客户端

```go
func NewMysqlBaseDaoCustom(endPoints, callName, targetTag string, masterConfig *Config, slaveConfig *Config) (*MysqlBaseDao, error)
```
```go
func NewMysqlBaseDaoCustomEnc(endPoints, callName, targetTag, encryCode string, masterConfig *Config, slaveConfig *Config) (*MysqlBaseDao, error)
```

#### Sharding客户端初始化
* 初始化客户端

```go
func NewMysqlShardDao(endPoints, callName, targetTag string, masterConfig []*ShardConfig, slaveConfig []*ShardConfig, shardCount uint64, shardAlg int) (*MysqlShardDao, error)
```
```go
func NewMysqlShardDaoEnc(endPoints, callName, targetTag, encryCode string, masterConfig []*ShardConfig, slaveConfig []*ShardConfig, shardCount uint64, shardAlg int) (*MysqlShardDao, error)
```

1. `etcdUser` etcd用户名
2. `etcdPwd` etcd密码
3. `encryCode` 密钥(8位)，用户自定义

注：参数含义与gobase/dao/mysql sharding的一致


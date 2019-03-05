#### 使用说明

* 初始化客户端

```go
func NewRedisClient(endPoints string, callName, targetTag, masterPwd, slavePwd string) *RedisBaseDao
```
```go
func NewRedisClientEnc(endPoints string, callName, targetTag, encryCode string) *RedisBaseDao
```

1. `endPoints` etcd的ip port，多个ip port用逗号隔开，如bjac.etcd1.beta.pdtv.io:2379,bjac.etcd2.beta.pdtv.io:2379,bjac.etcd3.beta.pdtv.io:2379
2. `callName` 使用方的名字，如villa
3. `targetTag` tag标签，如web,cache等，由用户自定义，需和etcd对应
4. `encryCode` 密钥(8位)，由用户自定义

* 自定义连接参数，初始化客户端

```go
func NewRedisClientCustom(endPoints, callName, targetTag, masterPwd, slavePwd string, connectTimeout, readTimeout, writeTimeout, idleTimeout time.Duration, maxActive, maxIdle int, waitConn bool) *RedisBaseDao
```
```go
func NewRedisClientCustomEnc(endPoints, callName, targetTag, masterPwd, slavePwd, encryCode string, masterConfig *Config, slaveConfig *Config) *RedisBaseDao
```

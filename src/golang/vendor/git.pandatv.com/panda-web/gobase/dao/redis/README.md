#### 使用说明
#### 常量及结构体
```go
const (
    Default_connect_timeout = 2 * time.Second
    Default_read_timeout    = 500 * time.Millisecond
    Default_write_timeout   = 500 * time.Millisecondd
    Default_idle_timeout    = 60 * time.Second
    Default_max_active      = 50
    Default_max_idle        = 3
    Default_wait            = true
)

var ErrNil = redis.ErrNil

type ScorePair struct {
    Member string `json:"member"`
    Score  int    `json:"ScorePair"`
}

type SubscribeT struct {
    //...
    C    chan []byte
}
```

#### 初始化客户端
```go
func NewRedisClient(master string, masterPwd string, slave string, slavePwd string) *RedisBaseDao
```

cliPool := NewRedisClient("10.20.1.20:6379", "dfjallf", "10.20.1.20:7300,10.20.1.20:6380", "fdjfla")

按默认参数初始化客户端，slave可传多个`ip:port`，用`,`隔开

#### 定制参数初始化客户端
```go
func NewRedisClientCustom(master string, masterPwd string, slave string, slavePwd string, connectTimeout, readTimeout, writeTimeout, idleTimeout time.Duration, maxActive, maxIdle int, waitConn bool) *RedisBaseDao
```

cliPool := NewRedisClient("10.20.1.20:6379", "dfjallf", "10.20.1.20:7300,10.20.1.20:6380", "fdjfla", time.Second, time.Second, time.Second, 10*time.Second, 50, 10, true)

- `connectTimeout` 连接超时
- `readTimeout` 读超时
- `writeTimeout` 写超时
- `idleTimeout` 空闲连接被删除的时间
- `maxActive` 池拥有的最大连接数，包括idle的
- `maxIdle` 最大空闲的连接
- `waitConn` 达到最大连接时，true：将等待可用连接，false：将返回并给出异常

#### 使用janna初始化客户端
```go
func NewRedisClientJanna(servConf janna.ServiceClient, callName, targetTag, masterPwd, slavePwd string) *RedisBaseDao
```
```go
func NewRedisClientJannaEnc(servConf janna.ServiceClient, callName, targetTag string) *RedisBaseDao
```


cliPool := NewRedisClientJanna(servConf, "test", "default", "fdfaf", "fdafd")

- `serConf` 实现了janna ServiceClient接口的实例
- `callName` 使用方的名字，例如是业务test将使用redis
- `targetTag` 带有目标tag的实例会被获取，用来区分业务使用多套redis的情况

#### 使用janna初始化客户端，并定制参数
```go
func NewRedisClientJannaCustom(servConf janna.ServiceClient, callName, targetTag, masterPwd, slavePwd string, connectTimeout, readTimeout, writeTimeout, idleTimeout time.Duration, maxActive, maxIdle int, waitConn bool) *RedisBaseDao
```

cliPool := NewRedisClientJanna(servConf, "test", "default", "fdfaf", "fdafd", time.Second, time.Second, time.Second, 10*time.Second, 50, 10, true)

#### 注：git.pandatv.com/panda-web/gobase/janna/etcd/redis也可初始化客户端，集成了etcd的初始化及redis的初始化，详情参考相关README

#### 一些方法的说明
```go
//Scan
func (dao *RedisBaseDao) Scan(cursor int, pattern string, count int) (int, []string, error)
```

参数：
- `cursor` 游标
- `pattern` 正则字符串，可以为空
- `count` 期望返回的个数，如果为0，个数由redis决定

返回值：
- `int` 下一次迭代的游标，如果为0，表示迭代完成
- `[]string` 得到的key

```go
//Get,GetRaw
func (dao *RedisBaseDao) Get(key string) (string, error)
func (dao *RedisBaseDao) GetRaw(key string) (string, error)
```

如果key不存在，Get返回空串，无错误，GetRaw返回空串，ErrNil错误

```go
//Set
func (dao *RedisBaseDao) Set(key, val string, ttl int) (string, error)
```

Set需传ttl，如果是0，则不过期

```go
//HScan
func (dao *RedisBaseDao) HScan(key string, cursor int, pattern string, count int) (int, map[string]string, error)
```

参数：
cursor，pattern，count与Scan意义相同，pattern是对field作匹配

返回值：
- `int` 下一次迭代的游标，如果为0，表示迭代完成
- `map[string]string` 得到field，value

```go
//SScan
func SScan(key string, cursor int, pattern string, count int) (int, []string, error)
```

参数：
cursor，pattern，count与Scan意义相同，pattern是对member作匹配

返回值：
- `int` 下一次迭代的游标，如果为0，表示迭代完成
- `[]string` 得到的member

```go
//Script,LoadScript,Eval
func (dao *RedisBaseDao) Script(keyCount int, src string) *LuaScript
func (dao *RedisBaseDao) LoadScript(script *LuaScript) error
func (dao *RedisBaseDao) Eval(script *LuaScript, keysAndArgs ...interface{}) (interface{}, error)
```

```
Demo：
    构造lua脚本对象
    script := cliPool.Script(1, `return redis.call('get', KEYS[1])`)
    上传脚本
    err := cliPool.LoadScript(script)
    if err != nil {
        fmt.Println(err)
        return
    }
    val, err := cliPool.Eval(script, "foo")
    if err != nil {
        fmt.Println(err)
        return
    }

    fmt.Println(string(val.([]byte)))
```

注意：lua脚本执行均在master上

参数：

Script：
- `keyCount` lua脚本字符串中参数的个数
- `src` lua脚本字符串

LoadScript：
- `script` lua脚本对象

Eval：
- `script` lua脚本对象
- `keysAndArgs` 执行需要的参数

返回值：

Script：
- `LuaScript` lua脚本对象

Eval：
- `interface{}` 脚本执行的结果，为空接口类型，用户解析

```go
//PipeLine,PipeSend,PipeExec,PipeClose
func (dao *RedisBaseDao) PipeLine(readonly bool) (*Pipe, error)
func (dao *RedisBaseDao) PipeSend(pipe *Pipe, cmd string, args ...interface{}) error
func (dao *RedisBaseDao) PipeExec(pipe *Pipe) (interface{}, error)
func (dao *RedisBaseDao) PipeClose(pipe *Pipe) error
```

```
Demo:
    构造pipe对象
    pipe, err := cliPool.PipeLine(true)
    if err != nil {
        fmt.Println(err)
        return
    }
    添加指令
    err = cliPool.PipeSend(pipe, "GET", "foo")
    if err != nil {
        fmt.Println(err)
        return
    }
    err = cliPool.PipeSend(pipe, "SET", "foo", "bar")
    if err != nil {
        fmt.Println(err)
        return
    }
    执行执行
    val, err := cliPool.PipeExec(pipe)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(val)
    释放对象
    err = cliPool.PipeClose(pipe)
    if err != nil {
        fmt.Println(err)
        return
    }
```

参数：

PipeLine：
- `readonly` true会将指令发送到从，false，会将指令发送到主

PipeSend：
- `pipe` pipe对象
- `cmd` 待执行的指令
- `args` 执行需要的参数

PipeExec：
- `pipe` pipe对象

PipeClose：
- `pipe` pipe对象

返回值：

PipeLine：
- `Pipe` pipe对象

PipeExec：
- `interface{}` pipe执行的结果，为空接口类型，由用户解析

```go
//Publish,Subscribe,SubscribeClose
func (dao *RedisBaseDao) Publish(channel, message string) (int, error)
func (dao *RedisBaseDao) Subscribe(channel string) (*SubscribeT, error)
func (dao *RedisBaseDao) SubscribeClose(st *SubscribeT) error
```

```
Demo:
    向test_channel发送消息
    valI, err := cliPool.Publish("test_channel", "haha")
    if err != nil {
        fmt.Println(err)
        return
    }

Demo:
    订阅__keyevent@0__:set频道
    st, err := cliPool.Subscribe("__keyevent@0__:set")
    if err != nil {
        fmt.Println(err)
        return
    }
    获取内容，如果连接正常，channel不会关闭
    for key := range st.C {
        fmt.Println(key)
    }
    err = cliPool.SubscribeClose(st)
    if err != nil {
        fmt.Println(err)
        return
    }
```

参数：

Publish：
- `channel` 频道
- `message` 消息内容

Subscribe：
- `channel` 频道

SubscribeClose：
- `st` 订阅得到的对象

返回值：

Publish：
- `int` 接受到消息的订阅者的数量

Subscribe：
- `SubscribeT` 订阅频道生成的对象


#### 注：

- 读取的相关方法，做了ErrNIl屏蔽；因为ZSCORE返回的值的原因，没有做屏蔽；屏蔽带来的影响，见源码中相关方法的注释


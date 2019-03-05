#### 使用说明

#### 常量及错误

```go
const (
    OptPut    = "PUT"
    OptDelete = "DELETE"
)

var ErrFormatKey = errors.New("key is wrong format")
```

#### 结构体

```go
type Service struct {
    Key     string   //must
    Tag             []string //optional
    Address string   //optional
    Port    int      //optional
    Weight  int      //optional
}

type ServiceEvent struct {
    Opt string
    Service
}

type Kv struct {
    Key   string
    Value string
}

type KvEvent struct {
    Opt string
    Kv
}
```

#### 接口

```go
type Client interface {
    ServiceClient
    KvClient
    Close() error
}

type KvClient interface {
    KvPut(string, string) error
    KvGet(string) (string, error)
    KvDelete(string) error
    KvAddWatch(string) error
    KvGetRemoveWatch(string) error
    KvGetWatch() (chan KvEvent, error)
    KvCloseWatch() error
}

type ServiceClient interface {
    ServiceRegister(*Service) error
    ServiceDeregister(string) error
    ServiceGet(string) (*Service, error)Service
    ServiceGetAll(string) ([]Service, error)
    ServiceAddWatch(string) error
    ServiceRemoveWatch(string) error
    ServiceGetWatch() (chan ServiceEvent, error)
    ServiceCloseWatch() error
}
```

#### key的生成与解析

* kv key生成

```go
func KvKey(callName, subKey string) string

callName: 使用方的名字
subKey: 自定义的子key
```

* kv key解析

```go
func SplitKvKey(key string) (string, string, error)
```

* service prefix key生成

```go
//可用于获取所有服务列表
func ServicePrefixKey(callName, serviceName string) string

callName: 使用方的名字
serviceName: 依赖的服务名
```

* service prefix key解析

```go
func SplitServicePrefixKey(key string) (string, string, error)
```

* service key生成

```go
//指向具体的服务实例
func ServiceKey(callName, serviceName, serviceKey string) string

callName: 使用方的名字
serviceName: 依赖的服务名
serviceKey: 具体服务的key，用以区分实例
```

* service key解析

```go
func SplitServiceKey(key string) (string, string, string, error)
```

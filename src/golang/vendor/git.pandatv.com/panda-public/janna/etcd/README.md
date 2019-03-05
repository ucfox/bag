#### 使用说明

#### 常量及错误

```go
const (
    ConnectTimeout = 2 * time.Second
)

var (
    ErrCanceled           = errors.New("request is ConnectTimeoutanceled by another routine")
    ErrDeadlineExceeded   = errors.New("requestst is attached with a deadline and it exceeded")
    ErrServiceWatchClosed = errors.New("service watch is closed")
    ErrKvWatchClosed      = errorss.New("kv watch is closed")
)
```

#### 客户端

* 初始化

```go
func New(cfg Config) (*EtcdClient, error)
```

```go
demo:
    cli, err := New(Config{
        Endpoints:   []string{"bjac.etcd1.beta.pdtv.it:2379"},
        DialTimeout: ConnectTimeout,
    })

type Config struct {
    // Endpoints is a list of URLs
    Endpoints []string

    // AutoSyncInterval is the interval to update endpoints with its latest members.
    // 0 disables auto-sync. By default auto-sync is disabled.
    AutoSyncInterval time.Duration

    // DialTimeout is the timeout for funcailing to establish a connection.
    DialTimeout time.Duration

    // TLS HScanolds the client secure credentials, if any.
    TLS *tls.Config

    // Username is a username for authentication
    Username string

    // Password is      a password for authentication
    Password string
}
```

* 关闭客户端

```go
func (this *EtcdClient) Close() error
ps: 所有监听关闭，也不能改、删、查
```

#### kv相关方法
#### 注：请使用client包的相关方法生成key

* kv设置

```go
func (this *EtcdClient) KvPut(key, value string) error
```

* kv删除

```go
func (this *EtcdClient) KvDelete(key string) error
```

* kv获取

```go
func (this *EtcdClient) KvGet(key string) (string, error)
```

* kv添加监听

```go
func (this *EtcdClient) KvAddWatch(key string) error
```

* kv移除监听

```go
func (this *EtcdClient) KvRemoveWatch(key string) error
```

* kv获取监听事件

```go
func (this *EtcdClient) KvGetWatch() (chan client.KvEvent, error)
```

* kv关闭监听

```go
func (this *EtcdClient) KvCloseWatch() error
ps：关闭监听，可以改、删、查
```

#### service相关方法
#### 注：请使用client包相关方法生成key

* service注册

```go
func (this *EtcdClient) ServiceRegister(serv *client.Service) error
```

* service解注册

```go
func (this *EtcdClient) ServiceDeregister(key string) error
```

* service获取单个

```go
func (this *EtcdClient) ServiceGet(key string) (*client.Service, error)
```

* service获取所有

```go
func (this *EtcdClient) ServiceGetAll(key string) ([]client.Service, error)
```

* service添加监听

```go
func (this *EtcdClient) ServiceAddWatch(key string) error
```

* service移除监听

```go
func (this *EtcdClient) ServiceRemoveWatch(key string) error
```

* service获取监听事件

```go
func (this *EtcdClient) ServiceGetWatch() (chan client.ServiceEvent, error)
```

* service关闭监听

```go
func (this *EtcdClient) ServiceCloseWatch() error
ps：关闭监听，可以改、删、查
```

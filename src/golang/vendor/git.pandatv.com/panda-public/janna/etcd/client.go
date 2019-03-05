package etcd

import (
	"errors"
	"sync"
	"time"

	"git.pandatv.com/panda-public/janna/client"
	"git.pandatv.com/panda-public/janna/xcrypto"
	"github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	ConnectTimeout = 2 * time.Second
	writeTimeout   = time.Second
	readTimeout    = 2 * time.Second
	watchTimeout   = time.Minute * 5
	notifyInterval = time.Minute * 21 //10 * 2 + 1
)

var (
	ErrCanceled           = errors.New("request is canceled by another routine")
	ErrDeadlineExceeded   = errors.New("request is attached with a deadline and it exceeded")
	ErrServiceWatchClosed = errors.New("service watch is closed")
	ErrKvWatchClosed      = errors.New("kv watch is closed")
	ErrSyncEtcdMember     = errors.New("sync etcd member fail")
	ErrEmptyEndpoints     = errors.New("empty etcd enpoints")
)

/*
//clientv3.Config detail
type Config struct {
	// Endpoints is a list of URLs
	Endpoints []string

	// AutoSyncInterval is the interval to update endpoints with its latest members.
	// 0 disables auto-sync. By default auto-sync is disabled.
	AutoSyncInterval time.Duration

	// DialTimeout is the timeout for failing to establish a connection.
	DialTimeout time.Duration

	// TLS holds the client secure credentials, if any.
	TLS *tls.Config

	// Username is a username for authentication
	Username string

	// Password is a password for authentication
	Password string
}
*/
type Config clientv3.Config

type ContextCancel struct {
	ctx     context.Context
	cancel  context.CancelFunc
	version int64
	tm      *time.Timer
}

type EtcdClient struct {
	baseClient *clientv3.Client
	kvClient   clientv3.KV

	serviceWatchClient clientv3.Watcher
	serviceContext     map[string]*ContextCancel
	serviceMux         *sync.Mutex
	serviceChan        chan client.ServiceEvent
	serviceWatchClosed bool

	kvWatchClient clientv3.Watcher
	kvContext     map[string]*ContextCancel
	kvMux         *sync.Mutex
	kvChan        chan client.KvEvent
	kvWatchClosed bool

	crypter xcrypto.Crypter
}

func New(cfg Config) (*EtcdClient, error) {
	if len(cfg.Endpoints) == 0 {
		return nil, ErrEmptyEndpoints
	}

	var cli *clientv3.Client
	var err error
	eps := cfg.Endpoints
	for i := 0; i < len(eps); i++ {
		cfg.Endpoints = eps[i : i+1]
		cli, err = clientv3.New(clientv3.Config(cfg))
		if err == nil {
			break
		}
		if err == grpc.ErrClientConnTimeout {
			continue
		} else {
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	err = cli.Sync(ctx)
	cancel()
	if err != nil {
		cli.Close()
		return nil, ErrSyncEtcdMember
	}

	etcdClient := new(EtcdClient)
	etcdClient.baseClient = cli
	etcdClient.kvClient = clientv3.NewKV(cli)

	etcdClient.serviceWatchClient = clientv3.NewWatcher(cli)
	etcdClient.serviceContext = make(map[string]*ContextCancel)
	etcdClient.serviceMux = new(sync.Mutex)
	etcdClient.serviceChan = make(chan client.ServiceEvent, 100)
	etcdClient.serviceWatchClosed = false

	etcdClient.kvWatchClient = clientv3.NewWatcher(cli)
	etcdClient.kvContext = make(map[string]*ContextCancel)
	etcdClient.kvMux = new(sync.Mutex)
	etcdClient.kvChan = make(chan client.KvEvent, 100)
	etcdClient.kvWatchClosed = false

	return etcdClient, nil
}

func NewWithDes(cfg Config, key []byte) (*EtcdClient, error) {
	dc, err := xcrypto.NewDes(key)
	if err != nil {
		return nil, err
	}

	etcdClient, err := New(cfg)
	if err != nil {
		return nil, err
	}

	etcdClient.crypter = dc
	return etcdClient, nil
}

func (this *EtcdClient) Close() error {
	this.KvCloseWatch()
	this.ServiceCloseWatch()
	return this.baseClient.Close()
}

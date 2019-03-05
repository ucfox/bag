package etcd

import (
	"encoding/base64"
	"time"

	"git.pandatv.com/panda-public/janna/client"
	"git.pandatv.com/panda-web/gobase/log"
	"github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
)

func (this *EtcdClient) KvPut(key, value string, kvopts ...client.KvOption) error {
	opt := new(client.KvOpt)
	for _, kvopt := range kvopts {
		kvopt(opt)
	}

	if opt.Crypt && this.crypter != nil {
		valueByte, err := this.crypter.Encrypt([]byte(value))
		if err != nil {
			return err
		}
		value = base64.StdEncoding.EncodeToString(valueByte)
	}

	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	_, err := this.kvClient.Put(ctx, key, value)
	cancel()
	if err != nil {
		logkit.Errorf("put key:%s value:%s error:%s", key, value, err)
		return checkEtcdError(err)
	}

	return nil
}

func (this *EtcdClient) KvDelete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	_, err := this.kvClient.Delete(ctx, key)
	cancel()
	if err != nil {
		logkit.Errorf("delete key:%s error:%s", key, err)
		return checkEtcdError(err)
	}

	return nil
}

func (this *EtcdClient) KvGet(key string, kvopts ...client.KvOption) (string, error) {
	opt := new(client.KvOpt)
	for _, kvopt := range kvopts {
		kvopt(opt)
	}

	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	resp, err := this.kvClient.Get(ctx, key)
	cancel()
	if err != nil {
		logkit.Errorf("get key %s error:%s", key, err)
		return "", checkEtcdError(err)
	}

	if len(resp.Kvs) == 0 {
		return "", nil
	}

	kv, err := this.decodeKv(opt, resp.Kvs[0].Key, resp.Kvs[0].Value)
	if err != nil {
		return "", err
	}

	return kv.Value, nil
}

func (this *EtcdClient) KvGetAll(key string, kvopts ...client.KvOption) ([]client.Kv, error) {
	opt := new(client.KvOpt)
	for _, kvopt := range kvopts {
		kvopt(opt)
	}

	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	resp, err := this.kvClient.Get(ctx, key, clientv3.WithPrefix())
	cancel()
	if err != nil {
		logkit.Errorf("get keys %s error:%s", key, err)
		return nil, checkEtcdError(err)
	}

	kvList := make([]client.Kv, 0)

	if len(resp.Kvs) == 0 {
		return kvList, nil
	}

	for _, rkv := range resp.Kvs {
		kv, err := this.decodeKv(opt, rkv.Key, rkv.Value)
		if err != nil {
			return nil, err
		}
		kvList = append(kvList, *kv)
	}

	return kvList, nil
}

func (this *EtcdClient) KvAddWatch(key string, kvopts ...client.KvOption) error {
	opt := new(client.KvOpt)
	for _, kvopt := range kvopts {
		kvopt(opt)
	}

	this.kvMux.Lock()
	defer this.kvMux.Unlock()

	if this.kvWatchClosed {
		return ErrKvWatchClosed
	}

	_, ok := this.kvContext[key]
	if ok {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	//ctx, cancel := context.WithTimeout(context.Background(), watchTimeout)
	this.kvContext[key] = &ContextCancel{ctx, cancel, 0, time.NewTimer(notifyInterval)}
	go func(ct *ContextCancel) {
		select {
		case <-ct.ctx.Done():
			return
		case <-ct.tm.C:
			ct.cancel()
		}
	}(this.kvContext[key])

	respCh := this.kvWatchClient.Watch(ctx, key, clientv3.WithCreatedNotify(), clientv3.WithProgressNotify(), clientv3.WithPrevKV())

	go this.kvWatch(key, respCh, opt)

	return nil
}

func (this *EtcdClient) KvRemoveWatch(key string) error {
	this.kvMux.Lock()
	defer this.kvMux.Unlock()

	ctxCancel, ok := this.kvContext[key]
	if !ok {
		return nil
	}

	ctxCancel.cancel()
	ctxCancel.tm.Stop()
	delete(this.kvContext, key)

	return nil
}

func (this *EtcdClient) KvGetWatch() (chan client.KvEvent, error) {
	return this.kvChan, nil
}

func (this *EtcdClient) kvWatch(key string, respCh clientv3.WatchChan, opt *client.KvOpt) {
	for resp := range respCh {
		if resp.Err() != nil {
			if !resp.Canceled {
				logkit.Errorf("watch key:%s response error:%s", key, resp.Err())
			}
			time.Sleep(time.Millisecond * 100)
			break
		}

		this.kvMux.Lock()
		ctxCancel, ok := this.kvContext[key]
		if ok {
			ctxCancel.version = resp.Header.Revision
			ctxCancel.tm.Reset(notifyInterval)
		}
		this.kvMux.Unlock()

		for _, ev := range resp.Events {
			if ev.Type.String() == client.OptPut {
				kv, err := this.decodeKv(opt, ev.Kv.Key, ev.Kv.Value)
				if err != nil {
					continue
				}
				this.kvChan <- client.KvEvent{client.OptPut, *kv}
			} else if ev.Type.String() == client.OptDelete {
				var kv *client.Kv
				var err error
				if ev.PrevKv != nil {
					kv, err = this.decodeKv(opt, ev.Kv.Key, ev.PrevKv.Value)
					if err != nil {
						continue
					}
				} else {
					kv = new(client.Kv)
					kv.Key = string(ev.Kv.Key)
				}
				this.kvChan <- client.KvEvent{client.OptDelete, *kv}
			} else {
				logkit.Errorf("wrong type, %s %q : %q %d\n", ev.Type, ev.Kv.Key, ev.Kv.Value, ev.Kv.Version)
			}
		}
	}

	this.kvMux.Lock()
	defer this.kvMux.Unlock()
	if this.kvWatchClosed {
		logkit.Infof("key:%s watch exit!", key)
		return
	}

	ctxCancel, ok := this.kvContext[key]
	if !ok {
		return
	}

	ctxCancel.cancel()
	ctxCancel.tm.Stop()
	version := ctxCancel.version
	delete(this.kvContext, key)

	logkit.Infof("key:%s watch restart!", key)

	ctx, cancel := context.WithCancel(context.Background())
	//ctx, cancel := context.WithTimeout(context.Background(), watchTimeout)
	this.kvContext[key] = &ContextCancel{ctx, cancel, version, time.NewTimer(notifyInterval)}
	go func(ct *ContextCancel) {
		select {
		case <-ct.ctx.Done():
			return
		case <-ct.tm.C:
			ct.cancel()
		}
	}(this.kvContext[key])

	var newRespCh clientv3.WatchChan
	if version == 0 {
		newRespCh = this.kvWatchClient.Watch(ctx, key, clientv3.WithCreatedNotify(), clientv3.WithProgressNotify(), clientv3.WithPrevKV())
	} else {
		newRespCh = this.kvWatchClient.Watch(ctx, key, clientv3.WithCreatedNotify(), clientv3.WithProgressNotify(), clientv3.WithRev(version), clientv3.WithPrevKV())
	}

	go this.kvWatch(key, newRespCh, opt)

	return
}

func (this *EtcdClient) KvCloseWatch() error {
	this.kvMux.Lock()
	defer this.kvMux.Unlock()

	if this.kvWatchClosed {
		return nil
	} else {
		this.kvWatchClosed = true
	}

	for key, ctxCancel := range this.kvContext {
		ctxCancel.cancel()
		ctxCancel.tm.Stop()
		delete(this.kvContext, key)
		logkit.Infof("close key %s watch", key)
	}

	err := this.kvWatchClient.Close()
	if err == context.Canceled {
		err = nil
	}

	close(this.kvChan)

	return err
}

func (this *EtcdClient) decodeKv(opt *client.KvOpt, key, value []byte) (*client.Kv, error) {
	kv := new(client.Kv)
	kv.Key = string(key)
	var valueByte []byte

	if opt.Crypt && this.crypter != nil {
		if len(value) != 0 {
			base64Result, err := base64.StdEncoding.DecodeString(string(value))
			if err != nil {
				logkit.Errorf("base64 decode pwd %s error:%s", value, err)
				return nil, err
			}
			valueByte, err = this.crypter.Decrypt(base64Result)
			if err != nil {
				logkit.Errorf("decode kv error, %s %s", string(key), base64Result)
				return nil, err
			}
		}
	} else {
		valueByte = value
	}

	kv.Value = string(valueByte)

	return kv, nil
}

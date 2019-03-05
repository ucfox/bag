package etcd

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"git.pandatv.com/panda-public/janna/client"
	"git.pandatv.com/panda-web/gobase/log"
	"github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
)

func (this *EtcdClient) ServiceRegister(serv *client.Service) error {
	_, _, _, err := client.SplitServiceKey(serv.Key)
	if err != nil {
		return err
	}

	if this.crypter != nil {
		cryptPwd, err := this.crypter.Encrypt([]byte(serv.Password))
		if err != nil {
			return err
		}
		serv.Password = base64.StdEncoding.EncodeToString(cryptPwd)
	}

	servByte, err := json.Marshal(serv.ServiceValue)
	if err != nil {
		logkit.Errorf("marshal service %+v error:%s", serv, err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	_, err = this.kvClient.Put(ctx, serv.Key, string(servByte))
	cancel()
	if err != nil {
		logkit.Errorf("put service %+v error:%s", serv, err)
		return checkEtcdError(err)
	}

	return nil
}

func (this *EtcdClient) ServiceDeregister(key string) error {
	_, _, _, err := client.SplitServiceKey(key)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	_, err = this.kvClient.Delete(ctx, key)
	cancel()
	if err != nil {
		logkit.Errorf("delete service %s error:%s", key, err)
		return checkEtcdError(err)
	}

	return nil
}

func (this *EtcdClient) ServiceGet(key string) (*client.Service, error) {
	_, _, _, err := client.SplitServiceKey(key)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	resp, err := this.kvClient.Get(ctx, key)
	cancel()
	if err != nil {
		logkit.Errorf("get service %s error:%s", key, err)
		return nil, checkEtcdError(err)
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	serv, err := this.decodeService(resp.Kvs[0].Key, resp.Kvs[0].Value)
	if err != nil {
		return nil, err
	}

	return serv, nil
}

func (this *EtcdClient) ServiceGetAll(key string) ([]client.Service, error) {
	//_, _, err := client.SplitServicePrefixKey(key)
	//if err != nil {
	//	return nil, err
	//}

	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	resp, err := this.kvClient.Get(ctx, key, clientv3.WithPrefix())
	cancel()
	if err != nil {
		logkit.Errorf("get services %s error:%s", key, err)
		return nil, checkEtcdError(err)
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	servList := make([]client.Service, 0)
	for _, kv := range resp.Kvs {
		serv, err := this.decodeService(kv.Key, kv.Value)
		if err != nil {
			return nil, err
		}
		servList = append(servList, *serv)
	}

	return servList, nil
}

func (this *EtcdClient) ServiceAddWatch(key string) error {
	//_, _, err := client.SplitServicePrefixKey(key)
	//if err != nil {
	//	return err
	//}

	this.serviceMux.Lock()
	defer this.serviceMux.Unlock()

	if this.serviceWatchClosed {
		return ErrServiceWatchClosed
	}

	_, ok := this.serviceContext[key]
	if ok {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	//ctx, cancel := context.WithTimeout(context.Background(), watchTimeout)
	this.serviceContext[key] = &ContextCancel{ctx, cancel, 0, time.NewTimer(notifyInterval)}
	go func(ct *ContextCancel) {
		select {
		case <-ct.ctx.Done():
			return
		case <-ct.tm.C:
			ct.cancel()
		}
	}(this.serviceContext[key])

	respCh := this.serviceWatchClient.Watch(ctx, key, clientv3.WithPrefix(), clientv3.WithCreatedNotify(), clientv3.WithProgressNotify(), clientv3.WithPrevKV())

	go this.serviceWatch(key, respCh)

	return nil
}

func (this *EtcdClient) ServiceRemoveWatch(key string) error {
	//_, _, err := client.SplitServicePrefixKey(key)
	//if err != nil {
	//	return err
	//}

	this.serviceMux.Lock()
	defer this.serviceMux.Unlock()

	ctxCancel, ok := this.serviceContext[key]
	if !ok {
		return nil
	}

	ctxCancel.cancel()
	ctxCancel.tm.Stop()
	delete(this.serviceContext, key)

	return nil
}

func (this *EtcdClient) ServiceGetWatch() (chan client.ServiceEvent, error) {
	return this.serviceChan, nil
}

func (this *EtcdClient) serviceWatch(key string, respCh clientv3.WatchChan) {
	for resp := range respCh {
		if resp.Err() != nil {
			if !resp.Canceled {
				logkit.Errorf("watch service %s response error:%s", key, resp.Err())
			}
			time.Sleep(time.Millisecond * 100)
			break
		}

		this.serviceMux.Lock()
		ctxCancel, ok := this.serviceContext[key]
		if ok {
			ctxCancel.version = resp.Header.Revision
			ctxCancel.tm.Reset(notifyInterval)
		}
		this.serviceMux.Unlock()

		for _, ev := range resp.Events {
			if ev.Type.String() == client.OptPut {
				serv, err := this.decodeService(ev.Kv.Key, ev.Kv.Value)
				if err != nil {
					continue
				}
				this.serviceChan <- client.ServiceEvent{client.OptPut, *serv}
			} else if ev.Type.String() == client.OptDelete {
				var serv *client.Service
				var err error
				if ev.PrevKv != nil {
					serv, err = this.decodeService(ev.Kv.Key, ev.PrevKv.Value)
					if err != nil {
						continue
					}
				} else {
					serv = new(client.Service)
					serv.Key = string(ev.Kv.Key)
				}
				this.serviceChan <- client.ServiceEvent{client.OptDelete, *serv}
			} else {
				logkit.Errorf("wrong type, %s %q : %q %d\n", ev.Type, ev.Kv.Key, ev.Kv.Value, ev.Kv.Version)
			}
		}
	}

	this.serviceMux.Lock()
	defer this.serviceMux.Unlock()
	if this.serviceWatchClosed {
		logkit.Infof("service %s watch exit!", key)
		return
	}

	ctxCancel, ok := this.serviceContext[key]
	if !ok {
		return
	}
	version := ctxCancel.version
	ctxCancel.cancel()
	ctxCancel.tm.Stop()
	delete(this.serviceContext, key)

	logkit.Infof("service %s watch restart!", key)

	ctx, cancel := context.WithCancel(context.Background())
	//ctx, cancel := context.WithTimeout(context.Background(), watchTimeout)
	this.serviceContext[key] = &ContextCancel{ctx, cancel, version, time.NewTimer(notifyInterval)}
	go func(ct *ContextCancel) {
		select {
		case <-ct.ctx.Done():
			return
		case <-ct.tm.C:
			ct.cancel()
		}
	}(this.serviceContext[key])

	var newRespCh clientv3.WatchChan
	if version == 0 {
		newRespCh = this.serviceWatchClient.Watch(ctx, key, clientv3.WithPrefix(), clientv3.WithCreatedNotify(), clientv3.WithProgressNotify(), clientv3.WithPrevKV())
	} else {
		newRespCh = this.serviceWatchClient.Watch(ctx, key, clientv3.WithPrefix(), clientv3.WithCreatedNotify(), clientv3.WithProgressNotify(), clientv3.WithRev(version), clientv3.WithPrevKV())
	}
	go this.serviceWatch(key, newRespCh)

	return
}

func (this *EtcdClient) ServiceCloseWatch() error {
	this.serviceMux.Lock()
	defer this.serviceMux.Unlock()

	if this.serviceWatchClosed {
		return nil
	} else {
		this.serviceWatchClosed = true
	}

	for key, ctxCancel := range this.serviceContext {
		ctxCancel.cancel()
		ctxCancel.tm.Stop()
		delete(this.serviceContext, key)
		logkit.Infof("close service %s watch", key)
	}

	err := this.serviceWatchClient.Close()
	if err == context.Canceled {
		err = nil
	}

	close(this.serviceChan)

	return err
}

func (this *EtcdClient) decodeService(key, value []byte) (*client.Service, error) {
	serv := new(client.Service)
	var servVal client.ServiceValue

	err := json.Unmarshal(value, &servVal)
	if err != nil {
		logkit.Errorf("unmarshal service %s %s error:%s", string(key), string(value), err)
		return nil, err
	}

	if this.crypter != nil && len(servVal.Password) != 0 {
		base64Result, err := base64.StdEncoding.DecodeString(servVal.Password)
		if err != nil {
			logkit.Errorf("base64 decode pwd %s error:%s", servVal.Password, err)
			return nil, err
		}
		decryptPwd, err := this.crypter.Decrypt(base64Result)
		if err != nil {
			logkit.Errorf("unmarshal service %s %s error:%s", string(key), string(value), err)
			return nil, err
		}
		servVal.Password = string(decryptPwd)
	}

	serv.Key = string(key)
	serv.ServiceValue = servVal

	return serv, nil
}

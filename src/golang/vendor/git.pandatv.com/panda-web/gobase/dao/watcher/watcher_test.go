package watcher

import (
	"fmt"
	"testing"
	"time"

	"git.pandatv.com/panda-public/janna/etcd"
)

func TestGetAllInstance(t *testing.T) {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{"etcd.demo.pdtv.io:2379"},
		DialTimeout: etcd.ConnectTimeout,
		Username:    "janna",
		Password:    "RDX19t0dNBoAsaTO",
	})

	if err != nil {
		fmt.Printf("init etcd error:%s\n", err)
		return
	}

	w, err := NewWatcher(etcdClient)
	if err != nil {
		fmt.Printf("init watcher error:%s\n", err)
		return
	}

	servList, err := w.GetAllInstance("test", "mysql", "")
	if err != nil {
		fmt.Printf("get all instance error:%s\n", err)
		return
	}

	fmt.Printf("%+v\n", servList)

	w.Close()
}

func TestWatchInstance(t *testing.T) {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{"etcd.demo.pdtv.io:2379"},
		DialTimeout: etcd.ConnectTimeout,
		Username:    "janna",
		Password:    "RDX19t0dNBoAsaTO",
	})

	if err != nil {
		fmt.Printf("init etcd error:%s\n", err)
		return
	}

	w, err := NewWatcher(etcdClient)
	if err != nil {
		fmt.Printf("init watcher error:%s\n", err)
		return
	}

	ch, err := w.WatchInstance("test", "mysql", "")
	if err != nil {
		fmt.Printf("get all instance error:%s\n", err)
		return
	}

	fmt.Println("wait event")
	select {
	case <-time.After(3 * time.Second):
	case servList := <-ch:
		fmt.Printf("%+v\n", servList)
	}

	w.Close()
}

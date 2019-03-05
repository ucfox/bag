package etcd

import (
	"fmt"
	"testing"

	"git.pandatv.com/panda-public/janna/client"

	"github.com/stretchr/testify/assert"
)

func getKvCli() *EtcdClient {
	cli, err := NewWithDes(Config{
		Endpoints:   []string{"etcd.demo.pdtv.io:2379"},
		DialTimeout: ConnectTimeout,
	},
		[]byte("hahahehe"),
	)
	var t *testing.T
	assert.Nil(t, err, "new client failed")
	return cli
}

func TestKvCloseWatch(t *testing.T) {
	cli := getKvCli()
	defer cli.Close()
	err := cli.KvCloseWatch()
	assert.Nil(t, err, "close watch kv failed")
}

func TestKvGetWatch(t *testing.T) {
	cli := getKvCli()
	defer cli.Close()
	_, err := cli.KvGetWatch()
	assert.Nil(t, err, "get watch kv failed")

}

func TestKvAddWatch(t *testing.T) {
	cli := getKvCli()
	defer cli.KvCloseWatch()
	defer cli.Close()
	err := cli.KvAddWatch(client.KvKey("test", "key1"))
	assert.Nil(t, err, "add watch kv failed")
}

func TestKvPut(t *testing.T) {
	cli := getKvCli()
	defer cli.Close()
	err := cli.KvPut(client.KvKey("test", "key1"), "value1")
	assert.Nil(t, err, "Kv put failed")
	err = cli.KvPut(client.KvKey("test", "key2"), "value1", client.WithCrypt())
	assert.Nil(t, err, "Kv put failed")
}

func TestKvRemoveWatch(t *testing.T) {
	cli := getKvCli()
	defer cli.Close()
	err := cli.KvRemoveWatch(client.KvKey("test", "key1"))
	assert.Nil(t, err, "remove watch kv failed")
}

func TestKvGet(t *testing.T) {
	cli := getKvCli()
	defer cli.Close()
	kvResult, err := cli.KvGet(client.KvKey("test", "key1"))
	assert.Nil(t, err, "Kv get failed")
	assert.Equal(t, kvResult, "value1", "Kv get err result")
	kvResult, err = cli.KvGet(client.KvKey("test", "key2"), client.WithCrypt())
	assert.Nil(t, err, "Kv get failed")
	assert.Equal(t, kvResult, "value1", "Kv get err result")
}

func TestKvGetAll(t *testing.T) {
	cli := getKvCli()
	defer cli.Close()
	kvResults, err := cli.KvGetAll(client.KvPrefixKey("test"))
	assert.Nil(t, err, "Kv getall failed")
	fmt.Println(kvResults)
}

func TestKvDelete(t *testing.T) {
	cli := getKvCli()
	defer cli.Close()
	err := cli.KvDelete(client.KvKey("test", "key1"))
	assert.Nil(t, err, "Kv delete failed")
}

//同时调用多个接口，检测KvGetWatch返回的channel
func TestKvWatch(t *testing.T) {
	cli := getKvCli()
	defer cli.Close()
	defer cli.KvCloseWatch()
	cli.KvAddWatch(client.KvKey("test", "key1"))
	cli.KvPut(client.KvKey("test", "key1"), "value1")

	cli.KvAddWatch(client.KvKey("test2", "key2"), client.WithCrypt())
	cli.KvPut(client.KvKey("test2", "key2"), "value2", client.WithCrypt())
	ch, err := cli.KvGetWatch()
	assert.Nil(t, err, "get watch kv failed")

	kvevent_1 := <-ch
	if kvevent_1.Key != "/kv/test/key1" || kvevent_1.Opt != "PUT" || kvevent_1.Value != "value1" {
		t.Error(t, "kv watch have err")
	}
	kvevent_2 := <-ch
	if kvevent_2.Key != "/kv/test2/key2" || kvevent_2.Opt != "PUT" || kvevent_2.Value != "value2" {
		t.Error(t, "kv watch have err")
	}

}

func TestDecodeKv(t *testing.T) {
	cli := getKvCli()
	opt := new(client.KvOpt)
	opt.Crypt = true
	kv, err := cli.decodeKv(opt, []byte("test"), nil)
	assert.Nil(t, err, "decode kv failed")
	fmt.Println(*kv)
}

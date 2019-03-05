package etcd

import (
	"testing"

	"git.pandatv.com/panda-public/janna/client"
	"github.com/stretchr/testify/assert"
)

func getservCli() *EtcdClient {
	cli, err := NewWithDes(Config{
		Endpoints:   []string{"etcd.demo.pdtv.io:2379"},
		DialTimeout: ConnectTimeout,
	},
		[]byte("key12345"),
	)
	var t *testing.T
	assert.Nil(t, err, "new client failed")
	return cli
}

func getserv1() *client.Service {
	serv := new(client.Service)
	serv.Key = client.ServiceKey("test", "mongo", "default_slave_1")
	serv.Tag = []string{"default", "slave"}
	serv.Address = "10.20.1.20"
	serv.Port = 7300
	serv.Weight = 100
	serv.Password = "dev123"
	return serv
}

func getserv2() *client.Service {
	serv := new(client.Service)
	serv.Key = client.ServiceKey("test", "mongo", "default_master_1")
	serv.Tag = []string{"default", "master"}
	serv.Address = "10.20.1.20"
	serv.Port = 7300
	serv.Weight = 0
	serv.User = "dev"
	serv.Password = "dev123"
	return serv
}

func TestServiceCloseWatch(t *testing.T) {
	cli := getservCli()
	defer cli.Close()
	err := cli.ServiceCloseWatch()
	assert.Nil(t, err, "close watch service failed")
}

func TestServiceGetWatch(t *testing.T) {
	cli := getservCli()
	defer cli.Close()
	_, err := cli.ServiceGetWatch()
	assert.Nil(t, err, "get service watch failed")
}

func TestServiceAddWatch(t *testing.T) {
	cli := getservCli()
	defer cli.Close()
	defer cli.ServiceCloseWatch()
	err := cli.ServiceAddWatch(client.ServicePrefixKey("test", "mongo"))
	assert.Nil(t, err, "add watch service failed")
}

func TestServiceRegister(t *testing.T) {
	cli := getservCli()
	defer cli.Close()
	serv := getserv1()
	err := cli.ServiceRegister(serv)
	assert.Nil(t, err, "register service failed")
	serv = getserv2()
	err = cli.ServiceRegister(serv)
	assert.Nil(t, err, "register service failed")
}

func TestServiceRemoveWatch(t *testing.T) {
	cli := getservCli()
	defer cli.Close()
	err := cli.ServiceRemoveWatch(client.ServicePrefixKey("test", "mongo"))
	assert.Nil(t, err, "remove watch service failed")
}

func TestServiceGet(t *testing.T) {
	cli := getservCli()
	defer cli.Close()
	serv := getserv1()
	servResult, err := cli.ServiceGet(serv.Key)
	assert.Nil(t, err, "get service failed")
	assert.Equal(t, serv, servResult, "func ServiceGet  return err result")
	serv = getserv2()
	servResult, err = cli.ServiceGet(serv.Key)
	assert.Nil(t, err, "get service failed")
	assert.Equal(t, serv, servResult, "func ServiceGet  return err result")
}

func TestServiceGetAll(t *testing.T) {
	cli := getservCli()
	defer cli.Close()
	serv1 := getserv1()
	serv2 := getserv2()
	servListResult, err := cli.ServiceGetAll(client.ServicePrefixKey("test", "mongo"))
	assert.Nil(t, err, "get all service failed")
	assert.Equal(t, serv1, &servListResult[1], "func ServiceGetAll  return err result")
	assert.Equal(t, serv2, &servListResult[0], "func ServiceGetAll  return err result")
}

func TestServiceDeregister(t *testing.T) {
	cli := getservCli()
	defer cli.Close()
	serv := getserv1()
	err := cli.ServiceDeregister(serv.Key)
	assert.Nil(t, err, "deregister service failed")
}

//同时调用多个接口，检测ServiceGetWatch返回的channel
func TestServiceWatch(t *testing.T) {
	serv1 := getserv1()
	serv2 := getserv2()
	cli := getservCli()
	defer cli.Close()
	defer cli.ServiceCloseWatch()
	cli.ServiceAddWatch(client.ServicePrefixKey("test", "mongo"))
	cli.ServiceRegister(serv1)
	cli.ServiceRegister(serv2)
	ch, err := cli.ServiceGetWatch()
	assert.Nil(t, err, "get service watch failed")
	servevent1 := <-ch
	if servevent1.Address != serv1.Address || servevent1.Key != serv1.Key || servevent1.Port != serv1.Port || servevent1.Weight != serv1.Weight {
		t.Error(t, "service watch have err")
	}
	servevent2 := <-ch
	if servevent2.Address != serv2.Address || servevent2.Key != serv2.Key || servevent2.Port != serv2.Port || servevent2.Weight != serv2.Weight {
		t.Error(t, "service watch have err")
	}
}

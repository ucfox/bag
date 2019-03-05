package etcd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cli, err := New(Config{
		Endpoints:   []string{"etcd.demo.pdtv.io:2379"},
		DialTimeout: ConnectTimeout,
	})
	assert.Nil(t, err, "new client failed")
	assert.Nil(t, cli.crypter, "new client failed")

	err = cli.Close()
	assert.Nil(t, err, "close client failed")
}

func TestNewWithDes(t *testing.T) {
	cli, err := NewWithDes(Config{
		Endpoints:   []string{"etcd.demo.pdtv.io:2379"},
		DialTimeout: ConnectTimeout,
	},
		[]byte("hahahehe"),
	)
	assert.Nil(t, err, "new client failed")
	assert.NotNil(t, cli.crypter, "new client failed")

	err = cli.Close()
	assert.Nil(t, err, "close client failed")
}

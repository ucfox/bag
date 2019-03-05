package client

import (
	"errors"
)

var ErrFormatKey = errors.New("key is wrong format")

const (
	OptPut    = "PUT"
	OptDelete = "DELETE"
)

type Client interface {
	ServiceClient
	KvClient
	Close() error
}

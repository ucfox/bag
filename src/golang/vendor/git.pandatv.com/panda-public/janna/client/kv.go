package client

import (
	"fmt"
	"strings"
)

//   /kv/grank/
//   /kv/grank/foo
const (
	kvPrefixKeyFormat = "/kv/%s/"
	kvKeyFormat       = "/kv/%s/%s"
)

type KvOpt struct {
	Crypt bool
}

type KvOption func(*KvOpt)

func WithCrypt() KvOption {
	return func(ko *KvOpt) {
		ko.Crypt = true
	}
}

type Kv struct {
	Key   string
	Value string
}

type KvEvent struct {
	Opt string
	Kv
}

type KvClient interface {
	KvPut(string, string, ...KvOption) error
	KvGet(string, ...KvOption) (string, error)
	KvGetAll(string, ...KvOption) ([]Kv, error)
	KvDelete(string) error
	KvAddWatch(string, ...KvOption) error
	KvRemoveWatch(string) error
	KvGetWatch() (chan KvEvent, error)
	KvCloseWatch() error
}

//callName is the name of your service
//subKey is the key you can self define
func KvKey(callName, subKey string) string {
	return fmt.Sprintf(kvKeyFormat, callName, subKey)
}

func SplitKvKey(key string) (string, string, error) {
	keySlice := strings.SplitN(key, "/", 4)
	if len(keySlice) != 4 {
		return "", "", ErrFormatKey
	}
	if keySlice[1] != "kv" {
		return "", "", ErrFormatKey
	}

	return keySlice[2], keySlice[3], nil
}

func KvPrefixKey(callName string) string {
	return fmt.Sprintf(kvPrefixKeyFormat, callName)
}

func SplitKvPrefixKey(key string) (string, error) {
	keySlice := strings.SplitN(key, "/", 4)
	if len(keySlice) != 4 {
		return "", ErrFormatKey
	}
	if keySlice[1] != "kv" {
		return "", ErrFormatKey
	}

	return keySlice[2], nil
}

package redis

import (
	"log"
	"testing"
)

func TestNew(t *testing.T) {
	cliPool := NewRedisClient("etcd.demo.pdtv.io:2379", "test", "userPwd", "", "")
	if cliPool == nil {
		log.Panicln("empty redis")
	}

	_, err := cliPool.Del("foo")
	if err != nil {
		log.Panicln(err.Error())
	}

	_, err = cliPool.Set("foo", "bar", 0)
	if err != nil {
		log.Panicln(err.Error())
	}

	val, err := cliPool.Get("foo")
	if err != nil {
		log.Panicln(err.Error())
	}
	if val != "bar" {
		log.Panicln("value %s not equal bar", val)
	}
}

func TestNewEnc(t *testing.T) {
	cliPool := NewRedisClientEnc("etcd.demo.pdtv.io:2379", "test", "userPwd", "key12345")
	if cliPool == nil {
		log.Panicln("empty redis")
	}

	_, err := cliPool.Del("foo")
	if err != nil {
		log.Panicln(err.Error())
	}

	_, err = cliPool.Set("foo", "bar", 0)
	if err != nil {
		log.Panicln(err.Error())
	}

	val, err := cliPool.Get("foo")
	if err != nil {
		log.Panicln(err.Error())
	}
	if val != "bar" {
		log.Panicln("value %s not equal bar", val)
	}
}

type WrapRedis struct {
	*RedisBaseDao
}

func TestWrap(t *testing.T) {
	cliPool := NewRedisClient("etcd.demo.pdtv.io:2379", "test", "userPwd", "", "")
	if cliPool == nil {
		log.Panicln("empty redis")
	}

	cliPoolWrap := &WrapRedis{cliPool}

	_, err := cliPoolWrap.Del("foo")
	if err != nil {
		log.Panicln(err.Error())
	}

	_, err = cliPoolWrap.Set("foo", "bar", 0)
	if err != nil {
		log.Panicln(err.Error())
	}

	val, err := cliPoolWrap.Get("foo")
	if err != nil {
		log.Panicln(err.Error())
	}
	if val != "bar" {
		log.Panicln("value %s not equal bar", val)
	}
}

func TestWrapEnc(t *testing.T) {
	cliPool := NewRedisClientEnc("etcd.demo.pdtv.io:2379", "test", "userPwd", "key12345")
	if cliPool == nil {
		log.Panicln("empty redis")
	}

	cliPoolWrap := &WrapRedis{cliPool}

	_, err := cliPoolWrap.Del("foo")
	if err != nil {
		log.Panicln(err.Error())
	}

	_, err = cliPoolWrap.Set("foo", "bar", 0)
	if err != nil {
		log.Panicln(err.Error())
	}

	val, err := cliPoolWrap.Get("foo")
	if err != nil {
		log.Panicln(err.Error())
	}
	if val != "bar" {
		log.Panicln("value %s not equal bar", val)
	}
}

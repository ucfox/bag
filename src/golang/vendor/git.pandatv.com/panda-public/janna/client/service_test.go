package client

import (
	"fmt"
	"testing"
)

func TestSplitServiceKey(t *testing.T) {
	callName, serviceName, serviceKey, err := SplitServiceKey("/service/test/redis/1")
	if err != nil {
		panic(err)
	}
	fmt.Println(callName, serviceName, serviceKey)

	callName, serviceName, serviceKey, err = SplitServiceKey("/service/test/redis")
	if err == nil {
		panic("wrong split service key case")
	}
	fmt.Println(callName, serviceName, serviceKey)

	callName, serviceName, serviceKey, err = SplitServiceKey("/service/test/redis/1/slave")
	if err != nil {
		panic(err)
	}
	fmt.Println(callName, serviceName, serviceKey)
}

func TestSplitServicePrefixKey(t *testing.T) {
	callName, serviceName, err := SplitServicePrefixKey("/service/test/redis/")
	if err != nil {
		panic(err)
	}
	fmt.Println(callName, serviceName)

	callName, serviceName, err = SplitServicePrefixKey("/service/test/")
	if err == nil {
		panic("wrong split service prefix key case")
	}
	fmt.Println(callName, serviceName)

	callName, serviceName, err = SplitServicePrefixKey("/service/test/redis/1/")
	if err != nil {
		panic(err)
	}
	fmt.Println(callName, serviceName)
}

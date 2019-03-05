package client

import (
	"fmt"
	"testing"
)

func TestSplitKvKey(t *testing.T) {
	callName, subKey, err := SplitKvKey("/kv/test/foo")
	if err != nil {
		panic(err)
	}
	fmt.Println(callName, subKey)

	callName, subKey, err = SplitKvKey("/kv/test")
	if err == nil {
		panic("wrong split kv key case")
	}
	fmt.Println(callName, subKey)

	callName, subKey, err = SplitKvKey("/kv/test/foo/bar")
	if err != nil {
		panic(err)
	}
	fmt.Println(callName, subKey)
}

func TestSplitKvPrefixKey(t *testing.T) {
	callName, err := SplitKvPrefixKey("/kv/test/")
	if err != nil {
		panic(err)
	}
	fmt.Println(callName)

	callName, err = SplitKvPrefixKey("/kv/")
	if err == nil {
		panic("wrong split kv prefix key case")
	}
	fmt.Println(callName)

	callName, err = SplitKvPrefixKey("/kv/test/foo/")
	if err != nil {
		panic(err)
	}
	fmt.Println(callName)
}

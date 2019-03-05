package utils

import (
	"os"
	"testing"
)

func TestUtils(t *testing.T) {
	var errno int = 5
	var errmsg string = "test"
	var data string = "test"
	res := RespJson(errno, errmsg, data)
	os.Stdout.Write(res)

	resjsonp := RespFormat(errno, errmsg, data, "callback")
	os.Stdout.Write(resjsonp)
	resjsonp1 := RespFormat(errno, errmsg, data, "")
	os.Stdout.Write(resjsonp1)
}

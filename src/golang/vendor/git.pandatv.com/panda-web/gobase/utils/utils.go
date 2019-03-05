package utils

import (
	"encoding/json"
)

type Result struct {
	Errno  int         `json:"errno"`
	Errmsg string      `json:"errmsg"`
	Data   interface{} `json:"data"`
}

func RespJson(errno int, errmsg string, data interface{}) []byte {
	var result = new(Result)
	result.Errno = errno
	result.Errmsg = errmsg
	result.Data = data
	res, _ := json.Marshal(result)
	return res
}

func RespFormat(errno int, errmsg string, data interface{}, callback string) []byte {
	res := RespJson(errno, errmsg, data)
	if callback != "" {
		res = []byte(callback + "(" + string(res) + ")")
	}
	return res
}

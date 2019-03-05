package common

const (
	SUCCESS        = 0
	SYSTEM_ERROR   = 1001
	INVALID_PARAMS = 7000
)

var RespMsg = map[int]string{
	0:    "success",
	1001: "system error",
	7000: "invalid params",
}

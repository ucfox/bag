package pconfig

import (
	"os"
	"testing"
)

func Test_config(t *testing.T) {
	file, _ := os.Getwd() //获取路径
	dir := file           //内部会查找golang.env文件, 用来做测试

	Init(dir)
	ReadLine(configPath)
	if conf["DB_USER"] != "dev" {
		t.Error(conf["DB_USER"])
	}
}

package main

import (
	// "fmt"
	"git.pandatv.com/panda-public/janna_dao/etcd/mysql"
	"git.pandatv.com/panda-web/gobase/log"
	"os"
	"strings"
	"testing"
)

func InitTestCase() {
	logkit.Logger.Init("bag_consume_packet", logkit.LevelInfo)

	MysqlClient, _ = mysql.NewMysqlBaseDao("etcd.demo.pdtv.io:2379", "ibag", "default", "pdev", "jQXtkYd2drQWj43c", "ibag", "p_dev", "r2evKK0QJj1oHQe1", "ibag")
	os.Setenv("FIRST_CHARGE_GOODS", "1,2,4")
	os.Setenv("FIRST_CHARGE_NUM", "10,50,1")

	os.Setenv("GUUID_IDC", "test")
	os.Setenv("GUUID_ENV", "dev")

	FirstChargeGoods = strings.Split(os.Getenv("FIRST_CHARGE_GOODS"), ",")
	FirstChargeGoodsNum = strings.Split(os.Getenv("FIRST_CHARGE_NUM"), ",")

}

func Test_Execute(t *testing.T) {
	InitTestCase()
	// test数据
	str := `{"rid":27400276,"packlimit":20}`
	data := []byte(str)
	execute(data)
}

func Test_SendMsg(t *testing.T) {
	uid := "27400276"
	FirstChargeGoodsNum = strings.Split(os.Getenv("FIRST_CHARGE_NUM"), ",")
	sendMsg(uid)
}

func Test_QaTest(t *testing.T) {
	InitTestCase()
	// test数据
	str := `{"rid":274002761,"packlimit":20}`
	data := []byte(str)
	execute(data)
}

func Test_genGuuid(t *testing.T) {
	InitTestCase()
	// test数据
	uid := "27400276"
	id := genGuuid(uid)
	t.Log(id)
}

func Test_genSelfGuuid(t *testing.T) {
	InitTestCase()
	// test数据
	uid := "27400276"
	id := genSelfGuuid(uid)
	t.Log(id)
	id = genSelfGuuid(uid)
	t.Log(id)
}

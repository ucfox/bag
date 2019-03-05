package mysql

import (
	"testing"
	"time"

	"fmt"

	janna "git.pandatv.com/panda-public/janna/client"
)

func Test_Insert(t *testing.T) {
	MysqlClient, err := NewMysqlBaseDao("p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "gobase_mysql_demo", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, "p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "gobase_mysql_demo", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, DEFAULT_MASTER_SLAVE)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 5; i++ {
		sql := "Insert INTO `test` (`ver`, `createtime`, `updatetime`, `name`) VALUES (?, ?, ?, ?)"
		now := time.Now().Format(time.RFC3339)
		var args = make([]interface{}, 0)
		args = append(args, 1)
		args = append(args, now)
		args = append(args, now)
		args = append(args, "haha")
		id, err := MysqlClient.Insert(sql, args...)
		if err != nil {
			t.Error(err)
		}
		t.Log(id)
	}
}

func Test_FetchRow(t *testing.T) {
	MysqlClient, err := NewMysqlBaseDao("p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "gobase_mysql_demo", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, "p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "gobase_mysql_demo", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, DEFAULT_MASTER_SLAVE)
	if err != nil {
		t.Error(err)
	}

	var args = make([]interface{}, 0)
	args = append(args, "haha")
	res, err := MysqlClient.FetchRow("SELECT * FROM `test` WHERE `name` = ? ORDER BY `id` DESC LIMIT 1", args...)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
}

func Test_FetchRows(t *testing.T) {
	MysqlClient, err := NewMysqlBaseDao("p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "gobase_mysql_demo", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, "p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "gobase_mysql_demo", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, DEFAULT_MASTER_SLAVE)
	if err != nil {
		t.Error(err)
	}

	var args = make([]interface{}, 0)
	args = append(args, "haha")
	res, err := MysqlClient.FetchRows("SELECT * FROM `test` WHERE `name` = ? ORDER BY id DESC LIMIT 5", args...)
	if err != nil {
		t.Error(err)
	}
	for _, v := range res {
		fmt.Println(v)
	}
}

func Test_Update(t *testing.T) {
	MysqlClient, err := NewMysqlBaseDao("p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "gobase_mysql_demo", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, "p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "gobase_mysql_demo", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, DEFAULT_MASTER_SLAVE)
	if err != nil {
		t.Error(err)
	}

	var args = make([]interface{}, 0)
	now := time.Now().Format(time.RFC3339)
	args = append(args, now)
	args = append(args, "lala")
	args = append(args, "haha")
	res, err := MysqlClient.Update("UPDATE `test` SET `ver` = ver + 1, `updatetime` = ?, `name` = ? WHERE `name` = ?", args...)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
}

func Test_Delete(t *testing.T) {
	MysqlClient, err := NewMysqlBaseDao("p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "gobase_mysql_demo", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, "p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "gobase_mysql_demo", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, DEFAULT_MASTER_SLAVE)
	if err != nil {
		t.Error(err)
	}

	var args = make([]interface{}, 0)
	args = append(args, "lala")
	res, err := MysqlClient.Delete("DELETE FROM `test` WHERE `name` = ?", args...)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
}

func Test_masterSlave(t *testing.T) {
	MysqlClient, err := NewMysqlBaseDao("p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "gobase_mysql_demo", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, "p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "gobase_mysql_demo", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, false)
	if err != nil {
		t.Error(err)
	}

	var args = make([]interface{}, 0)
	args = append(args, "haha")
	res, err := MysqlClient.FetchRow("SELECT * FROM `test` WHERE `name` = ? ORDER BY `id` LIMIT 1", args...)
	if err != nil {
		t.Error(err)
	}
	t.Log(res)
}

func Test_diff(t *testing.T) {
	MysqlClient, err := NewMysqlBaseDao("p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "gobase_mysql_demo", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, "p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "gobase_mysql_demo", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, false)
	if err != nil {
		t.Error(err)
	}

	var args = make([]interface{}, 0)
	args = append(args, "haha")
	res, err := MysqlClient.FetchRow("SELECT * FROM `test` WHERE `name` = ? ORDER BY `id` LIMIT 1", args...)
	if err != nil {
		t.Error(err)
	}
	t.Log(res)
	//新建一个buffon数据库的连接
	BuffonMysqlClient, err := NewMysqlBaseDao("p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "buffon_zhaorui", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, "p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "buffon_zhaorui", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, false)
	if err != nil {
		t.Error(err)
	}

	args = make([]interface{}, 0)
	ret, err := BuffonMysqlClient.FetchRows("SELECT * FROM `project`", args...)
	if err != nil {
		t.Error(err)
	}
	t.Log(ret)
}

func Test_FetchRowForMaster(t *testing.T) {
	MysqlClient, err := NewMysqlBaseDao("p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "gobase_mysql_demo", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, "p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "gobase_mysql_demo", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, DEFAULT_MASTER_SLAVE)
	if err != nil {
		t.Error(err)
	}

	var args = make([]interface{}, 0)
	args = append(args, "haha")
	res, err := MysqlClient.FetchRowForMaster("SELECT * FROM `test` WHERE `name` = ? ORDER BY `id` LIMIT 1", args...)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
}

func Test_FetchRowsForMaster(t *testing.T) {
	MysqlClient, err := NewMysqlBaseDao("p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "gobase_mysql_demo", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, "p_dev", "r2evKK0QJj1oHQe1", "10.20.1.11", "3306", "gobase_mysql_demo", DEFAULT_MAX_OPEN_CONNS, DEFAULT_MAX_IDLE_CONNS, DEFAULT_MASTER_SLAVE)
	if err != nil {
		t.Error(err)
	}
	t.Log(MysqlClient)

	var args = make([]interface{}, 0)
	args = append(args, "haha")
	res, err := MysqlClient.FetchRowsForMaster("SELECT * FROM `test` WHERE `name` = ?", args...)
	if err != nil {
		t.Error(err)
	}
	for _, v := range res {
		t.Log(v)
	}
}

func Test_Custom(t *testing.T) {
	var masterConfig = NewConfig()
	masterConfig.User = "p_dev"
	masterConfig.Passwd = "r2evKK0QJj1oHQe1"
	masterConfig.Addr = "10.20.1.11:3306"
	masterConfig.DBName = "gobase_mysql_demo"
	masterConfig.Net = "tcp"
	masterConfig.Timeout = 1 * time.Second
	masterConfig.ReadTimeout = 1 * time.Second
	masterConfig.WriteTimeout = 1 * time.Second

	var slaveConfig = NewConfig()
	slaveConfig.User = "p_dev"
	slaveConfig.Passwd = "r2evKK0QJj1oHQe1"
	slaveConfig.Addr = "10.20.1.11:3306"
	slaveConfig.DBName = "gobase_mysql_demo"
	slaveConfig.Net = "tcp"
	slaveConfig.Timeout = 1 * time.Second
	slaveConfig.ReadTimeout = 1 * time.Second
	slaveConfig.WriteTimeout = 1 * time.Second

	MysqlClient, err := NewMysqlBaseDaoCustom(masterConfig, slaveConfig, DEFAULT_MASTER_SLAVE)
	MysqlClient.SetMaxOpenConns(100)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 5; i++ {
		sql := "Insert INTO `test` (`ver`, `createtime`, `updatetime`, `name`) VALUES (?, ?, ?, ?)"
		now := time.Now().Format(time.RFC3339)
		var args = make([]interface{}, 0)
		args = append(args, 1)
		args = append(args, now)
		args = append(args, now)
		args = append(args, "haha")
		id, err := MysqlClient.Insert(sql, args...)
		if err != nil {
			t.Error(err)
		}
		t.Log(id)
	}

	var args = make([]interface{}, 0)
	args = append(args, "haha")
	res, err := MysqlClient.FetchRows("SELECT * FROM `test` WHERE `name` = ?", args...)
	if err != nil {
		t.Error(err)
	}
	for _, v := range res {
		t.Log(v)
	}
}

type testJannaClient struct {
	ch chan janna.ServiceEvent
}

func newTestJannaClient() *testJannaClient {
	return &testJannaClient{make(chan janna.ServiceEvent)}
}

func (this *testJannaClient) ServiceRegister(service *janna.Service) error {
	return nil
}

func (this *testJannaClient) ServiceDeregister(key string) error {
	return nil
}

func (this *testJannaClient) ServiceGet(key string) (*janna.Service, error) {
	return nil, nil
}

func (this *testJannaClient) ServiceGetAll(key string) ([]janna.Service, error) {
	services := make([]janna.Service, 0)
	services = append(services, janna.Service{Key: "/service/test/mysql/default_master_1", ServiceValue: janna.ServiceValue{Tag: []string{masterTag, "default"}, Address: "10.20.1.11", Port: 3306, Weight: 100, User: "p_dev", Password: "r2evKK0QJj1oHQe1"}})
	services = append(services, janna.Service{Key: "/service/test/mysql/default_slave_1", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "default"}, Address: "10.20.1.11", Port: 3306, Weight: 100, User: "p_dev", Password: "r2evKK0QJj1oHQe1"}})
	services = append(services, janna.Service{Key: "/service/test/mysql/default_slave_2", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "default"}, Address: "10.20.1.11", Port: 3306, Weight: 100, User: "p_dev", Password: "r2evKK0QJj1oHQe1"}})
	services = append(services, janna.Service{Key: "/service/test/mysql/target0_master_1", ServiceValue: janna.ServiceValue{Tag: []string{masterTag, "target0"}, Address: "10.20.1.11", Port: 3306, Weight: 100, User: "p_dev", Password: "r2evKK0QJj1oHQe1"}})
	services = append(services, janna.Service{Key: "/service/test/mysql/target1_master_2", ServiceValue: janna.ServiceValue{Tag: []string{masterTag, "target1"}, Address: "10.20.1.11", Port: 3306, Weight: 100, User: "p_dev", Password: "r2evKK0QJj1oHQe1"}})
	services = append(services, janna.Service{Key: "/service/test/mysql/target0_slave_1", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "target0"}, Address: "10.20.1.11", Port: 3306, Weight: 100, User: "p_dev", Password: "r2evKK0QJj1oHQe1"}})
	services = append(services, janna.Service{Key: "/service/test/mysql/target1_slave_2", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "target1"}, Address: "10.20.1.11", Port: 3306, Weight: 100, User: "p_dev", Password: "r2evKK0QJj1oHQe1"}})

	return services, nil
}

func (this *testJannaClient) ServiceAddWatch(key string) error {
	return nil
}

func (this *testJannaClient) ServiceRemoveWatch(key string) error {
	return nil
}

func (this *testJannaClient) ServiceGetWatch() (chan janna.ServiceEvent, error) {
	return this.ch, nil
}

func (this *testJannaClient) ServiceCloseWatch() error {
	close(this.ch)

	return nil
}

func Test_JannaCustom(t *testing.T) {
	var masterConfig = NewConfig()
	masterConfig.User = "p_dev"
	masterConfig.Passwd = "r2evKK0QJj1oHQe1"
	masterConfig.DBName = "gobase_mysql_demo"

	var slaveConfig = NewConfig()
	slaveConfig.User = "p_dev"
	slaveConfig.Passwd = "r2evKK0QJj1oHQe1"
	slaveConfig.DBName = "gobase_mysql_demo"

	tjc := newTestJannaClient()

	MysqlClient, err := NewMysqlBaseDaoJannaCustom(tjc, "test", "default", masterConfig, slaveConfig)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 5; i++ {
		sql := "Insert INTO `test` (`ver`, `createtime`, `updatetime`, `name`) VALUES (?, ?, ?, ?)"
		now := time.Now().Format(time.RFC3339)
		var args = make([]interface{}, 0)
		args = append(args, 1)
		args = append(args, now)
		args = append(args, now)
		args = append(args, "haha")
		id, err := MysqlClient.Insert(sql, args...)
		if err != nil {
			t.Error(err)
		}
		t.Log(id)
	}

	var args = make([]interface{}, 0)
	args = append(args, "haha")
	res, err := MysqlClient.FetchRows("SELECT * FROM `test` WHERE `name` = ?", args...)
	if err != nil {
		t.Error(err)
	}
	for _, v := range res {
		t.Log(v)
	}
}

func Test_JannaCustomEnc(t *testing.T) {
	var masterConfig = NewConfig()
	masterConfig.DBName = "gobase_mysql_demo"

	var slaveConfig = NewConfig()
	slaveConfig.DBName = "gobase_mysql_demo"

	tjc := newTestJannaClient()

	MysqlClient, err := NewMysqlBaseDaoJannaCustomEnc(tjc, "test", "default", masterConfig, slaveConfig)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 5; i++ {
		sql := "Insert INTO `test` (`ver`, `createtime`, `updatetime`, `name`) VALUES (?, ?, ?, ?)"
		now := time.Now().Format(time.RFC3339)
		var args = make([]interface{}, 0)
		args = append(args, 1)
		args = append(args, now)
		args = append(args, now)
		args = append(args, "haha")
		id, err := MysqlClient.Insert(sql, args...)
		if err != nil {
			t.Error(err)
		}
		t.Log(id)
	}

	tjc.ch <- janna.ServiceEvent{janna.OptPut, janna.Service{Key: "/service/test/mysql/default_slave_2", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "default"}, Address: "10.20.1.11", Port: 3306, Weight: 100, User: "p_dev", Password: "r2evKK0QJj1oHQ"}}}
	time.Sleep(time.Second)

	var args = make([]interface{}, 0)
	args = append(args, "haha")
	res, err := MysqlClient.FetchRows("SELECT * FROM `test` WHERE `name` = ?", args...)
	if err != nil {
		t.Error(err)
	}
	for _, v := range res {
		t.Log(v)
	}

	res, err = MysqlClient.FetchRows("SELECT * FROM `test` WHERE `name` = ?", args...)
	if err != nil {
		t.Error(err)
	}
	for _, v := range res {
		t.Log(v)
	}
}

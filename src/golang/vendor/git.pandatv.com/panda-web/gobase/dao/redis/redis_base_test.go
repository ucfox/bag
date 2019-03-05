package redis

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	janna "git.pandatv.com/panda-public/janna/client"
)

func equalInt(a, b int, msg string) {
	if a != b {
		log.Panicln(msg)
	}
}

func equalError(a, b error, msg string) {
	if a != b {
		log.Panicln(msg)
	}
}

func notEqualError(a, b error, msg string) {
	if a == b {
		log.Panicln(msg)
	}
}

func equalString(a, b, msg string) {
	if a != b {
		log.Panicln(msg)
	}
}

func equalBool(a, b bool, msg string) {
	if a != b {
		log.Panicln(msg)
	}
}

func equalRegexp(pattern, str, msg string) {
	match, err := regexp.Match(pattern, []byte(str))
	if err != nil || !match {
		log.Panicln(msg)
	}
}

func TestInitCustomPool(t *testing.T) {
	fmt.Println("==============test init new pool==============")
	cliPool := NewRedisClientCustom("10.20.1.20:7300", "", "10.20.1.20:7301", "", 1000*time.Millisecond, 500*time.Millisecond, 500*time.Millisecond, 60*time.Second, 50, 3, true)
	_, err := cliPool.Del("foo", "testfoo")
	equalError(nil, err, "del failed")

	val_s, err := cliPool.Set("foo", "bar", 0)
	equalString(val_s, "OK", "set failed")
	equalError(nil, err, "set failed")

	val_s, err = cliPool.Get("foo")
	equalString(val_s, "bar", "get failed")
	equalError(nil, err, "get failed")

	val_i, err := cliPool.IncrBy("testfoo", 3)
	equalInt(val_i, 3, "incrby failed")
	equalError(nil, err, "incrby failed")

	cliPool.CloseRedis()
}

func TestConnectTimeout(t *testing.T) {
	fmt.Println("==============test connect timeout==============")

	cliPool := NewRedisClientCustom("10.20.1.20:7300", "", "10.20.1.20:7301", "", 1*time.Nanosecond, 500*time.Millisecond, 500*time.Millisecond, 60*time.Second, 50, 3, true)

	_, err := cliPool.Set("foo", "bar", 0)
	notEqualError(nil, err, "connect timeout failed")

	cliPool.CloseRedis()

}

func TestReadTimeout(t *testing.T) {
	fmt.Println("==============test read timeout==============")

	cliPool := NewRedisClientCustom("10.20.1.20:7300", "", "10.20.1.20:7301", "", 1000*time.Millisecond, 5*time.Nanosecond, 500*time.Millisecond, 60*time.Second, 50, 3, true)

	_, err := cliPool.Set("foo", "bar", 0)
	notEqualError(nil, err, "read timeout failed")

	_, err = cliPool.Get("foo")
	notEqualError(nil, err, "read timeout failed")

	cliPool.CloseRedis()

}

func TestWriteTimeout(t *testing.T) {
	fmt.Println("==============test write timeout==============")

	cliPool := NewRedisClientCustom("10.20.1.20:7300", "", "10.20.1.20:7301", "", 1000*time.Millisecond, 500*time.Millisecond, 5*time.Nanosecond, 60*time.Second, 50, 3, true)

	_, err := cliPool.Set("foo", "bar", 0)
	notEqualError(nil, err, "write timeout failed")

	_, err = cliPool.Get("foo")
	notEqualError(nil, err, "write timeout failed")

	cliPool.CloseRedis()

}

func TestMaxActive(t *testing.T) {
	fmt.Println("==============test max active==============")

	cliPool := NewRedisClientCustom("10.20.1.20:7300", "", "10.20.1.20:7301", "", 1000*time.Millisecond, 500*time.Millisecond, 500*time.Millisecond, 1*time.Second, 2, 2, false)
	num := 5

	for i := 0; i < num; i++ {
		go func() {
			_, err := cliPool.Set("foo", "bar", 0)
			if err != nil {
				fmt.Println("set fail! err:", err)
			}

			val, err := cliPool.Get("foo")
			if err != nil {
				fmt.Println("get fail! err:", err)
			} else {
				fmt.Println(val)
			}
		}()
	}

	select {
	case <-time.After(time.Second * 2):
	}
	cliPool.CloseRedis()
}

func TestWait(t *testing.T) {
	fmt.Println("==============test wait==============")

	cliPool := NewRedisClientCustom("10.20.1.20:7300", "", "10.20.1.20:7301", "", 1000*time.Millisecond, 500*time.Millisecond, 500*time.Millisecond, 1*time.Second, 1, 1, true)
	num := 5

	for i := 0; i < num; i++ {
		go func() {
			val_s, err := cliPool.Set("foo", "bar", 0)
			equalString(val_s, "OK", "set failed")
			equalError(nil, err, "wait failed")

			val_s, err = cliPool.Get("foo")
			equalString(val_s, "bar", "get failed")
			equalError(nil, err, "wait failed")
		}()
	}

	select {
	case <-time.After(time.Millisecond * 500):
	}
	cliPool.CloseRedis()
}

func TestReadStale(t *testing.T) {
	fmt.Println("==============test read stale==============")
	cliPool := NewRedisClient("10.20.1.20:7300", "", "127.0.0.1:44444", "")

	val_s, err := cliPool.Set("foo", "bar", 0)
	equalString(val_s, "OK", "set failed")
	equalError(nil, err, "set failed")

	val_s, err = cliPool.Get("foo")
	notEqualError(nil, err, "get failed")

	cliPool.SetReadStale(false)

	val_s, err = cliPool.Get("foo")
	equalString(val_s, "bar", "get failed")
	equalError(nil, err, "get failed")

	cliPool.CloseRedis()
}

func TestFilterFailConn(t *testing.T) {
	fmt.Println("==============test filter fail conn==============")
	cliPool := NewRedisClient("10.20.1.20:7300", "", "10.20.1.20:7301,10.20.1.20:73000,10.20.1.20:73000", "")

	_, err := cliPool.Get("foo")
	equalError(nil, err, "get failed")

	cliPool.SetFilterFailConn(false)

	_, err = cliPool.Get("foo")
	notEqualError(nil, err, "get failed")

	cliPool.CloseRedis()
}

func TestCheckBadConn(t *testing.T) {
	fmt.Println("==============test check bad conn==============")
	cliPool := NewRedisClient("10.20.1.20:7300", "", "10.20.1.20:7301,10.20.1.20:73000", "")

	_, err := cliPool.Get("foo")
	equalError(nil, err, "get failed")

	_, err = cliPool.Get("foo")
	notEqualError(nil, err, "get failed")

	time.Sleep(time.Second * 3)

	_, err = cliPool.Get("foo")
	equalError(nil, err, "get failed")

	_, err = cliPool.Get("foo")
	equalError(nil, err, "get failed")

	cliPool.CloseRedis()
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
	services = append(services, janna.Service{Key: "/service/test/redis/default_master_1", ServiceValue: janna.ServiceValue{Tag: []string{masterTag, "default"}, Address: "10.20.1.20", Port: 7300, Weight: 100, User: "", Password: ""}})
	services = append(services, janna.Service{Key: "/service/test/redis/default_slave_1", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "default"}, Address: "10.20.1.20", Port: 7301, Weight: 100, User: "", Password: ""}})
	services = append(services, janna.Service{Key: "/service/test/redis/web_master_1", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "web"}, Address: "10.20.1.20", Port: 7300, Weight: 100}})
	services = append(services, janna.Service{Key: "/service/test/redis/web_slave_1", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "web"}, Address: "10.20.1.20", Port: 7301, Weight: 100}})
	services = append(services, janna.Service{Key: "/service/test/redis/web_slave_2", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "web"}, Address: "10.20.1.20", Port: 7300, Weight: 100}})

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

func TestClientWithJanna(t *testing.T) {
	fmt.Println("==============test client with janna==============")
	tjc := newTestJannaClient()
	cliPool := NewRedisClientJanna(tjc, "test", "default", "", "")

	_, err := cliPool.Del("foo", "testfoo")
	equalError(nil, err, "del failed")

	val_s, err := cliPool.Set("foo", "bar", 0)
	equalString(val_s, "OK", "set failed")
	equalError(nil, err, "set failed")

	val_s, err = cliPool.Get("foo")
	equalString(val_s, "bar", "get failed")
	equalError(nil, err, "get failed")

	val_i, err := cliPool.IncrBy("testfoo", 3)
	equalInt(val_i, 3, "incrby failed")
	equalError(nil, err, "incrby failed")

	tjc.ch <- janna.ServiceEvent{janna.OptPut, janna.Service{Key: "/service/test/redis/default_slave_1", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "default"}, Address: "10.20.1.20", Port: 7300, Weight: 100}}}

	time.Sleep(time.Second)

	val_s, err = cliPool.Get("foo")
	equalString(val_s, "bar", "get failed")
	equalError(nil, err, "get failed")

	tjc.ch <- janna.ServiceEvent{janna.OptPut, janna.Service{Key: "/service/test/redis/default_slave_1", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "default"}, Address: "10.20.1.20", Port: 7301, Weight: 100}}}
	time.Sleep(time.Second)
	tjc.ch <- janna.ServiceEvent{janna.OptDelete, janna.Service{Key: "/service/test/redis/default_slave_1", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "default"}, Address: "10.20.1.20", Port: 7301, Weight: 100}}}
	time.Sleep(time.Second)

	cliPool.CloseRedis()
}

func TestClientWithJannaEnc(t *testing.T) {
	fmt.Println("==============test client with janna enc==============")
	tjc := newTestJannaClient()
	cliPool := NewRedisClientJannaEnc(tjc, "test", "default")

	_, err := cliPool.Del("foo", "testfoo")
	equalError(nil, err, "del failed")

	val_s, err := cliPool.Set("foo", "bar", 0)
	equalString(val_s, "OK", "set failed")
	equalError(nil, err, "set failed")

	val_s, err = cliPool.Get("foo")
	equalString(val_s, "bar", "get failed")
	equalError(nil, err, "get failed")

	val_i, err := cliPool.IncrBy("testfoo", 3)
	equalInt(val_i, 3, "incrby failed")
	equalError(nil, err, "incrby failed")

	tjc.ch <- janna.ServiceEvent{janna.OptPut, janna.Service{Key: "/service/test/redis/default_slave_1", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "default"}, Address: "10.20.1.20", Port: 7300, Weight: 100, User: "", Password: ""}}}

	time.Sleep(time.Second)

	val_s, err = cliPool.Get("foo")
	equalString(val_s, "bar", "get failed")
	equalError(nil, err, "get failed")

	tjc.ch <- janna.ServiceEvent{janna.OptPut, janna.Service{Key: "/service/test/redis/default_slave_1", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "default"}, Address: "10.20.1.20", Port: 7301, Weight: 100, User: "", Password: ""}}}
	time.Sleep(time.Second)
	tjc.ch <- janna.ServiceEvent{janna.OptDelete, janna.Service{Key: "/service/test/redis/default_slave_1", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "default"}, Address: "10.20.1.20", Port: 7301, Weight: 100, User: "", Password: ""}}}
	time.Sleep(time.Second)

	cliPool.CloseRedis()
}

func TestClientWithJannaWithTag(t *testing.T) {
	fmt.Println("==============test client with janna with tag==============")
	tjc := newTestJannaClient()
	cliPool := NewRedisClientJanna(tjc, "test", "web", "", "")

	val_s, err := cliPool.Get("foo")
	equalString(val_s, "bar", "get failed")
	equalError(nil, err, "get failed")

	val_s, err = cliPool.Get("foo")
	equalString(val_s, "bar", "get failed")
	equalError(nil, err, "get failed")

	val_s, err = cliPool.Get("foo")
	equalString(val_s, "bar", "get failed")
	equalError(nil, err, "get failed")

	cliPool.CloseRedis()
}

func TestCheckBadPoolWithJanna(t *testing.T) {
	fmt.Println("==============test check bad pool with janna==============")
	tjc := newTestJannaClient()
	cliPool := NewRedisClientJanna(tjc, "test", "default", "", "")

	_, err := cliPool.Del("foo")
	equalError(nil, err, "del failed")

	val_s, err := cliPool.Set("foo", "bar", 0)
	equalString(val_s, "OK", "set failed")
	equalError(nil, err, "set failed")

	val_s, err = cliPool.Get("foo")
	equalString(val_s, "bar", "get failed")
	equalError(nil, err, "get failed")

	tjc.ch <- janna.ServiceEvent{janna.OptPut, janna.Service{Key: "/service/test/redis/default_slave_2", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "default"}, Address: "10.20.1.20", Port: 73000, Weight: 100}}}
	time.Sleep(time.Second)

	_, err = cliPool.Get("foo")
	notEqualError(nil, err, "get failed")
	_, err = cliPool.Get("foo")
	equalError(nil, err, "get failed")

	time.Sleep(time.Second * 3)

	_, err = cliPool.Get("foo")
	equalError(nil, err, "get failed")

	_, err = cliPool.Get("foo")
	equalError(nil, err, "get failed")

	tjc.ch <- janna.ServiceEvent{janna.OptPut, janna.Service{Key: "/service/test/redis/default_slave_2", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "default"}, Address: "10.20.1.20", Port: 7300, Weight: 100}}}
	time.Sleep(time.Second * 3)

	_, err = cliPool.Get("foo")
	equalError(nil, err, "get failed")

	_, err = cliPool.Get("foo")
	equalError(nil, err, "get failed")

	cliPool.CloseRedis()
}

func TestGetConnWithJanna(t *testing.T) {
	fmt.Println("==============test get conn==============")
	tjc := newTestJannaClient()
	cliPool := NewRedisClientJanna(tjc, "test", "default", "", "")

	_, err := cliPool.Del("foo")
	equalError(nil, err, "del failed")

	val_s, err := cliPool.Set("foo", "bar", 0)
	equalString(val_s, "OK", "set failed")
	equalError(nil, err, "set failed")

	val_s, err = cliPool.Get("foo")
	equalString(val_s, "bar", "get failed")
	equalError(nil, err, "get failed")

	tjc.ch <- janna.ServiceEvent{janna.OptPut, janna.Service{Key: "/service/test/redis/default_slave_1", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "default"}, Address: "10.20.1.20", Port: 73000, Weight: 100}}}
	tjc.ch <- janna.ServiceEvent{janna.OptPut, janna.Service{Key: "/service/test/redis/default_slave_2", ServiceValue: janna.ServiceValue{Tag: []string{slaveTag, "default"}, Address: "10.20.1.20", Port: 73001, Weight: 100}}}
	time.Sleep(time.Second)

	_, err = cliPool.Get("foo")
	notEqualError(nil, err, "get failed")
	_, err = cliPool.Get("foo")
	notEqualError(nil, err, "get failed")

	time.Sleep(time.Second * 3)

	_, err = cliPool.Get("foo")
	notEqualError(nil, err, "get failed")
	_, err = cliPool.Get("foo")
	notEqualError(nil, err, "get failed")

	cliPool.CloseRedis()
}

func BenchmarkGetSet(b *testing.B) {
	cliPool := NewRedisClientCustom("10.20.1.20:7300", "", "10.20.1.20:7301,10.20.1.20:7300", "", Default_connect_timeout, Default_read_timeout, Default_write_timeout, Default_idle_timeout, 50, 50, Default_wait)
	defer cliPool.CloseRedis()

	b.SetParallelism(100)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			val_s, err := cliPool.Set("foo", "bar", 0)
			if val_s != "OK" || err != nil {
				fmt.Println("set error:", err)
			}

			val_s, err = cliPool.Get("foo")
			if val_s != "bar" || err != nil {
				fmt.Println("get error:", err)
			}
		}
	})
}

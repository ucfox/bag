package redis

import (
	"fmt"
	"testing"
)

func TestKeyMethod(t *testing.T) {
	fmt.Println("==============test key method==============")
	cliPool := NewRedisClient("10.20.1.20:7300", "", "10.20.1.20:7301", "")

	//key
	_, err := cliPool.Del("foo", "foo1", "foo2")
	equalError(nil, err, "del failed")

	val_i, err := cliPool.Expire("foo", 10)
	equalInt(0, val_i, "expire failed")
	equalError(nil, err, "expire failed")

	val_b, err := cliPool.Exists("foo")
	equalBool(false, val_b, "exists failed")
	equalError(nil, err, "exists failed")

	_, err = cliPool.RandomKey()
	equalError(nil, err, "randomkey failed")

	ttl, err := cliPool.TTL("foo")
	equalError(nil, err, "ttl failed")
	fmt.Println("ttl:", ttl)

	cliPool.CloseRedis()
}

func TestStringMethod(t *testing.T) {
	fmt.Println("==============test string method==============")
	cliPool := NewRedisClient("10.20.1.20:7300", "", "10.20.1.20:7301", "")

	cliPool.Del("foo", "foo1", "foo2")
	//string
	val_i, err := cliPool.Append("foo", "bar")
	equalInt(val_i, 3, "append failed")
	equalError(nil, err, "append failed")

	val_s, err := cliPool.Get("foo")
	equalString(val_s, "bar", "get failed")
	equalError(nil, err, "get failed")

	val_s, err = cliPool.GetRaw("fdjlafjka")
	equalString(val_s, "", "getraw failed")
	equalError(err, ErrNil, "getraw failed")

	val_s, err = cliPool.GetSet("foo", "bar")
	equalString(val_s, "bar", "getset failed")
	equalError(nil, err, "getset failed")

	val_i, err = cliPool.IncrBy("foo", 10)
	equalInt(0, val_i, "incrby failed")
	notEqualError(nil, err, "incrby failed")

	val_ss, err := cliPool.MGet("foo", "foo2")
	equalString(val_ss[0], "bar", "mget failed")
	equalString(val_ss[1], "", "mget failed")
	equalError(nil, err, "mget failed")

	val_s, err = cliPool.MSet("foo1", "bar1", "foo2", "bar2")
	equalString(val_s, "OK", "mset failed")
	equalError(nil, err, "mset failed")

	val_s, err = cliPool.Set("foo1", "bar1", 0)
	equalString(val_s, "OK", "set failed")
	equalError(nil, err, "set failed")

	val_s, err = cliPool.SetEPX("foo1", "bar1", 10, 20000, 2)
	equalString(val_s, "OK", "set failed")
	equalError(nil, err, "set failed")

	val_i, err = cliPool.SetNx("foo1", "bar2")
	equalInt(0, val_i, "setnx failed")
	equalError(nil, err, "setnx failed")
	val_s, _ = cliPool.Get("foo1")
	equalString(val_s, "bar1", "setnx failed")

	val_s, err = cliPool.MSet("foo1", "bar1", "foo2", "bar2")
	equalString(val_s, "OK", "setex failed")
	equalError(nil, err, "setex failed")

	val_i, err = cliPool.StrLen("foo")
	equalInt(val_i, 3, "strlen failed")
	equalError(nil, err, "strlen failed")

	cliPool.CloseRedis()
}

func TestBitMethod(t *testing.T) {
	fmt.Println("==============test bit method==============")
	cliPool := NewRedisClient("10.20.1.20:7300", "", "10.20.1.20:7301", "")

	//bit
	cliPool.Set("mykey", "foobar", 0)
	val_i, err := cliPool.BitCount("mykey")
	equalInt(val_i, 26, "bitcount failed")
	equalError(nil, err, "bitcount failed")

	val_i, err = cliPool.BitCountWithPos("mykey", 1, 1)
	equalInt(val_i, 6, "bitcount failed")
	equalError(nil, err, "bitcount failed")

	val_i, err = cliPool.BitCount("mykeyNoExist")
	equalInt(val_i, 0, "bitcount failed")
	equalError(nil, err, "bitcount failed")
	cliPool.Del("mykey")

	cliPool.Set("mykey1", "foobar", 0)
	cliPool.Set("mykey2", "abcdef", 0)
	val_i, err = cliPool.BitOp("AND", "mydest", "mykey1", "mykey2")
	equalInt(val_i, 6, "bitop failed")
	equalError(nil, err, "bitop failed")
	val_s, err := cliPool.Get("mydest")
	equalString(val_s, "`bc`ab", "bitop failed")
	equalError(nil, err, "bitop failed")
	cliPool.Del("mydest", "mykey1", "mykey2")

	cliPool.Set("mykey", "\xff\xf0\x00", 0)
	val_i, err = cliPool.BitPos("mykey", 0)
	equalInt(val_i, 12, "bitpos failed")
	equalError(nil, err, "bitpos failed")

	cliPool.Set("mykey", "\x00\x00\x00", 0)
	val_i, err = cliPool.BitPos("mykey", 1)
	equalInt(val_i, -1, "bitpos failed")
	equalError(nil, err, "bitpos failed")

	cliPool.Set("mykey", "\x00\xff\xf0", 0)
	val_i, err = cliPool.BitPosWithPos("mykey", 1, 0, -1)
	equalInt(val_i, 8, "bitpos failed")
	equalError(nil, err, "bitpos failed")

	val_i, err = cliPool.BitPosWithPos("mykey", 1, 2, -1)
	equalInt(val_i, 16, "bitpos failed")
	equalError(nil, err, "bitpos failed")
	cliPool.Del("mykey")

	val_i, err = cliPool.SetBit("mykey", 7, 1)
	equalInt(val_i, 0, "setbit failed")
	equalError(nil, err, "setbit failed")

	val_i, err = cliPool.GetBit("mykey", 0)
	equalInt(val_i, 0, "getbit failed")
	equalError(nil, err, "getbit failed")

	val_i, err = cliPool.GetBit("mykey", 7)
	equalInt(val_i, 1, "getbit failed")
	equalError(nil, err, "getbit failed")

	val_i, err = cliPool.GetBit("mykey", 100)
	equalInt(val_i, 0, "getbit failed")
	equalError(nil, err, "getbit failed")
	cliPool.Del("mykey")

	cliPool.CloseRedis()
}

func TestHashMethod(t *testing.T) {
	fmt.Println("==============test hash method==============")
	cliPool := NewRedisClient("10.20.1.20:7300", "", "10.20.1.20:7301", "")

	//hash
	_, err := cliPool.Del("hfoo", "hfoo1", "hfoo2")
	_, err = cliPool.HDel("hfoo", "hfoo1", "hfoo2")
	equalError(nil, err, "hdel failed")

	val_b, err := cliPool.HExists("hfoo", "bar")
	equalBool(false, val_b, "hexists failed")
	equalError(nil, err, "hexists failed")

	val_s, err := cliPool.HGet("hfoo", "bar")
	equalString(val_s, "", "hget failed")
	equalError(nil, err, "hget failed")

	val_mss, err := cliPool.HGetAll("hfoo")
	equalInt(0, len(val_mss), "hgetall failed")
	equalError(nil, err, "hgetall failed")

	val_i, err := cliPool.HIncr("hfoo", "ibar")
	equalInt(val_i, 1, "hincr failed")
	equalError(nil, err, "hincr failed")

	val_i, err = cliPool.HIncrBy("hfoo", "ibar", 2)
	equalInt(val_i, 3, "hincrby failed")
	equalError(nil, err, "hincrby failed")

	cliPool.HDel("hfoo", "ibar")

	val_ss, err := cliPool.HKeys("hfoo")
	equalInt(0, len(val_ss), "hkeys failed")
	equalError(nil, err, "hkeys failed")

	val_i, err = cliPool.HLen("hfoo")
	equalInt(0, val_i, "hlen failed")
	equalError(nil, err, "hlen failed")

	val_ss, err = cliPool.HMGet("hfoo", "bar", "bar1")
	equalString(val_ss[0], "", "hmget failed")
	equalString(val_ss[1], "", "hmget failed")
	equalError(nil, err, "hmget failed")

	val_s, err = cliPool.HMSet("hfoo", "bar", "0", "bar1", "1")
	equalString(val_s, "OK", "hmset failed")
	equalError(nil, err, "hmset failed")

	val_i, err = cliPool.HSet("hfoo", "bar2", "2")
	equalInt(val_i, 1, "hset failed")
	equalError(nil, err, "hset failed")

	val_ss, err = cliPool.HVals("hfoo")
	equalString(val_ss[0], "0", "hvals failed")
	equalString(val_ss[1], "1", "hvals failed")
	equalString(val_ss[2], "2", "hvals failed")
	equalError(nil, err, "hvals failed")

	cliPool.CloseRedis()
}

func TestListMethod(t *testing.T) {
	fmt.Println("==============test list method==============")
	cliPool := NewRedisClient("10.20.1.20:7300", "", "10.20.1.20:7301", "")

	//list
	_, err := cliPool.Del("hfoo")
	_, err = cliPool.LIndex("hfoo", 0)
	equalError(nil, err, "lindex failed")

	val_i, err := cliPool.LInsert("hfoo", "BEFORE", "bar", "bar0")
	equalError(nil, err, "linsert failed")

	val_i, err = cliPool.LLen("hfoo")
	equalInt(0, val_i, "llen failed")
	equalError(nil, err, "llen failed")

	val_s, err := cliPool.LPop("hfoo")
	equalError(nil, err, "lindex failed")

	val_i, err = cliPool.LPush("hfoo", "bar", "bar1")
	equalInt(val_i, 2, "lpush failed")
	equalError(nil, err, "lpush failed")

	val_ss, err := cliPool.LRange("hfoo", 0, -1)
	equalString(val_ss[0], "bar1", "lrange failed")
	equalString(val_ss[1], "bar", "lrange failed")
	equalError(nil, err, "lrange failed")

	val_i, err = cliPool.LRem("hfoo", 0, "bar1")
	equalInt(val_i, 1, "lrem failed")
	equalError(nil, err, "lrem failed")

	val_s, err = cliPool.LSet("hfoo", 0, "barbar")
	equalString(val_s, "OK", "lset failed")
	equalError(nil, err, "lset failed")

	val_s, err = cliPool.LTrim("hfoo", 1, -1)
	equalString(val_s, "OK", "ltrim failed")
	equalError(nil, err, "ltrim failed")

	val_s, err = cliPool.RPop("hfoo")
	equalString(val_s, "", "rpop failed")
	equalError(nil, err, "rpop failed")

	val_i, err = cliPool.RPush("hfoo", "bar")
	equalInt(val_i, 1, "rpush failed")
	equalError(nil, err, "rpush failed")

	cliPool.CloseRedis()
}

func TestSetMethod(t *testing.T) {
	fmt.Println("==============test set method==============")
	cliPool := NewRedisClient("10.20.1.20:7300", "", "10.20.1.20:7301", "")

	//set
	_, err := cliPool.Del("sfoo", "sfoo1")
	val_i, err := cliPool.SAdd("sfoo", "bar", "bar1")
	equalInt(val_i, 2, "sadd failed")
	equalError(nil, err, "sadd failed")
	_, _ = cliPool.SAdd("sfoo1", "bar2", "bar1")

	val_i, err = cliPool.SCard("sfoo")
	equalInt(val_i, 2, "scard failed")
	equalError(nil, err, "scard failed")

	val_ss, err := cliPool.SDiff("sfoo", "sfoo1")
	equalInt(len(val_ss), 1, "sdiff failed")
	equalError(nil, err, "sdiff failed")

	val_ss, err = cliPool.SInter("sfoo", "sfoo1")
	equalInt(len(val_ss), 1, "sinter failed")
	equalError(nil, err, "sinter failed")

	val_b, err := cliPool.SIsMember("sfoo", "bar")
	equalBool(true, val_b, "sismember failed")
	equalError(nil, err, "sismember failed")

	val_ss, err = cliPool.SMembers("sfoo")
	equalInt(len(val_ss), 2, "smembers failed")
	equalError(nil, err, "smembers failed")

	_, err = cliPool.SPop("sfoo")
	equalError(nil, err, "spop failed")

	_, err = cliPool.SRandMember("sfoo")
	equalError(nil, err, "srandmember failed")

	val_i, err = cliPool.SRem("sfoo1", "bar1", "bar2")
	equalInt(val_i, 2, "srem failed")
	equalError(nil, err, "srem failed")

	val_ss, err = cliPool.SUnion("sfoo", "sfoo1")
	equalInt(len(val_ss), 1, "sunion failed")
	equalError(nil, err, "sunion failed")

	cliPool.CloseRedis()
}

func TestZsetMethod(t *testing.T) {
	fmt.Println("==============test zset method==============")
	cliPool := NewRedisClient("10.20.1.20:7300", "", "10.20.1.20:7301", "")

	//zset
	_, err := cliPool.Del("zfoo")
	val_i, err := cliPool.ZAdd("zfoo", 1, "bar1")
	equalInt(val_i, 1, "zadd failed")
	equalError(nil, err, "zadd failed")

	val_i, err = cliPool.ZMAdd("zfoo", &ScorePair{"bar2", 2}, &ScorePair{"bar3", 3})
	equalInt(val_i, 2, "zmadd failed")
	equalError(nil, err, "zmadd failed")

	val_i, err = cliPool.ZCard("zfoo")
	equalInt(val_i, 3, "zcard failed")
	equalError(nil, err, "zcard failed")

	val_i, err = cliPool.ZCount("zfoo", "-inf", "+inf")
	equalInt(val_i, 3, "zcount failed")
	equalError(nil, err, "zcount failed")

	val_i, err = cliPool.ZIncrBy("zfoo", 10, "bar1")
	equalInt(val_i, 11, "zincrby failed")
	equalError(nil, err, "zincrby failed")

	val_ss, err := cliPool.ZRange("zfoo", 0, -1)
	equalInt(len(val_ss), 3, "zrange failed")
	equalError(nil, err, "zrange failed")

	val_sps, err := cliPool.ZRangeWithScores("zfoo", 0, -1)
	equalInt(len(val_sps), 3, "zrangewithscores failed")
	equalError(nil, err, "zrangewithscores failed")

	val_i, err = cliPool.ZRank("zfoo", "bar1")
	equalInt(val_i, 2, "zrank failed")
	equalError(nil, err, "zrank failed")

	val_i, err = cliPool.ZRem("zfoo", "bar1", "bar2")
	equalInt(val_i, 2, "zrem failed")
	equalError(nil, err, "zrem failed")

	_, _ = cliPool.ZMAdd("zfoo", &ScorePair{"bar2", 2}, &ScorePair{"bar3", 3})
	val_i, err = cliPool.ZRemRangeByRank("zfoo", 0, -1)
	equalInt(val_i, 2, "zremrangebyrank failed")
	equalError(nil, err, "zremrangebyrank failed")

	_, _ = cliPool.ZMAdd("zfoo", &ScorePair{"bar2", 2}, &ScorePair{"bar3", 3})
	val_ss, err = cliPool.ZRevRange("zfoo", 0, -1)
	equalInt(len(val_ss), 2, "zrevrange failed")
	equalError(nil, err, "zrevrange failed")

	val_sps, err = cliPool.ZRevRangeWithScores("zfoo", 0, -1)
	equalInt(len(val_sps), 2, "zrevrangewithscores failed")
	equalError(nil, err, "zrevrangewithscores failed")

	val_i, err = cliPool.ZRevRank("zfoo", "bar3")
	equalInt(val_i, 0, "zrevrank failed")
	equalError(nil, err, "zrevrank failed")

	val_i, err = cliPool.ZScore("zfoo", "bar3")
	equalInt(val_i, 3, "zscore failed")
	equalError(nil, err, "zscore failed")

	cliPool.Set("foo", "bar", 0)
	script := cliPool.Script(1, "return redis.call('get', KEYS[1])")
	err = cliPool.LoadScript(script)
	equalError(nil, err, "loadscript failed")
	val_inter, err := cliPool.Eval(script, "foo")
	equalString(string(val_inter.([]byte)), "bar", "eval failed")
	equalError(nil, err, "eval failed")
	cliPool.Del("foo")

	pipe, err := cliPool.PipeLine(false)
	equalError(nil, err, "pipe line failed")
	err = cliPool.PipeSend(pipe, "set", "foo", "bar")
	equalError(nil, err, "pipe send failed")
	err = cliPool.PipeSend(pipe, "get", "foo")
	equalError(nil, err, "pipe send failed")
	val_inter, err = cliPool.PipeExec(pipe)
	equalString(val_inter.([]interface{})[0].(string), "OK", "pipe exec failed")
	equalString(string(val_inter.([]interface{})[1].([]byte)), "bar", "pipe exec failed")
	equalError(nil, err, "pipe exec failed")
	err = cliPool.PipeClose(pipe)
	equalError(nil, err, "pipe close failed")
	cliPool.Del("foo")

	cliPool.CloseRedis()
}

func TestGeoMethod(t *testing.T) {
	fmt.Println("==============test geo method==============")
	cliPool := NewRedisClient("10.20.1.20:7200", "", "10.20.1.20:7200", "")

	cliPool.Del("geo|test")

	val_i, err := cliPool.GeoAdd("geo|test", &GeoPair{Geo{116.30965, 39.99173}, "北邮"}, &GeoPair{Geo{126.65226, 45.82012}, "北科"})
	equalInt(val_i, 2, "geoadd failed")
	equalError(nil, err, "geoadd failed")

	dist, err := cliPool.GeoDist("geo|test", "北邮", "北科", "m")
	fmt.Println(dist)
	equalError(nil, err, "geodist failed")

	hash, err := cliPool.GeoHash("geo|test", "北科", "北语", "北邮")
	fmt.Println(hash)
	equalError(nil, err, "geohash failed")

	geopos, err := cliPool.GeoPos("geo|test", "北科", "北语", "北邮", "北航")
	for _, p := range geopos {
		fmt.Println(p)
	}
	equalError(nil, err, "geopos failed")

	val_slice, err := cliPool.GeoRadius("geo|test", 116.30965, 39.99173, 5, "m", true, true, true, 0, 0)
	for _, vs := range val_slice {
		fmt.Println(vs)
	}
	equalError(nil, err, "georadius failed")

	val_slice, err = cliPool.GeoRadiusByMember("geo|test", "北邮", 5000, "km", true, true, true, 0, 0)
	for _, vs := range val_slice {
		fmt.Println(vs)
	}
	equalError(nil, err, "georadiusbymember failed")
}

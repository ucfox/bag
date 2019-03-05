package redis

import (
	"net"
	"strconv"

	"git.pandatv.com/panda-web/gobase/log"
	"github.com/garyburd/redigo/redis"
)

// zset 的 member 和 score 对，字段名根据业务需要再修改
type ScorePair struct {
	Member string `json:"member"`
	Score  int    `json:"score"`
}

type Geo struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type GeoPair struct {
	Geo
	Member string `json:"member"`
}

type GeoPairDistHash struct {
	GeoPair
	Dist float64
	Hash int
}

type LuaScript struct {
	*redis.Script
	write bool
}

type Pipe struct {
	redis.Conn
}

type Message struct {
	redis.Message
}

type SubscribeT struct {
	conn redis.PubSubConn
	C    chan []byte
}

func strSliToInterfSli(keys ...string) []interface{} {
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		args[i] = key
	}

	return args
}

func strSliToInterfSliTwo(key string, members ...string) []interface{} {
	args := make([]interface{}, len(members)+1)
	args[0] = key
	for i, member := range members {
		args[i+1] = member
	}

	return args
}

func (dao *RedisBaseDao) doInt(read bool, methodname string, masknil bool, args ...interface{}) (int, error) {
	var reply interface{}
	var err error
	if read {
		reply, err = dao.doRead(methodname, args...)
	} else {
		reply, err = dao.doWrite(methodname, args...)
	}

	value, err := redis.Int(reply, err)
	if masknil && err == redis.ErrNil {
		return -1, nil
	}

	return value, err
}

func (dao *RedisBaseDao) doFloat(read bool, methodname string, masknil bool, args ...interface{}) (float64, error) {
	var reply interface{}
	var err error
	if read {
		reply, err = dao.doRead(methodname, args...)
	} else {
		reply, err = dao.doWrite(methodname, args...)
	}

	value, err := redis.Float64(reply, err)
	if masknil && err == redis.ErrNil {
		return -1, nil
	}

	return value, err
}

func (dao *RedisBaseDao) doBool(read bool, methodname string, masknil bool, args ...interface{}) (bool, error) {
	var reply interface{}
	var err error
	if read {
		reply, err = dao.doRead(methodname, args...)
	} else {
		reply, err = dao.doWrite(methodname, args...)
	}

	value, err := redis.Bool(reply, err)
	if masknil && err == redis.ErrNil {
		return false, nil
	}

	return value, err
}

func (dao *RedisBaseDao) doString(read bool, methodname string, masknil bool, args ...interface{}) (string, error) {
	var reply interface{}
	var err error
	if read {
		reply, err = dao.doRead(methodname, args...)
	} else {
		reply, err = dao.doWrite(methodname, args...)
	}

	value, err := redis.String(reply, err)
	if masknil && err == redis.ErrNil {
		return "", nil
	}

	return value, err
}

func (dao *RedisBaseDao) doStringSlice(read bool, methodname string, masknil bool, args ...interface{}) ([]string, error) {
	var reply interface{}
	var err error
	if read {
		reply, err = dao.doRead(methodname, args...)
	} else {
		reply, err = dao.doWrite(methodname, args...)
	}

	value, err := redis.Strings(reply, err)
	if masknil && err == redis.ErrNil {
		return []string{}, nil
	}

	return value, err
}

//Key 相关
func (dao *RedisBaseDao) Del(keys ...string) (int, error) {
	args := strSliToInterfSli(keys...)
	return dao.doInt(false, "DEL", false, args...)
}

func (dao *RedisBaseDao) Expire(key string, ttl int) (int, error) {
	return dao.doInt(false, "EXPIRE", false, key, ttl)
}

func (dao *RedisBaseDao) Exists(key string) (bool, error) {
	return dao.doBool(true, "EXISTS", false, key)
}

func (dao *RedisBaseDao) RandomKey() (string, error) {
	//ErrNil时，表示没有key
	return dao.doString(true, "RANDOMKEY", true)
}

func (dao *RedisBaseDao) TTL(key string) (int, error) {
	return dao.doInt(false, "TTL", false, key)
}

//if pattern=="", no regex, if count=0, decide by server
func (dao *RedisBaseDao) Scan(cursor int, pattern string, count int) (int, []string, error) {
	params := make([]interface{}, 0, 6)
	params = append(params, cursor)
	if pattern != "" {
		params = append(params, "MATCH", pattern)
	}
	if count > 0 {
		params = append(params, "COUNT", count)
	}

	reply, err := dao.doRead("SCAN", params...)
	if err != nil {
		return 0, nil, err
	}

	replySlice, _ := reply.([]interface{})
	newCursor, _ := redis.Int(replySlice[0], nil)
	result, _ := redis.Strings(replySlice[1], nil)

	return newCursor, result, nil
}

//string 相关
func (dao *RedisBaseDao) Append(key, val string) (int, error) {
	return dao.doInt(false, "APPEND", false, key, val)
}

func (dao *RedisBaseDao) Get(key string) (string, error) {
	return dao.doString(true, "GET", true, key)
}

func (dao *RedisBaseDao) GetRaw(key string) (string, error) {
	return dao.doString(true, "GET", false, key)
}

func (dao *RedisBaseDao) GetSet(key, val string) (string, error) {
	return dao.doString(false, "GETSET", true, key, val)
}

func (dao *RedisBaseDao) IncrBy(key string, step int) (int, error) {
	return dao.doInt(false, "INCRBY", false, key, step)
}

func (dao *RedisBaseDao) MGet(keys ...string) ([]string, error) {
	args := strSliToInterfSli(keys...)

	return dao.doStringSlice(true, "MGET", false, args...)
}

func (dao *RedisBaseDao) MSet(keyvals ...string) (string, error) {
	args := strSliToInterfSli(keyvals...)
	return dao.doString(false, "MSET", false, args...)
}

func (dao *RedisBaseDao) Set(key, val string, ttl int) (string, error) {
	return dao.SetEPX(key, val, ttl, 0, 0)
}

//ex or px: if ex and px not zeor, only px will work; control by server
//nxxx = 0: for nothing
//nxxx = 1: for NX
//nxxx = 2: for XX
func (dao *RedisBaseDao) SetEPX(key, val string, ex int, px int, nxxx int) (string, error) {
	args := make([]interface{}, 0)
	args = append(args, key, val)
	if ex > 0 {
		args = append(args, "EX", ex)
	}
	if px > 0 {
		args = append(args, "PX", px)
	}
	if nxxx == 1 {
		args = append(args, "NX")
	} else if nxxx == 2 {
		args = append(args, "XX")
	}

	return dao.doString(false, "SET", false, args...)
}

func (dao *RedisBaseDao) SetNx(key, val string) (int, error) {
	return dao.doInt(false, "SETNX", false, key, val)
}

func (dao *RedisBaseDao) StrLen(key string) (int, error) {
	return dao.doInt(true, "STRLEN", false, key)
}

//bit 相关
func (dao *RedisBaseDao) BitCount(key string) (int, error) {
	return dao.doInt(true, "BITCOUNT", false, key)
}

func (dao *RedisBaseDao) BitCountWithPos(key string, start, end int) (int, error) {
	return dao.doInt(true, "BITCOUNT", false, key, start, end)
}

func (dao *RedisBaseDao) BitOp(operation, destkey, key string, key2 ...string) (int, error) {
	params := []string{operation, destkey, key}
	params = append(params, key2...)
	args := strSliToInterfSli(params...)
	return dao.doInt(false, "BITOP", false, args...)
}

func (dao *RedisBaseDao) BitPos(key string, bit int) (int, error) {
	return dao.doInt(true, "BITPOS", false, key, bit)

}
func (dao *RedisBaseDao) BitPosWithPos(key string, bit, start, end int) (int, error) {
	return dao.doInt(true, "BITPOS", false, key, bit, start, end)
}

func (dao *RedisBaseDao) GetBit(key string, offset int) (int, error) {
	return dao.doInt(true, "GETBIT", false, key, offset)
}

func (dao *RedisBaseDao) SetBit(key string, offset, value int) (int, error) {
	return dao.doInt(false, "SETBIT", false, key, offset, value)
}

//hash 相关
func (dao *RedisBaseDao) HDel(key string, fields ...string) (int, error) {
	args := strSliToInterfSliTwo(key, fields...)
	return dao.doInt(false, "HDEL", false, args...)
}

func (dao *RedisBaseDao) HExists(key, field string) (bool, error) {
	return dao.doBool(true, "HEXISTS", false, key, field)
}

func (dao *RedisBaseDao) HGet(key, field string) (string, error) {
	return dao.doString(true, "HGET", true, key, field)
}

func (dao *RedisBaseDao) HGetAll(key string) (map[string]string, error) {
	values, err := dao.doStringSlice(true, "HGETALL", false, key)
	ret := make(map[string]string)
	for i := 0; i < len(values); i += 2 {
		key := values[i]
		value := values[i+1]

		ret[key] = value
	}

	return ret, err
}

func (dao *RedisBaseDao) HIncr(key, field string) (int, error) {
	return dao.doInt(false, "HINCRBY", false, key, field, 1)
}

func (dao *RedisBaseDao) HIncrBy(key, field string, increment int) (int, error) {
	return dao.doInt(false, "HINCRBY", false, key, field, increment)
}

func (dao *RedisBaseDao) HKeys(key string) ([]string, error) {
	return dao.doStringSlice(true, "HKEYS", false, key)
}

func (dao *RedisBaseDao) HLen(key string) (int, error) {
	return dao.doInt(true, "HLEN", false, key)
}

func (dao *RedisBaseDao) HMGet(key string, fields ...string) ([]string, error) {
	args := strSliToInterfSliTwo(key, fields...)
	return dao.doStringSlice(true, "HMGET", false, args...)
}

func (dao *RedisBaseDao) HMSet(key string, fieldvals ...string) (string, error) {
	args := strSliToInterfSliTwo(key, fieldvals...)
	return dao.doString(false, "HMSET", false, args...)
}

func (dao *RedisBaseDao) HSet(key, field, val string) (int, error) {
	return dao.doInt(false, "HSET", false, key, field, val)
}

func (dao *RedisBaseDao) HVals(key string) ([]string, error) {
	return dao.doStringSlice(true, "HVALS", false, key)
}

//if pattern=="", no regex, if count=0, decide by server
func (dao *RedisBaseDao) HScan(key string, cursor int, pattern string, count int) (int, map[string]string, error) {
	params := make([]interface{}, 0, 6)
	params = append(params, key, cursor)
	if pattern != "" {
		params = append(params, "MATCH", pattern)
	}
	if count > 0 {
		params = append(params, "COUNT", count)
	}

	reply, err := dao.doRead("HSCAN", params...)
	if err != nil {
		return 0, nil, err
	}

	replySlice, _ := reply.([]interface{})
	newCursor, _ := redis.Int(replySlice[0], nil)
	result, _ := redis.Strings(replySlice[1], nil)
	ret := make(map[string]string)
	for i := 0; i < len(result); i += 2 {
		key := result[i]
		value := result[i+1]

		ret[key] = value
	}

	return newCursor, ret, nil
}

//list 相关
func (dao *RedisBaseDao) LIndex(key string, index int) (string, error) {
	//ErrNil时，表示对应index不存在
	return dao.doString(true, "LINDEX", true, key, index)
}

func (dao *RedisBaseDao) LInsert(key, op, pivot, val string) (int, error) {
	return dao.doInt(false, "LINSERT", false, key, op, pivot, val)
}

func (dao *RedisBaseDao) LLen(key string) (int, error) {
	return dao.doInt(true, "LLEN", false, key)
}

func (dao *RedisBaseDao) LPop(key string) (string, error) {
	//ErrNil时，表示对应list不存在
	return dao.doString(false, "LPOP", true, key)
}

func (dao *RedisBaseDao) LPush(key string, vals ...string) (int, error) {
	args := strSliToInterfSliTwo(key, vals...)
	return dao.doInt(false, "LPUSH", false, args...)
}

func (dao *RedisBaseDao) LRange(key string, start, stop int) ([]string, error) {
	return dao.doStringSlice(true, "LRANGE", false, key, start, stop)
}

func (dao *RedisBaseDao) LRem(key string, count int, val string) (int, error) {
	return dao.doInt(false, "LREM", false, key, count, val)
}

func (dao *RedisBaseDao) LSet(key string, index int, val string) (string, error) {
	return dao.doString(false, "LSET", false, key, index, val)
}

func (dao *RedisBaseDao) LTrim(key string, start, stop int) (string, error) {
	return dao.doString(false, "LTRIM", false, key, start, stop)
}

func (dao *RedisBaseDao) RPop(key string) (string, error) {
	//ErrNil时，表示对应list不存在
	return dao.doString(false, "RPOP", true, key)
}

func (dao *RedisBaseDao) RPush(key string, vals ...string) (int, error) {
	args := strSliToInterfSliTwo(key, vals...)
	return dao.doInt(false, "RPUSH", false, args...)
}

//set 结构操作
func (dao *RedisBaseDao) SAdd(key string, members ...string) (int, error) {
	args := strSliToInterfSliTwo(key, members...)
	return dao.doInt(false, "SADD", false, args...)
}

func (dao *RedisBaseDao) SCard(key string) (int, error) {
	return dao.doInt(true, "SCARD", false, key)
}

func (dao *RedisBaseDao) SDiff(keys ...string) ([]string, error) {
	args := strSliToInterfSli(keys...)
	return dao.doStringSlice(true, "SDIFF", false, args...)
}

func (dao *RedisBaseDao) SInter(keys ...string) ([]string, error) {
	args := strSliToInterfSli(keys...)
	return dao.doStringSlice(true, "SINTER", false, args...)
}

func (dao *RedisBaseDao) SIsMember(key, member string) (bool, error) {
	return dao.doBool(true, "SISMEMBER", false, key, member)
}

func (dao *RedisBaseDao) SMembers(key string) ([]string, error) {
	return dao.doStringSlice(true, "SMEMBERS", false, key)
}

func (dao *RedisBaseDao) SPop(key string) (string, error) {
	//ErrNil时，表示对应set不存在
	return dao.doString(false, "SPOP", true, key)
}

func (dao *RedisBaseDao) SRandMember(key string) (string, error) {
	//ErrNil时，表示对应set不存在
	return dao.doString(true, "SRANDMEMBER", true, key)
}

func (dao *RedisBaseDao) SRem(key string, members ...string) (int, error) {
	args := strSliToInterfSliTwo(key, members...)
	return dao.doInt(false, "SREM", false, args...)
}

func (dao *RedisBaseDao) SUnion(keys ...string) ([]string, error) {
	args := strSliToInterfSli(keys...)
	return dao.doStringSlice(true, "SUNION", false, args...)
}

//if pattern=="", no regex, if count=0, decide by server
func (dao *RedisBaseDao) SScan(key string, cursor int, pattern string, count int) (int, []string, error) {
	params := make([]interface{}, 0, 6)
	params = append(params, key, cursor)
	if pattern != "" {
		params = append(params, "MATCH", pattern)
	}
	if count > 0 {
		params = append(params, "COUNT", count)
	}

	reply, err := dao.doRead("SSCAN", params...)
	if err != nil {
		return 0, nil, err
	}

	replySlice, _ := reply.([]interface{})
	newCursor, _ := redis.Int(replySlice[0], nil)
	result, _ := redis.Strings(replySlice[1], nil)

	return newCursor, result, nil
}

//zset 相关
func parseScorePair(strAry []string) []*ScorePair {
	ret := make([]*ScorePair, len(strAry)/2)

	for i := 0; i < len(strAry); i = i + 2 {
		member := strAry[i]
		score, err := strconv.ParseInt(strAry[i+1], 10, 64)

		if err == nil {
			ret[i/2] = &ScorePair{member, int(score)}
		}
	}

	return ret
}

func (dao *RedisBaseDao) ZAdd(key string, score int, member string) (int, error) {
	return dao.doInt(false, "ZADD", false, key, score, member)
}

func (dao *RedisBaseDao) ZMAdd(key string, members ...*ScorePair) (int, error) {
	args := make([]interface{}, len(members)*2+1)
	args[0] = key
	for i, member := range members {
		args[2*i+1] = member.Score
		args[2*i+2] = member.Member
	}
	reply, err := dao.doWrite("ZADD", args...)

	return redis.Int(reply, err)
}

func (dao *RedisBaseDao) ZCard(key string) (int, error) {
	return dao.doInt(true, "ZCARD", false, key)
}

//min, max 可以是+inf, -inf
func (dao *RedisBaseDao) ZCount(key string, min, max string) (int, error) {
	return dao.doInt(true, "ZCOUNT", false, key, min, max)
}

func (dao *RedisBaseDao) ZIncrBy(key string, incrNum int, member string) (int, error) {
	return dao.doInt(false, "ZINCRBY", false, key, incrNum, member)
}

func (dao *RedisBaseDao) ZRange(key string, start, stop int) ([]string, error) {
	return dao.doStringSlice(true, "ZRANGE", false, key, start, stop)
}

func (dao *RedisBaseDao) ZRangeWithScores(key string, start, stop int) ([]*ScorePair, error) {
	reply, err := dao.doRead("ZRANGE", key, start, stop, "WITHSCORES")
	values, err := redis.Strings(reply, err)

	return parseScorePair(values), err
}

func (dao *RedisBaseDao) ZRank(key, member string) (int, error) {
	//ErrNil时，表示member不存在
	return dao.doInt(true, "ZRANK", true, key, member)
}

func (dao *RedisBaseDao) ZRem(key string, members ...string) (int, error) {
	args := strSliToInterfSliTwo(key, members...)
	return dao.doInt(false, "ZREM", false, args...)
}

func (dao *RedisBaseDao) ZRemRangeByRank(key string, start, stop int) (int, error) {
	return dao.doInt(false, "ZREMRANGEBYRANK", false, key, start, stop)
}

func (dao *RedisBaseDao) ZRevRange(key string, start, stop int) ([]string, error) {
	return dao.doStringSlice(true, "ZREVRANGE", false, key, start, stop)
}

func (dao *RedisBaseDao) ZRevRangeWithScores(key string, start, stop int) ([]*ScorePair, error) {
	reply, err := dao.doRead("ZREVRANGE", key, start, stop, "WITHSCORES")
	values, err := redis.Strings(reply, err)

	return parseScorePair(values), err
}

func (dao *RedisBaseDao) ZRevRank(key, member string) (int, error) {
	//ErrNil时，表示member不存在
	return dao.doInt(true, "ZREVRANK", true, key, member)
}

func (dao *RedisBaseDao) ZScore(key, member string) (int, error) {
	return dao.doInt(true, "ZSCORE", false, key, member)
}

//lua script相关
func (dao *RedisBaseDao) Script(keyCount int, src string) *LuaScript {
	return &LuaScript{redis.NewScript(keyCount, src), true}
}

func (dao *RedisBaseDao) ScriptWithIfWrite(keyCount int, src string, write bool) *LuaScript {
	return &LuaScript{redis.NewScript(keyCount, src), write}
}

func (dao *RedisBaseDao) LoadScript(script *LuaScript) error {
	var conn redis.Conn
	var err error
	if script.write {
		conn, err = dao.getWrite()
	} else {
		conn, err = dao.getRead()
	}
	if err != nil {
		return err
	}
	defer conn.Close()
	err = script.Load(conn)

	return err
}

func (dao *RedisBaseDao) Eval(script *LuaScript, keysAndArgs ...interface{}) (interface{}, error) {
	var conn redis.Conn
	var err error
	if script.write {
		conn, err = dao.getWrite()
	} else {
		conn, err = dao.getRead()
	}
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	reply, err := script.Do(conn, keysAndArgs...)

	return reply, err
}

//pipeline 相关
func (dao *RedisBaseDao) PipeLine(readonly bool) (*Pipe, error) {
	var conn redis.Conn
	var err error
	if readonly {
		conn, err = dao.getRead()
	} else {
		conn, err = dao.getWrite()
	}
	if err != nil {
		return nil, err
	}

	return &Pipe{conn}, nil
}

func (dao *RedisBaseDao) PipeSend(pipe *Pipe, cmd string, args ...interface{}) error {
	return pipe.Send(cmd, args...)
}

func (dao *RedisBaseDao) PipeExec(pipe *Pipe) (interface{}, error) {
	return pipe.Do("")
}

func (dao *RedisBaseDao) PipeClose(pipe *Pipe) error {
	return pipe.Close()
}

//server 相关
func (dao *RedisBaseDao) BgSave() (string, error) {
	return dao.doString(false, "BGSAVE", false)
}

//pubsub 相关  readtimeout need set to 0
func (dao *RedisBaseDao) Publish(channel, message string) (int, error) {
	return dao.doInt(false, "PUBLISH", false, channel, message)
}

func (dao *RedisBaseDao) Subscribe(channel string) (*SubscribeT, error) {
	conn, err := dao.getWrite()
	if err != nil {
		return nil, err
	}

	psc := redis.PubSubConn{conn}
	err = psc.Subscribe(channel)
	if err != nil {
		conn.Close()
		return nil, err
	}

	st := &SubscribeT{psc, make(chan []byte, 10)}

	go func() {
		defer func() {
			close(st.C)
		}()
		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				st.C <- v.Data
			case redis.Subscription:
				logkit.Infof("redis: %s: %s %d\n", v.Channel, v.Kind, v.Count)
			case error:
				netError, ok := v.(net.Error)
				if ok && netError.Timeout() {
					logkit.Infof("redis: %s", v)
				} else if v.Error() != "redigo: connection closed" {
					logkit.Errorf("redis: %s", v)
				}
				return
			}
		}
	}()

	return st, nil
}

func (dao *RedisBaseDao) SubscribeClose(st *SubscribeT) error {
	return st.conn.Close()
}

//geo 相关
func (dao *RedisBaseDao) GeoAdd(key string, members ...*GeoPair) (int, error) {
	args := make([]interface{}, len(members)*3+1)
	args[0] = key
	for i, member := range members {
		args[3*i+1] = member.Longitude
		args[3*i+2] = member.Latitude
		args[3*i+3] = member.Member
	}
	return dao.doInt(false, "GEOADD", false, args...)
}

//unit: [m|km|mi|ft], if empty then default:m, m:米,km:千米,mi:英里,ft:英尺
func (dao *RedisBaseDao) GeoDist(key, member1, member2, unit string) (float64, error) {
	if unit == "" {
		unit = "m"
	}

	return dao.doFloat(true, "GEODIST", false, key, member1, member2, unit)
}

func (dao *RedisBaseDao) GeoHash(key string, member ...string) ([]string, error) {
	args := make([]interface{}, 0)
	args = append(args, key)
	for _, m := range member {
		args = append(args, m)
	}
	return dao.doStringSlice(true, "GEOHASH", false, args...)
}

func parseGeo(strAry []string) []*Geo {
	ret := make([]*Geo, len(strAry)/2)

	for i := 0; i < len(strAry); i = i + 2 {
		ret[i/2] = &Geo{}
		ret[i/2].Longitude, _ = strconv.ParseFloat(strAry[i], 64)
		ret[i/2].Latitude, _ = strconv.ParseFloat(strAry[i+1], 64)
	}

	return ret
}

func (dao *RedisBaseDao) GeoPos(key string, member ...string) ([]*Geo, error) {
	args := make([]interface{}, 0)
	args = append(args, key)
	for _, m := range member {
		args = append(args, m)
	}
	reply, err := dao.doRead("GEOPOS", args...)
	if err != nil {
		return nil, err
	}
	replySlice := reply.([]interface{})
	values := make([]string, 0)
	for _, r := range replySlice {
		if r == nil {
			values = append(values, "0", "0")
		} else {
			rstr := r.([]interface{})
			rstr_1 := (rstr[0]).([]byte)
			rstr_2 := (rstr[1]).([]byte)
			values = append(values, string(rstr_1), string(rstr_2))
		}
	}

	return parseGeo(values), err
}

func parseGeoPairDistHash(reply interface{}, coord, dist, hash bool) []*GeoPairDistHash {
	replySlice, _ := reply.([]interface{})
	result := make([]*GeoPairDistHash, 0)
	for _, rs := range replySlice {
		gpdh := GeoPairDistHash{}
		if !coord && !dist && !hash {
			member, _ := rs.([]byte)
			gpdh.Member = string(member)
		} else {
			rsSlice, _ := rs.([]interface{})
			pos := 0
			member, _ := rsSlice[pos].([]byte)
			gpdh.Member = string(member)
			pos++
			if dist {
				sdist, _ := rsSlice[pos].([]byte)
				gpdh.Dist, _ = strconv.ParseFloat(string(sdist), 64)
				pos++
			}
			if hash {
				shash, _ := rsSlice[pos].(int64)
				gpdh.Hash = int(shash)
				pos++
			}
			if coord {
				scoord, _ := rsSlice[pos].([]interface{})
				longi, _ := scoord[0].([]byte)
				lati, _ := scoord[1].([]byte)
				gpdh.Longitude, _ = strconv.ParseFloat(string(longi), 64)
				gpdh.Latitude, _ = strconv.ParseFloat(string(lati), 64)
				pos++
			}
		}
		result = append(result, &gpdh)
	}

	return result
}

//unit: [m|km|mi|ft], if empty then default:m, m:米,km:千米,mi:英里,ft:英尺
//count: 0 means no limit
//st: 0-> no sort, 1-> asc, 2-> desc
func (dao *RedisBaseDao) GeoRadius(key string, longitude, latitude, radius float64, unit string, coord, dist, hash bool, count int, st int) ([]*GeoPairDistHash, error) {
	args := make([]interface{}, 0)
	args = append(args, key, longitude, latitude, radius, unit)
	if coord {
		args = append(args, "withcoord")
	}
	if dist {
		args = append(args, "withdist")
	}
	if hash {
		args = append(args, "withhash")
	}
	if count > 0 {
		args = append(args, "count", count)
	}
	if st == 1 {
		args = append(args, "asc")
	} else if st == 2 {
		args = append(args, "desc")
	}
	reply, err := dao.doWrite("GEORADIUS", args...)
	if err != nil {
		return nil, err
	}
	return parseGeoPairDistHash(reply, coord, dist, hash), nil
}

func (dao *RedisBaseDao) GeoRadiusByMember(key, member string, radius float64, unit string, coord, dist, hash bool, count int, st int) ([]*GeoPairDistHash, error) {
	args := make([]interface{}, 0)
	args = append(args, key, member, radius, unit)
	if coord {
		args = append(args, "withcoord")
	}
	if dist {
		args = append(args, "withdist")
	}
	if hash {
		args = append(args, "withhash")
	}
	if count > 0 {
		args = append(args, "count", count)
	}
	if st == 1 {
		args = append(args, "asc")
	} else if st == 2 {
		args = append(args, "desc")
	}
	reply, err := dao.doWrite("GEORADIUSBYMEMBER", args...)
	if err != nil {
		return nil, err
	}
	return parseGeoPairDistHash(reply, coord, dist, hash), nil
}

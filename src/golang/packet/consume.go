//这个没用了。
/**
  依赖 res.yaml, 已删：
      #first_charge
      FIRST_CHARGE_GOODS : "1,2,131"
      FIRST_CHARGE_NUM : "10,50,1"

*/
package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"git.pandatv.com/panda-public/janna_dao/etcd/mysql"
	"git.pandatv.com/panda-public/kafka-go"
	"git.pandatv.com/panda-web/gobase/http"
	"git.pandatv.com/panda-web/gobase/id"
	"git.pandatv.com/panda-web/gobase/log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Packet struct {
	Rid       int
	Packlimit int
}
type Result struct {
	Errno  int         `json:"errno"`
	Errmsg string      `json:"errmsg"`
	Data   interface{} `json:"data"`
}

const (
	TABLE         = "first_charge"
	STATUS_INIT   = "0" // 首充初始化状态
	STATUS_FINISH = "1" // 礼包返回完成状态
	STATUS_FAIL   = "4" // 礼包返回失败
	APP           = "pandaren"
	CALLER        = "first_charge"
)

// goodsId 1=>烤鱼 2=>饭团 4=>彩色弹幕卡
var FirstChargeGoods = []string{}

// num
var FirstChargeGoodsNum = []string{}

var (
	env = os.Getenv("ENV")
)

var MysqlClient *mysql.MysqlBaseDao

func init() {
}

// qa test
// var qaTest = []string{"27400276", "83537972", "39933722"}

func main() {
	defer func() {
		if v := recover(); v != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			//因为换行会被golang字符串截断，所以采用|#|替换\n
			lines := bytes.Split(buf[:n], []byte{'\n'})
			var stack string
			for _, v := range lines {
				tmp := string(bytes.TrimSpace(v))
				if tmp != "" {
					stack = stack + tmp + "|#|"
				}
			}
			logkit.Logger.Error(fmt.Sprintf("Error[%s] recover panic: %s stack:%s\n", v, stack))
		}
	}()
	var err error
	logkit.Logger.Init("bag_consume_packet", logkit.LevelInfo)

	logkit.Logger.Infof("env:%s", os.Getenv("ENV"))

	logkit.Logger.Info("mysql_host: " + os.Getenv("DB_USER") + " pwd: " + os.Getenv("DB_PWD") + " db_name: " + os.Getenv("DB_NAME") + " janna_mysql_endpoints: " + os.Getenv("MYSQL_ENDPOINTS") + " janna_mysql_callname: " + os.Getenv("MYSQL_CALLNAME") + " janna_mysql_target: " + os.Getenv("MYSQL_TARGET"))

	MysqlClient, err = mysql.NewMysqlBaseDao(os.Getenv("MYSQL_ENDPOINTS"), os.Getenv("MYSQL_CALLNAME"), os.Getenv("MYSQL_TARGET"), os.Getenv("DB_USER"), os.Getenv("DB_PWD"), os.Getenv("DB_NAME"), os.Getenv("DB_USER"), os.Getenv("DB_PWD"), os.Getenv("DB_NAME"))

	FirstChargeGoods = strings.Split(os.Getenv("FIRST_CHARGE_GOODS"), ",")
	FirstChargeGoodsNum = strings.Split(os.Getenv("FIRST_CHARGE_NUM"), ",")

	logkit.Logger.Infof("goods:%s num:%s", FirstChargeGoods, FirstChargeGoodsNum)

	if err != nil {
		logkit.Logger.Errorf("mysql conn fail err=%s", err.Error())
		panic(err)
	}

	// new consumer
	// zookeeper: kafka对应zk地址
	// group: kafka消费者分组
	// topics: 消费的话题，支持多个topic
	// resetOffsets: 是否重头开始消费
	logkit.Logger.Infof("zookeeper:%s", os.Getenv("BAG_PICKET_ZOOKEEPER"))
	logkit.Logger.Infof("topic:%s", os.Getenv("BAG_PICKET_TOPIC"))
	consumer, err := kafka.NewKafkaConsumer(os.Getenv("BAG_PICKET_ZOOKEEPER"), "bag", []string{os.Getenv("BAG_PICKET_TOPIC")}, kafka.ConfigResetOffset(false))
	if err != nil {
		fmt.Println(err)
		return
	}

	// handler err
	go func() {
		for err := range consumer.Errors() {
			logkit.Logger.Errorf("consumer err=%s", err.Error())
		}
	}()

	// read data
	for data := range consumer.Datas() {
		logkit.Logger.Infof("execute data=%s", string(data.Value))
		execute(data.Value)
	}

	logkit.Wait(func() {
		consumer.Close()
	})
}

func execute(data []byte) {
	var err error
	var packet = new(Packet)
	err = json.Unmarshal(data, &packet)
	if err != nil {
		logkit.Logger.Errorf("json decode fail data=%s err=%s", data, err.Error())
		return
	}
	var args = make([]interface{}, 0)
	args = append(args, packet.Rid)
	args = append(args, packet.Packlimit)
	selSql := "SELECT count(1) as total FROM " + TABLE + " WHERE uid = ? and packlimit = ?"
	total, err := MysqlClient.FetchRow(selSql, args...)
	if err != nil {
		logkit.Logger.Errorf("mysql select fail sql=%s args=%s err=%s", selSql, args, err.Error())
		return
	}
	itotal, _ := strconv.Atoi(total["total"])
	if itotal > 0 {
		logkit.Logger.Warnf("consume retry total=%d packet=%#v", itotal, packet)
		return
	}
	// 存mysql
	insertSql := "INSERT INTO " + TABLE + " (createtime, updatetime, uid, packlimit, status) VALUES (?, ?, ?, ?, ?)"
	now := time.Now().Format("2006-01-02 15:04:05")
	args = make([]interface{}, 0)
	args = append(args, now)
	args = append(args, now)
	args = append(args, packet.Rid)
	args = append(args, packet.Packlimit)
	args = append(args, STATUS_INIT)

	logkit.Logger.Infof("insert sqlstr: %s args: %s", insertSql, args)
	id, err := MysqlClient.Insert(insertSql, args...)
	if err != nil {
		logkit.Logger.Errorf("mysql insert fail sql=%s args=%s err=%s", insertSql, args, err.Error())
		return
	}

	var succIds []string
	uid := strconv.Itoa(packet.Rid)
	var retry int
	for index, goodsId := range FirstChargeGoods {
		// 重置重试次数
		retry = 0
		// 生成uuid
		uuid := genGuuid(uid)

		num := FirstChargeGoodsNum[index]
		// 首充逻辑, 存背包物品
		params := url.Values{}
		params.Add("app", APP)
		params.Add("uid", uid)
		params.Add("goods_id", goodsId)
		params.Add("num", num)
		params.Add("uuid", uuid)
		params.Add("_caller", CALLER)
		// 自定义超时时间10秒
		c := &http.Client{
			Timeout: time.Second * 5,
		}
		cli := httpclient.NewClient(c)
		uri := getIbagUri()
		var ret Result
		err = cli.PostAsJson(uri, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()), &ret)
		// 增加重试机制
		if err != nil {
			logkit.Logger.Infof("curl ibag fail goodsId=%s uid=%s num=%s app=%s uuid=%s retry=%d err=%s", goodsId, uid, num, APP, uuid, retry, err.Error())
			for retry < 3 {
				err = cli.PostAsJson(uri, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()), &ret)
				if err == nil {
					break
				}
				logkit.Logger.Infof("curl ibag fail goodsId=%s uid=%s num=%s app=%s uuid=%s retry=%d err=%s", goodsId, uid, num, APP, uuid, retry, err.Error())
				retry++
				time.Sleep(500 * time.Millisecond)
			}
		}

		if retry >= 3 {
			logkit.Logger.Errorf("curl ibag fail goodsId=%s uid=%s num=%s app=%s uuid=%s retry=%d err=%s", goodsId, uid, num, APP, uuid, retry, err.Error())
			continue
		}

		if ret.Errno != 0 && ret.Errno == 1001 {
			logkit.Logger.Infof("curl ibag return fail goodsId=%s uid=%s num=%s app=%s uuid=%s retry=%d ret=%#v", goodsId, uid, num, APP, uuid, retry, ret)
			// 再试一次
			time.Sleep(500 * time.Millisecond)
			err = cli.PostAsJson(uri, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()), &ret)
			if err != nil {
				logkit.Logger.Errorf("curl ibag fail goodsId=%s uid=%s num=%s app=%s uuid=%s retry=%d err=%s", goodsId, uid, num, APP, uuid, retry, err.Error())
			}
			if ret.Errno != 0 {
				logkit.Logger.Errorf("curl ibag return fail goodsId=%s uid=%s num=%s app=%s uuid=%s retry=%d ret=%#v", goodsId, uid, num, APP, uuid, retry, ret)
				continue
			}
		}

		logkit.Logger.Infof("curl ibag success goodsId=%s uid=%s num=%s app=%s uuid=%s retry=%d ret=%#v", goodsId, uid, num, APP, uuid, retry, ret)
		succIds = append(succIds, goodsId)
	}

	args = make([]interface{}, 0)
	if len(succIds) == 0 {
		args = append(args, STATUS_FAIL)
	} else {
		args = append(args, STATUS_FINISH)
	}
	args = append(args, strings.Join(succIds, ","))
	args = append(args, id)
	updateSql := "UPDATE " + TABLE + " SET status = ?, data = ? WHERE id = ?"
	logkit.Logger.Infof("update sqlstr: %s args: %s", updateSql, args)
	_, err = MysqlClient.Update(updateSql, args...)
	if err != nil {
		logkit.Logger.Errorf("mysql update fail sql=%s err=%s", updateSql, err.Error())
		return
	}

	// 发系统消息
	if len(succIds) > 0 {
		sendMsg(uid)
	}
}

func sendMsg(uid string) {
	var err error
	now := time.Now().Format("2006-01-02")
	content := `亲爱的熊猫直播用户：
    您于近期完成了第一次单笔达到 10 元的充值。为感谢您的支持，平台赠送给您价值 60 元的烤鱼和饭团，以及 ` + FirstChargeGoodsNum[2] + ` 张彩色弹幕卡，有效期 7 天。您可以在直播间的背包中查看和使用。快去送给您喜欢的主播吧！
    熊猫直播
    ` + now + `
    `
	fmt.Println(content)
	params := url.Values{}
	params.Add("title", "获得首次充值礼包通知")
	params.Add("cat", "5")
	params.Add("to_uid", uid)
	params.Add("content", content)
	params.Add("_caller", CALLER)
	// 自定义超时时间5秒
	c := &http.Client{
		Timeout: time.Second * 5,
	}
	cli := httpclient.NewClient(c)
	uri := getMessUri()
	var ret Result
	err = cli.PostAsJson(uri, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()), &ret)
	if err != nil {
		logkit.Logger.Infof("curl message fail params=%#v err=%s", params, err.Error())
	}
	if ret.Errno != 0 {
		logkit.Logger.Infof("curl message return fail params=%#v ret=%#v", params, ret)
	}

	logkit.Logger.Infof("curl message success params=%#v ret=%#v", params, ret)

}

func getIbagUri() string {
	prefix := "http://"
	url := "ibag.pdtv.io:8360/bag/add"
	if env == "online" {
		return prefix + url
	}
	if env == "dev" {
		return prefix + os.Getenv("USER") + "." + url
	}
	return prefix + "beta." + url
}

func getMessUri() string {
	return "http://message.pdtv.io:8360/Message/sendMessageToUser"
}

func StrSliceContains(strSlice []string, searchStr string) bool {
	for _, val := range strSlice {
		if val == searchStr {
			return true
		}
	}
	return false
}

func genGuuid(uid string) string {
	guuid, err := id.NewGuuid(os.Getenv("GUUID_IDC"), os.Getenv("GUUID_ENV"))
	if err != nil {
		logkit.Logger.Warnf("guuid fail uid=%s err=%s", uid, err.Error())
		return genSelfGuuid(uid)
	}
	id, err := guuid.Get()
	if err != nil {
		return genSelfGuuid(uid)
	}
	return strconv.FormatUint(id, 10)
}

// uid + 时间戳生成唯一id
func genSelfGuuid(uid string) string {
	now := time.Now().UnixNano()
	sNow := strconv.FormatInt(now, 10)
	key := uid + sNow
	res := Md5(key)
	return fmt.Sprintf("%x", res)
}

func Md5(str string) [16]byte {
	return md5.Sum([]byte(str))
}

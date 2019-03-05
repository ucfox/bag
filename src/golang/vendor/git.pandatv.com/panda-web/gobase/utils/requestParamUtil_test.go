package utils

import (
	"encoding/json"
	"net/url"
	"testing"

	"git.pandatv.com/panda-web/gobase/log"
)

func TestParseParamFromPhp(t *testing.T) {
	logkit.Logger.Init("request_util_test", logkit.LevelDebug)
	str := make(url.Values)
	str.Add("content[19]", "2:0")
	str.Add("content[58]", "2:0")
	str.Add("content[26]", "2:1")
	str.Add("content[10]", "2:0")
	str.Add("content[16]", "2:1")
	str.Add("content[2]", "1:2")
	str.Add("content[20]", "0:2")
	str.Add("content[15]", "2:0")
	str.Add("content[30]", "2:1")
	str.Add("content[50]", "0:0")
	str.Add("content[7]", "2:1")
	str.Add("content[14]", "2:1")
	str.Add("content[57]", "0:0")
	str.Add("content[52]", "0:0")
	str.Add("content[1]", "2:1")
	str.Add("content[5]", "2:0")
	str.Add("content[38]", "2:0")
	str.Add("content[45]", "2:0")
	str.Add("content[3]", "2:2")
	str.Add("content[17]", "2:1")
	str.Add("content[8]", "0:2")
	str.Add("content[9]", "2:0")
	str.Add("content[31]", "2:1")
	str.Add("content[21]", "2:0")
	str.Add("content[33]", "1:2")
	str.Add("content[54]", "0:0")
	str.Add("content[47]", "0:2")
	str.Add("content[37]", "0:2")
	str.Add("content[42]", "0:2")
	str.Add("content[55]", "0:0")
	str.Add("content[29]", "2:0")
	str.Add("content[46]", "0:2")
	str.Add("content[48]", "2:1")
	str.Add("content[23]", "2:0")
	str.Add("content[53]", "0:0")
	str.Add("content[27]", "2:0")
	str.Add("content[11]", "2:0")
	str.Add("content[4]", "0:2")
	str.Add("content[39]", "2:1")
	str.Add("content[6]", "2:0")
	str.Add("content[41]", "1:2")
	str.Add("content[40]", "1:2")
	str.Add("content[43]", "1:2")
	str.Add("content[25]", "2:0")
	str.Add("content[32]", "1:2")
	str.Add("content[35]", "2:0")
	str.Add("content[36]", "0:2")
	str.Add("content[12]", "2:1")
	str.Add("content[34]", "0:2")
	str.Add("content[28]", "2:0")
	str.Add("content[51]", "0:0")
	str.Add("content[13]", "2:0")
	str.Add("content[18]", "2:0")
	str.Add("content[44]", "0:2")
	str.Add("content[56]", "0:0")
	t.Log(len(str))
	v := ParseJsonParamFromPHP(str, "content")
	str.Add("from[_plat]", "_web")
	str.Add("to[toroom]", "2223435")
	data, _ := json.Marshal(v)
	v2 := ParseJsonParamFromPHP(str, "from")
	t.Logf("%s", string(data))
	d2, _ := json.Marshal(v2)
	logkit.Debugf("%s", data)
	t.Logf("%s", d2)
	cont := make(url.Values)
	cont.Add("content", "杀啥啊是")

}
func BenchmarkParsePhp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		str := url.Values{"from[key1][key2][key3]": {"v1"}, "from[key2][key3]": {"v2"}, "from[key1][key4]": {"v3"}}

		ParseJsonParamFromPHP(str, "from")
	}
}
func TestParseJson(t *testing.T) {
	logkit.Logger.Init("request_util_test", logkit.LevelDebug)
	str := "{\"k1\":\"v1\",\"k2\":\"v2\",\"k3\":[2,3,4]}"
	var js interface{}
	json.Unmarshal([]byte(str), &js)
	logkit.Debugf("1----%s", js)

	str = "[2,4,5,{\"k\":\"v\"}]"
	json.Unmarshal([]byte(str), &js)
	logkit.Debugf("2---%s", js)
}

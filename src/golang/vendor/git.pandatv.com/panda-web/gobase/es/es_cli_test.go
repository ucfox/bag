package es

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestQuery(t *testing.T) {
	client, _ := NewESClient("10.110.16.163:9200")
	index := "logstash-kafka-riven-gateway-log-2016.12.23"

	res, _ := client.NewSearch().Index(index).Query("123456789").From(0).Size(10).Do()
	r, _ := json.Marshal(res)
	t.Logf("%s", r)
	res, _ = client.NewSearch().Index(index).Query("-fields.alived:false,|action:bind").Field("fields.sid,fields.bind_time,fields.app,fields.roomid,fields.uid").From(0).Size(10).Do()
	r, _ = json.Marshal(res)
	t.Logf("%s", r)

}

func TestRangeQuery(t *testing.T) {
	client, _ := NewESClient("10.110.16.163:9200")
	index := "logstash-kafka-riven-gateway-log-2016.12.23"
	res, _ := client.NewSearch().Index(index).Query("<>fields.uid:>0").From(0).Size(10).Do()
	r, _ := json.Marshal(res)
	t.Logf("%s", r)
}
func TestGroupByQuery(t *testing.T) {
	client, _ := NewESClient("10.110.16.163:9200")
	index := "logstash-kafka-riven-gateway-log-2016.12.23"

	res, _ := client.NewSearch().Index(index).GroupBy("fields.gw_id.raw", "", 10).GroupBy("fields.roomid", "", 10).Do()
	r, _ := json.Marshal(res)
	t.Logf("%s", r)

}
func TestJsonMarshal(t *testing.T) {
	var flag = `{"d":{"a":"87702850d4926407","b":1,"c":143168533585583274},"e":""}`

	var d interface{}
	json.Unmarshal([]byte(flag), &d)
	fmt.Printf("11111111%s \n", d)

	decode := json.NewDecoder(strings.NewReader(flag))
	decode.UseNumber()

	if err := decode.Decode(&d); err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%v \n", d)
}

func TestRemoveEnation(t *testing.T) {
	b := []byte(`{"id":1.2222223E8}`)
	var v map[string]interface{}
	decode := json.NewDecoder(bytes.NewReader(b))
	decode.UseNumber()
	decode.Decode(&v)
	fmt.Printf("%v", v)
	fmt.Printf("%v", 122222222222222222)
}
func TestESQuery(t *testing.T) {
	url := "http://10.110.16.163:9200/logstash-kafka-riven-gateway-log-2016.09.18/_search"
	req, _ := http.NewRequest("POST", url, strings.NewReader(`{ "query": { "bool": { "must_not": [ { "term": { "fields.alived": false  }  }  ]  }  }  }`))
	t.Log(req)
}

func TestDateHistogram(t *testing.T) {
	client, _ := NewESClient("10.110.16.163:9200")
	index := "logstash-kafka-riven-gateway-log-2016.12.23"

	res, _ := client.NewSearch().Index(index).DateHistogram("fields.bind_time", "5m", "").DateHistogram("fields.unbind_time", "5m", "").Do()
	r, _ := json.Marshal(res)
	t.Logf("%s", r)

}

func TestDateHistogramGroupBy(t *testing.T) {
	client, _ := NewESClient("10.110.16.163:9200")
	index := "logstash-kafka-riven-gateway-log-2017.03.14"

	res, _ := client.NewSearch().Index(index).Query("-fields.alived:false,action:bind,<>fields.bind_time:<1489489200000").DateHistogram("fields.bind_time", "5m", "").GroupBy("fields.idc_id.raw", "", 100).Do()
	r, _ := json.Marshal(res)
	t.Logf("%s", r)

}

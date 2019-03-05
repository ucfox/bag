package es

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestUpdate(t *testing.T) {
	client, _ := NewESClient("10.110.16.163:9200")
	index := "logstash-kafka-riven-gateway-log-2016.12.23"
	id := "bjac-t20v-9049274144696247"

	doc := map[string]interface{}{"action": "bind", "fields": map[string]interface{}{"sid": id, "bind_time": "2016.12.23T11:31:47.824Z", "unbind_time": "2016.12.23T14:04:24.271Z", "alived": false}}
	result, err := client.NewUpdate().UpdateDoc(index, "center_log", id, doc).Do()

	if err != nil || !result.Success {
		fmt.Printf("update result %s, error %s", result, err)
	}
}
func TestBatchUpdate(t *testing.T) {
	client, _ := NewESClient("10.110.16.163:9200")
	index := "logstash-kafka-riven-gateway-log-2016.12.23"
	id := "bjac-t20v-9058840758236510"

	doc := map[string]interface{}{"action": "bind", "fields": map[string]interface{}{"sid": id, "bind_time": "2017-03-17T10:12:40.128+0800", "unbind_time": "2017-03-17T10:12:40.128+0800", "alived": false}}

	result, err := client.NewUpdate().UpdateDoc(index, "center_log", id, doc).Do()
	if err != nil || !result.Success {
		d, _ := json.Marshal(result)
		t.Fatalf("update result %s, error %s", d, err)
	}

}

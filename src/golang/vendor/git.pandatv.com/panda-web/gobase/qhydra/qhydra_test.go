package qhydra

import (
	"encoding/json"
	"testing"
)

const QHYDRA_EVENT_CLASSIFY = "villa_classify_update"

func Test_Qhydra(t *testing.T) {
	q, err := NewQhydra(QHYDRA_EVENT_CLASSIFY)
	if err != nil {
		t.Error(err)
	}

	eventData := make(map[string]interface{})
	eventData["roomid"] = "15931"

	oldc := make(map[string]string)
	oldc["ename"] = "lol"
	oldc["cname"] = "英雄联盟"

	newc := make(map[string]string)
	newc["ename"] = "zhuji"
	newc["cname"] = "主机游戏"

	eventData["old_classify"] = oldc
	eventData["new_classify"] = newc
	eventDataEncode, _ := json.Marshal(eventData)

	q.Trigger(QHYDRA_EVENT_CLASSIFY, "", eventDataEncode)
}

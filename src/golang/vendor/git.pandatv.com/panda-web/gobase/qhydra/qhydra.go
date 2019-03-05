package qhydra

import (
	"encoding/json"
	"log/syslog"
	"os"
	"time"
)

// 在业务调用中，请使用单例
const (
	EVENT_PREFIX = "pcgameq_"
)

type Qhydra struct {
	w *syslog.Writer
}

func NewQhydra(event string) (*Qhydra, error) {
	qhydra := new(Qhydra)
	event = EVENT_PREFIX + event
	writer, err := syslog.New(syslog.LOG_INFO|syslog.LOG_LOCAL4, event)
	if err != nil {
		return nil, err
	}
	qhydra.w = writer
	return qhydra, err
}

func (this *Qhydra) Trigger(event string, key string, data []byte) error {
	msg := make(map[string]string) //这里用map[string]interface{} data域会有问题
	msg["name"] = event
	msg["data"] = string(data)
	var hostname string
	hostname, _ = os.Hostname()
	msg["host"] = hostname
	msg["key"] = key
	msg["time"] = time.Now().Format("2006-01-02 15:04:05")
	msgJson, _ := json.Marshal(msg)

	err := this.w.Info(string(msgJson))
	if err != nil {
		return err
	}

	return err
}

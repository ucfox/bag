package logkit

import (
	"fmt"
	"log"
	"testing"
)

//检查/data/projlogs/logkit_test/下是否有相应的日志
func Test_Filelog(t *testing.T) {
	InitFileLog("logkit_test", LevelDebug, "./")
	Logger.Info("test info")
	Info("test info")
	Logger.Error("test error")
	Error("test error")
	Logger.Debug("test debug")
	Debug("test debug")
	Logger.Warn("test warn")
	Action("test", Fields{"a": "a", "b": 1})
	Actions("test", "{\"a\":1}")
	Actions("test", "hello")
	Warn("test warn")
	log.Printf("it's a log")
	log.Printf("[D]it's a debug")
	log.Printf("[I]it's a info")
	log.Printf("[W]it's a warn")
	log.Printf("[E]it's a error")
	Exit()
}

func Test_Wait(t *testing.T) {
	InitFileLog("logkit_test", LevelDebug, "./")
	Logger.Info("test wait")
	Info("test wait")
	Logger.Error("test wait")
	Error("test wait")
	Logger.Debug("test wait")
	Debug("test wait")
	Logger.Warn("test wait")
	Action("wait", Fields{"a": "a", "b": 1})
	Actions("test", "{\"a\":1}")
	Warn("test wait")
	Wait(func() {
		fmt.Println("wait>>>>")
	})
}

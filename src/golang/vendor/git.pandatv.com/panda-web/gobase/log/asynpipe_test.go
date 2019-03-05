package logkit

import (
	"testing"
)

//检查/data/projlogs/logkit_test/下是否有相应的日志
func Test_AsynPipe(t *testing.T) {
	InitAsynPipeLog("/data/projlogs/golangloger", "logkit_syslog", LevelDebug, 50000, 20)
	Logger.Info("test info")
	Info("test info")
	Logger.Error("test error")
	Error("test error")
	Logger.Debug("test debug")
	Debug("test debug")
	Logger.Warn("test warn")
	Warn("test warn")
	ch := make(chan int, 10)
	<-ch
}

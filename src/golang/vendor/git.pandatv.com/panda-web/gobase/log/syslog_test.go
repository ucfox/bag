package logkit

import (
	"testing"
)

//检查/data/projlogs/logkit_test/下是否有相应的日志
func Test_Syslog(t *testing.T) {
	Logger.Init("logkit_syslog", LevelDebug)
	Logger.Info("test info")
	Info("test info")
	Logger.Error("test error")
	Error("test error")
	Logger.Debug("test debug")
	Debug("test debug")
	Logger.Warn("test warn")
	Warn("test warn")
}

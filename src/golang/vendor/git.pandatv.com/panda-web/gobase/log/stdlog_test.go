package logkit

import "testing"

func TestStdlog(t *testing.T) {
	log := NewStdLog(LevelError, "test")
	log.Print("hello")
}

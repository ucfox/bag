package logkit

import (
	"fmt"
	"log/syslog"
	"sync"
)

type syslogWriter struct {
	logName string
	mu      sync.Mutex
	writers [6]*syslog.Writer
}

func newSyslogWriter(logName string) *syslogWriter {
	return &syslogWriter{
		logName: logName,
	}
}

func (w *syslogWriter) write(level Level, s string) {
	writer := w.writers[level]
	if writer == nil {
		writer = w.initWriter(level)
	}
	if writer != nil {
		writer.Write([]byte(" [" + level.String() + "] " + s))
	}
}

func (w *syslogWriter) initWriter(level Level) *syslog.Writer {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.writers[level] != nil {
		return w.writers[level]
	}
	writer, err := syslog.New(transToSyslogLevel(level)|syslog.LOG_LOCAL6, w.logName)
	if err != nil {
		fmt.Printf("new syslog writer err:%s\n", err)
		return nil
	}
	w.writers[level] = writer
	return writer
}

func (w *syslogWriter) exit() {
	for _, writer := range w.writers {
		if writer != nil {
			writer.Close()
		}
	}
}

func transToSyslogLevel(level Level) syslog.Priority {
	return syslog.Priority(8 - level)
}

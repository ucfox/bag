package logkit

import (
	"bufio"
	"fmt"
	"log/syslog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const severityMask = 0x07
const facilityMask = 0xf8

type pipeLogWriter struct {
	logName string
	pipe    *os.File
	writer  *bufio.Writer
	mu      sync.Mutex
}

func (w *pipeLogWriter) write(level Level, s string) {
	s = strings.Replace(s, "\n", "", -1)
	s = strconv.QuoteToASCII(s)
	nl := ""
	if !strings.HasSuffix(s, "\n") {
		nl = "\n"
	}
	timestamp := time.Now().Format(time.Stamp)
	syslogLevel := transToSyslogLevel(level)
	priority := ((syslogLevel | syslog.LOG_LOCAL6) & facilityMask) | (syslogLevel & severityMask)
	str := fmt.Sprintf("<%d>%s %s[%d]: [%s] %s%s", priority, timestamp, w.logName, os.Getpid(), level.String(), s[1:len(s)-1], nl)
	w.mu.Lock()
	defer w.mu.Unlock()
	n, err := w.writer.Write([]byte(str))
	if n < len(str) || err != nil {
		fmt.Printf("write pipe file: %d, %s\n", n, err.Error())
	}
}

func (w *pipeLogWriter) exit() {
	w.writer.Flush()
	w.pipe.Close()
}

func (w *pipeLogWriter) flushDaemon() {
	for _ = range time.NewTicker(flushInterval).C {
		w.flushAll()
	}
}

func (w *pipeLogWriter) flushAll() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.writer.Flush()
}

func newPipelogWriter(logName, pipePath string) *pipeLogWriter {
	if _, err := os.Stat(pipePath); os.IsNotExist(err) {
		dir := filepath.Dir(pipePath)
		os.MkdirAll(dir, 0777)
		syscall.Mkfifo(pipePath, 0777)
	}
	pipe, err := os.OpenFile(pipePath, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		fmt.Printf("new pipe writer err:%v\n", err)
		return nil
	}
	bw := bufio.NewWriterSize(pipe, 1024)
	w := &pipeLogWriter{
		logName: logName,
		writer:  bw,
		pipe:    pipe,
	}
	go w.flushDaemon()
	return w
}

package logkit

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const SeverityMask = 0x07
const FacilityMask = 0xf8

const (
	// Severity.

	// From /usr/include/sys/syslog.h.
	// These are the same on Linux, BSD, and OS X.
	LOG_EMERG Priority = iota
	LOG_ALERT
	LOG_CRIT
	LOG_ERR
	LOG_WARNING
	LOG_NOTICE
	LOG_INFO
	LOG_DEBUG
)

const (
	// Facility.

	// From /usr/include/sys/syslog.h.
	// These are the same up to LOG_FTP on Linux, BSD, and OS X.
	LOG_KERN Priority = iota << 3
	LOG_USER
	LOG_MAIL
	LOG_DAEMON
	LOG_AUTH
	LOG_SYSLOG
	LOG_LPR
	LOG_NEWS
	LOG_UUCP
	LOG_CRON
	LOG_AUTHPRIV
	LOG_FTP
	_ // unused
	_ // unused
	_ // unused
	_ // unused
	LOG_LOCAL0
	LOG_LOCAL1
	LOG_LOCAL2
	LOG_LOCAL3
	LOG_LOCAL4
	LOG_LOCAL5
	LOG_LOCAL6
	LOG_LOCAL7
)

type Priority int

type asynPipeWriter struct {
	Prior   Priority
	LogName string
	Wfd     int
	Rfd     int
	Queue   chan *string
	Exit    bool
}

func (w *asynPipeWriter) write(level Level, s string) {
	if w.Wfd != 0 {
		s = strings.Replace(s, "\n", "", -1)
		nl := ""
		if !strings.HasSuffix(s, "\n") {
			nl = "\n"
		}
		msg := "[asynpipe] [" + level.String() + "] " + s + nl
		w.Queue <- &msg
	}
}

func (w *asynPipeWriter) exit() {
	w.Exit = true
	close(w.Queue)
}

func newAsynPipeWriter(pipePath string, logName string, concurrency int, logNumPerRequest int) *asynPipeWriter {
	if _, err := os.Stat(pipePath); os.IsNotExist(err) {
		dir := filepath.Dir(pipePath)
		os.MkdirAll(dir, 0777)
		syscall.Mkfifo(pipePath, 0777)
	}
	queue := make(chan *string, concurrency*logNumPerRequest)
	// fifo的bug
	rfd, err := syscall.Open(pipePath, syscall.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		fmt.Printf("new pipe writer err:%v\n", err)
		return nil
	}
	wfd, err := syscall.Open(pipePath, syscall.O_WRONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		fmt.Printf("new pipe writer err:%v\n", err)
		return nil
	}
	ret := &asynPipeWriter{LOG_INFO | LOG_LOCAL6, logName, wfd, rfd, queue, false}
	go AsynPipeWrite(ret)
	return ret
}

func AsynPipeWrite(w *asynPipeWriter) {
	defer func() {
		if v := recover(); v != nil {
			fd, _ := os.OpenFile("/tmp/pipe.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			//因为换行会被golang字符串截断，所以采用|#|替换\n
			lines := bytes.Split(buf[:n], []byte{'\n'})
			var stack string
			for _, v := range lines {
				tmp := string(bytes.TrimSpace(v))
				if tmp != "" {
					stack = stack + tmp + "|#|"
				}
			}
			err := fmt.Sprintf("Error recover panic: %s stack:%s\n", v, stack)
			fd.Write([]byte(err))
			fd.Close()
			go AsynPipeWrite(w)
		}
	}()

	var msg *string
	priority := (w.Prior & FacilityMask) | (LOG_INFO & SeverityMask)
	for {
		if w.Exit {
			for msg := range w.Queue {
				timestamp := time.Now().Format(time.Stamp)
				str := fmt.Sprintf("<%d>%s %s[%d]: %s%s", priority, timestamp, w.LogName, os.Getpid(), *msg, "")
				syscall.Write(w.Wfd, []byte(str))
			}
			syscall.Close(w.Wfd)
			syscall.Close(w.Rfd)
			return
		}
		msg = <-w.Queue
		timestamp := time.Now().Format(time.Stamp)
		str := fmt.Sprintf("<%d>%s %s[%d]: %s%s", priority, timestamp, w.LogName, os.Getpid(), *msg, "")
		byteStr := []byte(str)
		for {
			m, err := syscall.Write(w.Wfd, byteStr)
			if m > 0 && m < len(byteStr) || err == syscall.EINTR {
				byteStr = byteStr[m:]
				continue
			}
			break
		}
	}
}

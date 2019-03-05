package logkit

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var Logger *XLogger
var inited bool

type Writer interface {
	write(Level, string)
	exit()
}

type Level byte

func (l *Level) Set(s string) error {
	for k, v := range levelName {
		level := Level(k)
		if level != levelDefault && v == s {
			*l = level
			return nil
		}
	}
	return errors.New("invaild level")
}

func (l *Level) String() string {
	return levelName[*l]
}

const (
	logTypeSyslog   int = 0
	logTypeFilelog  int = 1
	logTypePipelog  int = 2
	logTypeAsynPipe int = 3
)

const (
	levelDefault Level = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelAction
)

const (
	Concurrency      = 50000
	LogNumPerRequest = 20
)

var levelName = []string{
	levelDefault: "output",
	LevelDebug:   "debug",
	LevelInfo:    "info",
	LevelWarn:    "warn",
	LevelError:   "error",
	LevelAction:  "action",
}

var (
	flagAlseStdout bool
	flagLoglevel   Level
	flagLogname    string
	flagLogpath    string
)

func init() {
	Logger = &XLogger{
		logType:          logTypeSyslog,
		logName:          "logkit",
		logLevel:         LevelDebug,
		logPath:          "/data/projlogs",
		alsoStdout:       false,
		concurrency:      Concurrency,
		logNumPerRequest: LogNumPerRequest,
		pipePath:         "/tmp/golanglogger",
	}
	flag.BoolVar(&flagAlseStdout, "alsostdout", false, "log to standard error as well as files")
	flag.Var(&flagLoglevel, "loglevel", "log level[debug,info,warn,error]")
	flag.StringVar(&flagLogname, "logname", "", "log name")
	flag.StringVar(&flagLogpath, "logpath", "", "log path default for filelog(/data/projlogs)")
	log.SetFlags(0)
	log.SetOutput(NewLogWriter(levelDefault))
}

type XLogger struct {
	logType    int
	logPath    string
	logName    string
	logLevel   Level
	alsoStdout bool
	writer     Writer
	mu         sync.Mutex

	concurrency      int    // 并发度
	logNumPerRequest int    // 单请求日志量
	pipePath         string //管道路径
}

func (this *XLogger) level() Level {
	if flagLoglevel > levelDefault {
		return flagLoglevel
	}
	return this.logLevel
}

func (this *XLogger) initWriter() {
	this.mu.Lock()
	defer this.mu.Unlock()
	if this.writer != nil {
		return
	}
	if flagLogname != "" {
		this.logName = flagLogname
	}
	if flagAlseStdout {
		this.alsoStdout = true
	}
	switch this.logType {
	case logTypeSyslog:
		this.writer = newSyslogWriter(this.logName)
	case logTypeFilelog:
		logpath := this.logPath
		if flagLogpath != "" {
			logpath = flagLogpath
		}
		this.writer = newFileLog(this.logName, logpath)
	case logTypePipelog:
		logpath := this.logPath
		if flagLogpath != "" {
			logpath = flagLogpath
		}
		this.writer = newPipelogWriter(this.logName, logpath)
	case logTypeAsynPipe:
		this.writer = newAsynPipeWriter(this.pipePath, this.logName, this.concurrency, this.logNumPerRequest)
	}
}

func (this *XLogger) Init(logName string, logLevel Level) {
	Init(logName, logLevel)
}

func Init(logName string, logLevel Level) {
	if inited {
		fmt.Println("logkit has be inited")
	}
	SetName(logName)
	SetLevel(logLevel)
	inited = true
}

func SetLogType(logType int) {
	Logger.logType = logType
}

func SetConcurrency(concurrency int) {
	Logger.concurrency = concurrency
}

func SetLogNumPerRequest(logNumPerRequest int) {
	Logger.logNumPerRequest = logNumPerRequest
}

func InitAsynPipeLog(pipePath string, logName string, logLevel Level, concurrency int, logNumPerRequest int) {
	SetName(logName)
	SetConcurrency(concurrency)
	SetLogNumPerRequest(logNumPerRequest)
	SetLogType(logTypeAsynPipe)
	Logger.pipePath = pipePath
	SetLevel(logLevel)
	inited = true
}

func InitFileLog(logName string, logLevel Level, logPath string) {
	if inited {
		fmt.Println("logkit has be inited")
	}
	SetName(logName)
	SetLevel(logLevel)
	Logger.logType = logTypeFilelog
	if logPath == "" {
		logPath = "/data/projlogs/"
	}
	SetLogPath(logPath)
	inited = true
}

func InitPipeLog(logName string, logLevel Level, pipeFile string) {
	if inited {
		fmt.Println("logkit has be inited")
	}
	SetName(logName)
	SetLevel(logLevel)
	Logger.logType = logTypePipelog
	if pipeFile == "" {
		pipeFile = "/data/projlogs/logkit_pipe"
	}
	SetLogPath(pipeFile)
	inited = true
}

func SetLevel(logLevel Level) {
	if logLevel < LevelDebug || logLevel > LevelError {
		panic("invalid log level")
	}
	Logger.logLevel = logLevel
}

func SetName(logName string) {
	if logName == "" {
		panic("invalid log name")
	}
	Logger.logName = logName
}

func AlsoStdout(alsoStdout bool) {
	Logger.alsoStdout = alsoStdout
}

func SetLogPath(logpath string) {
	if logpath != "" {
		Logger.logPath = logpath
	}
}

func write(level Level, msg string) {
	if !inited {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05") + " [" + levelName[level] + "] " + msg)
		return
	}
	if Logger.writer == nil {
		Logger.initWriter()
	}
	Logger.writer.write(level, msg)
	if Logger.alsoStdout {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05") + " [" + levelName[level] + "] " + msg)
	}
}

func Exit() {
	if Logger.writer != nil {
		Logger.writer.exit()
	}
}

func Wait(fn func()) {
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGTERM)
	s := <-sigCh
	if fn != nil {
		fn()
	}
	fmt.Printf("signal [%v]\n", s)
	Exit()
}

func IsDebug() bool {
	return Logger.level() == LevelDebug
}

func (this *XLogger) Debug(str string) {
	if this.level() <= LevelDebug {
		write(LevelDebug, str)
	}
}

func Debug(str string) {
	if Logger.level() <= LevelDebug {
		write(LevelDebug, str)
	}
}

func (this *XLogger) Debugs(args ...interface{}) {
	if this.level() <= LevelDebug {
		write(LevelDebug, fmt.Sprintln(args))
	}
}

func Debugs(args ...interface{}) {
	if Logger.level() <= LevelDebug {
		write(LevelDebug, fmt.Sprintln(args))
	}
}

func (this *XLogger) Debugf(format string, args ...interface{}) {
	if this.level() <= LevelDebug {
		write(LevelDebug, fmt.Sprintf(format, args...))
	}
}

func Debugf(format string, args ...interface{}) {
	if Logger.level() <= LevelDebug {
		write(LevelDebug, fmt.Sprintf(format, args...))
	}
}

func (this *XLogger) Info(str string) {
	if this.level() <= LevelInfo {
		write(LevelInfo, str)
	}
}

func Info(str string) {
	if Logger.level() <= LevelInfo {
		write(LevelInfo, str)
	}
}

func (this *XLogger) Infos(args ...interface{}) {
	if this.level() <= LevelInfo {
		write(LevelInfo, fmt.Sprintln(args))
	}
}

func Infos(args ...interface{}) {
	if Logger.level() <= LevelInfo {
		write(LevelInfo, fmt.Sprintln(args))
	}
}

func (this *XLogger) Infof(format string, args ...interface{}) {
	if this.level() <= LevelInfo {
		write(LevelInfo, fmt.Sprintf(format, args...))
	}
}

func Infof(format string, args ...interface{}) {
	if Logger.level() <= LevelInfo {
		write(LevelInfo, fmt.Sprintf(format, args...))
	}
}

func (this *XLogger) Warn(str string) {
	if this.level() <= LevelWarn {
		write(LevelWarn, str)
	}
}

func Warn(str string) {
	if Logger.level() <= LevelWarn {
		write(LevelWarn, str)
	}
}

func (this *XLogger) Warns(args ...interface{}) {
	if this.level() <= LevelWarn {
		write(LevelWarn, fmt.Sprintln(args))
	}
}

func Warns(args ...interface{}) {
	if Logger.level() <= LevelWarn {
		write(LevelWarn, fmt.Sprintln(args))
	}
}

func (this *XLogger) Warnf(format string, args ...interface{}) {
	if this.level() <= LevelWarn {
		write(LevelWarn, fmt.Sprintf(format, args...))
	}
}

func Warnf(format string, args ...interface{}) {
	if Logger.level() <= LevelWarn {
		write(LevelWarn, fmt.Sprintf(format, args...))
	}
}

func (this *XLogger) Error(str string) {
	if this.level() <= LevelError {
		write(LevelError, str)
	}
}

func Error(str string) {
	if Logger.level() <= LevelError {
		write(LevelError, str)
	}
}

func (this *XLogger) Errors(args ...interface{}) {
	if this.level() <= LevelError {
		write(LevelError, fmt.Sprintln(args))
	}
}

func Errors(args ...interface{}) {
	if Logger.level() <= LevelError {
		write(LevelError, fmt.Sprintln(args))
	}
}

func (this *XLogger) Errorf(format string, args ...interface{}) {
	if this.level() <= LevelError {
		write(LevelError, fmt.Sprintf(format, args...))
	}
}

func Errorf(format string, args ...interface{}) {
	if Logger.level() <= LevelError {
		write(LevelError, fmt.Sprintf(format, args...))
	}
}

func Action(action string, fields Fields) {
	if action == "" {
		return
	}

	action = "{\"action\":\"" + action + "\""
	if fields != nil {
		data, err := json.Marshal(fields)
		if err == nil {
			action += ",\"fields\":" + string(data)
		}
	}
	action += "}"

	write(LevelAction, action)
}

func Actions(action string, fields string) {
	if action == "" {
		return
	}

	action = "{\"action\":\"" + action + "\""
	if fields != "" {
		if fields[0] == '{' || fields[0] == '[' {
			action += ",\"fields\":" + fields
		} else {
			action += ",\"fields\":\"" + fields + "\""
		}
	}
	action += "}"

	write(LevelAction, action)
}

type Fields map[string]interface{}

func NewStdLog(level Level, prefix string) *log.Logger {
	return log.New(NewLogWriter(level), prefix, 0)
}

func NewLogWriter(level Level) io.Writer {
	return &LogWriter{level}
}

type LogWriter struct {
	level Level
}

func (this *LogWriter) Write(data []byte) (int, error) {
	level := this.level
	if this.level == levelDefault && len(data) > 3 && data[0] == '[' { // from built-in log
		switch string(data[:3]) {
		case "[D]":
			level = LevelDebug
			data = data[3:]
		case "[I]":
			level = LevelInfo
			data = data[3:]
		case "[W]":
			level = LevelWarn
			data = data[3:]
		case "[E]":
			level = LevelError
			data = data[3:]
		}
	}

	if level == levelDefault || Logger.level() <= level {
		write(level, string(data))
	}
	return len(data), nil
}

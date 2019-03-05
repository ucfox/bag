### Usage

#### Init

- syslog模式

```go
logkit.Init("logname", logkit.LevelDebug)
```
- filelog 模式

```go
// 默认 path 为/data/projlogs
logkit.InitFileLog("logname",logkit.LevelDebug,"path") // path can be "",default /data/projlogs
```

- pipelog 模式

```go
logkit.InitPipeLog(logName, logLevel, pipeFile)
//参数一 表示日志名
//参数二 表示日志级别
//参数三 管道的路径 统一配置的（问庆斌）
```

#### golang日志量大影响qps问题分析
见[https://wiki.pandatv.com/pages/viewpage.action?pageId=4498057](https://wiki.pandatv.com/pages/viewpage.action?pageId=4498057)
#### 设置同时输出日志到 stdout

```go
logkit.AlsoStdout(true)
// or
xxx -alsostdout
```

#### 支持Flag
支持通过Flag覆盖原有设置：
- `alsostdout` 同时写到stdout
- `loglevel`  debug|info|warn|error
- `logname`  log name
- `logpath` 只针对 filelog 有效，设置 log 路径

#### 优雅退出
文件Log下，由于log缓存的问题，需要在退出时保证Log完全刷盘到硬盘，logkit 提供了两种优雅退出方式：

1. 在退出时手动调用`Exit` 函数

```go
logkit.Exit()
```

2. 使用 `Wait` 函数
```go
logkit.Wait(fn)
```

   对`Wait` 的调用将产生阻塞，并监听系统退出，在退出时首先调用传入的fn函数，然后退出 logkit 返回

####  Is Debug

判断是否Debug Level:
```go
logkit.IsDebug()
```
#### Action 日志

将以 json格式记录日志，用于特殊统计需求

```go
logkit.Action("action", fields)
```

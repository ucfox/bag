package logkit

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

//检查/data/projlogs/logkit_test/下是否有相应的日志
func Test_Pipelog(t *testing.T) {
	wait := &sync.WaitGroup{}
	InitPipeLog("logkit_syslog", LevelDebug, "./pipe")
	wait.Add(2000000)
	for i := 0; i < 2000000; i++ {
		go _write(wait, i)
		time.Sleep(time.Microsecond)
	}
	wait.Wait()
	fmt.Println("done")
	Exit()
}

func _write(wait *sync.WaitGroup, i int) {
	Infof("test info xxxxx测试xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx-%d", i)
	wait.Done()
}

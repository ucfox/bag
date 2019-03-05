package mysql

import (
	"sync/atomic"
)

type countInt32 int32

func (this *countInt32) Incr() int {
	val := atomic.AddInt32((*int32)(this), 1)
	if val >= 0 && val < 2000000000 {
		return int(val)
	} else {
		atomic.StoreInt32((*int32)(this), 0)
		return 0
	}
}

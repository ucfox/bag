package redis

import (
	"sync/atomic"
)

type indexInt32 int32

func (this *indexInt32) Incr() int {
	val := atomic.AddInt32((*int32)(this), 1)
	if val < 0 {
		if atomic.CompareAndSwapInt32((*int32)(this), val, 0) {
			return 0
		} else {
			return this.Incr()
		}
	}
	return int(val)
}

func (this *indexInt32) IncrAndMod(n int) int {
	return this.Incr() % n
}

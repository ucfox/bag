package concurrent

import (
	"fmt"
	//"reflect"
	"runtime"
	"strconv"
	"sync"
	"testing"
)

var (
	listN     int
	number    int
	list      [][]interface{}
	readCM_ht *ConcurrentMap
	readCM_rw *ConcurrentMap
	readLM    *lockMap
	readM     map[interface{}]interface{}
)

func init() {
	MAXPROCS := runtime.NumCPU()
	runtime.GOMAXPROCS(MAXPROCS)
	listN = MAXPROCS + 1
	number = 1000000
	fmt.Println("MAXPROCS is ", MAXPROCS, ", listN is", listN, ", n is ", number, "\n")

	list = make([][]interface{}, listN, listN)
	for i := 0; i < listN; i++ {
		list1 := make([]interface{}, 0, number)
		for j := 0; j < number; j++ {
			list1 = append(list1, j+(i)*number/10)
		}
		list[i] = list1
	}

	readCM_ht = NewConcurrentMap4(16, 0.75, 16, 0)
	readCM_rw = NewConcurrentMap4(16, 0.75, 16, 1)
	readM = make(map[interface{}]interface{})
	readLM = newLockMap()
	for i := range list[0] {
		readCM_ht.Put(i, i)
		readCM_rw.Put(i, i)
		readLM.put(i, i)
		readM[i] = i
	}
}

type lockMap struct {
	m  map[interface{}]interface{}
	rw *sync.RWMutex
}

func (t *lockMap) iterator(fn func(k, v interface{})) {
	t.rw.RLock()
	defer t.rw.RUnlock()
	for k, v := range t.m {
		fn(k, v)
	}
}

func (t *lockMap) put(k interface{}, v interface{}) {
	t.rw.Lock()
	defer t.rw.Unlock()
	t.m[k] = v
}

func (t *lockMap) putIfNotExist(k interface{}, v interface{}) (ok bool) {
	t.rw.Lock()
	defer t.rw.Unlock()
	if _, ok = t.m[k]; !ok {
		t.m[k] = v
	}
	return
}

func (t *lockMap) get(k interface{}) (v interface{}, ok bool) {
	t.rw.RLock()
	defer t.rw.RUnlock()
	v, ok = t.m[k]
	return
}

func (t *lockMap) len() int {
	t.rw.RLock()
	defer t.rw.RUnlock()
	return len(t.m)

}

func newLockMap() *lockMap {
	return &lockMap{make(map[interface{}]interface{}), new(sync.RWMutex)}
}

func newLockMap1(initCap int) *lockMap {
	return &lockMap{make(map[interface{}]interface{}, initCap), new(sync.RWMutex)}
}

func Benchmark_LockMap_Put(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := newLockMap()

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.put(j, j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func Benchmark_Map_Put(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := make(map[interface{}]interface{})

		//wg := new(sync.WaitGroup)
		//wg.Add(listN)
		for i := 0; i < listN; i++ {
			for _, j := range list[i] {
				cm[j] = j
			}
			//wg.Done()
		}
	}
}

func Benchmark_HT_ConcurrentMap_Put(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := NewConcurrentMap()

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.Put(j, j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func Benchmark_RW_ConcurrentMap_Put(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := NewConcurrentMap4(16, 0.75, 16, 0)

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.Put(j, j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func Benchmark_LockMap_Put_NoGrow(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := newLockMap1(listN * number)

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.put(j, j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func Benchmark_Map_Put_NoGrow(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := make(map[interface{}]interface{}, listN*number)

		//wg := new(sync.WaitGroup)
		//wg.Add(listN)
		for i := 0; i < listN; i++ {
			for _, j := range list[i] {
				cm[j] = j
			}
			//wg.Done()
		}
	}
}

func Benchmark_HT_ConcurrentMap_PutNoGrow(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := NewConcurrentMap4(listN*number, 0.75, 16, 0)

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.Put(j, j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func Benchmark_RW_ConcurrentMap_Put_NoGrow(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := NewConcurrentMap4(listN*number, 0.75, 16, 0)

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.Put(j, j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func Benchmark_LockMap_Put2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := newLockMap()

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.put(strconv.Itoa(j.(int)), j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func Benchmark_Map_Put2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := make(map[interface{}]interface{})

		//wg := new(sync.WaitGroup)
		//wg.Add(listN)
		for i := 0; i < listN; i++ {
			for _, j := range list[i] {
				cm[strconv.Itoa(j.(int))] = j
			}
			//wg.Done()
		}
	}
}

func Benchmark_HT_ConcurrentMap_Put2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := NewConcurrentMap()

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.Put(strconv.Itoa(j.(int)), j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func Benchmark_RW_ConcurrentMapPut2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := NewConcurrentMap4(16, 0.75, 16, 1)

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.Put(strconv.Itoa(j.(int)), j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func Benchmark_LockMap_Get(b *testing.B) {
	for n := 0; n < b.N; n++ {
		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			go func() {
				//itr := NewMapIterator(cm)
				//for itr.HasNext() {
				//	entry := itr.NextEntry()
				//	k := entry.key.(string)
				//	v := entry.value.(int)
				for k := range list[0] {
					_, _ = readLM.get(k)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func Benchmark_Map_Get(b *testing.B) {
	for n := 0; n < b.N; n++ {
		//wg := new(sync.WaitGroup)
		//wg.Add(listN)
		for i := 0; i < listN; i++ {
			for k := range list[0] {
				_, _ = readM[k]
			}
			//wg.Done()
		}
	}
}

func Benchmark_HT_ConcurrentMap_Get(b *testing.B) {
	for n := 0; n < b.N; n++ {
		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			go func() {
				for k := range list[0] {
					_, _ = readCM_ht.Get(k)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func Benchmark_RW_ConcurrentMap_Get(b *testing.B) {
	for n := 0; n < b.N; n++ {
		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			go func() {
				for k := range list[0] {
					_, _ = readCM_rw.Get(k)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func Benchmark_LockMap_PutAndGet(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := newLockMap()

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.put(j, j)
					_, _ = cm.get(j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func Benchmark_Map_PutAndGet(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := make(map[interface{}]interface{})

		//wg := new(sync.WaitGroup)
		//wg.Add(listN)
		for i := 0; i < listN; i++ {
			for _, j := range list[i] {
				cm[j] = j
				_ = cm[j]
			}
			//wg.Done()
		}
	}
}

func Benchmark_HT_ConcurrentMap_PutAndGet(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := NewConcurrentMap()

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.Put(j, j)
					_, _ = cm.Get(j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func Benchmark_RW_ConcurrentMapPutAndGet(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := NewConcurrentMap4(16, 0.75, 16, 1)

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.Put(j, j)
					_, _ = cm.Get(j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func Benchmark_LockMap_GetAll(b *testing.B) {
	for n := 0; n < b.N; n++ {
		readLM.iterator(func(k, v interface{}) {})
	}
}

func Benchmark_Map_GetAll(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for _, _ = range readM {
		}
	}
}

func Benchmark_ht_cm_getall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		wait := &sync.WaitGroup{}
		wait.Add(number)
		readCM_ht.Range(func(k, v interface{}) {
			wait.Done()
		})
		wait.Wait()
	}
}

func Benchmark_rw_cm_getall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		wait := &sync.WaitGroup{}
		wait.Add(number)
		readCM_rw.Range(func(k, v interface{}) {
			wait.Done()
		})
		wait.Wait()
	}
}
func Benchmark_ht_cm_getall_async(b *testing.B) {
	for n := 0; n < b.N; n++ {
		wait := &sync.WaitGroup{}
		wait.Add(number)
		readCM_ht.RangeSync(func(k, v interface{}) {
			wait.Done()
		})
		wait.Wait()
	}
}

func Benchmark_rw_cm_getall_async(b *testing.B) {
	for n := 0; n < b.N; n++ {
		wait := &sync.WaitGroup{}
		wait.Add(number)
		readCM_rw.RangeSync(func(k, v interface{}) {
			wait.Done()
		})
		wait.Wait()
	}
}

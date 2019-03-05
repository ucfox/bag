package concurrent

import (
	"sync"
)

type RwlockSegment struct {
	lock *sync.RWMutex
	bm   map[interface{}]interface{}
}

func NewRwlockSegment(cm *ConcurrentMap, initialCapacity int, lf float32) *RwlockSegment {
	return &RwlockSegment{
		lock: new(sync.RWMutex),
		bm:   make(map[interface{}]interface{}, initialCapacity),
	}
}

func (m *RwlockSegment) Get(k interface{}, hash uint32) interface{} {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if val, ok := m.bm[k]; ok {
		return val
	}
	return nil
}

func (m *RwlockSegment) Put(k interface{}, hash uint32, v interface{}, onlyIfAbsent bool) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	if val, ok := m.bm[k]; !ok {
		m.bm[k] = v
	} else if val != v {
		m.bm[k] = v
	} else {
		return false
	}
	return true
}

func (m *RwlockSegment) ContainsKey(k interface{}, hash uint32) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if _, ok := m.bm[k]; !ok {
		return false
	}
	return true
}

func (m *RwlockSegment) Remove(k interface{}, hash uint32, value interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.bm, k)
}

func (m *RwlockSegment) Size() uint32 {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return uint32(len(m.bm))
}

func (m *RwlockSegment) Clear() {
	m.lock.Lock()
	defer m.lock.Unlock()
	for k, _ := range m.bm {
		delete(m.bm, k)
	}
}

func (this *RwlockSegment) Range(fn func(k, v interface{})) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	for k, v := range this.bm {
		fn(k, v)
	}
}

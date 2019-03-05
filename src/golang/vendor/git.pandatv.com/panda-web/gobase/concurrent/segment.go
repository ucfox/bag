package concurrent

type Segment interface {
	Size() uint32
	Get(key interface{}, hash uint32) interface{}
	ContainsKey(key interface{}, hash uint32) bool
	Put(key interface{}, hash uint32, value interface{}, onlyIfAbsent bool) bool
	Remove(key interface{}, hash uint32, value interface{})
	Clear()
	Range(fn func(k, v interface{}))
}

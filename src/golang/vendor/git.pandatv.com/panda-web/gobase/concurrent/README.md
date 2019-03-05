### ConcurrentMap

#### New::1

```go
func NewConcurrentMap4(initialCapacity int, loadFactor float32, concurrencyLevel int, segmentImpl int) *ConcurrentMap
```

- `initialCapacity` 初始容量，默认16
- `loadFactor` 扩容阈值，默认0.75
- `concurrencyLevel` segment数，默认16
- `segmentImpl` Segment实现，默认 hashtable(0) ，可选 rwlock_map(1)

#### New::2

```go
func NewConcurrentMap2(initialCapacity int, concurrencyLevel int) *ConcurrentMap
```

- 参数同上

#### New::3

```go
func NewConcurrentMap() *ConcurrentMap
```

#### New::4

```go
func NewConcurrentMapFromMap(m map[interface{}]interface{}) *ConcurrentMap
```
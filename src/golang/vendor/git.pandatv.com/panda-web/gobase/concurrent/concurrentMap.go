package concurrent

import (
	"errors"
	//"fmt"
	"io"
	"math"
	"reflect"
	"sync/atomic"
	"unsafe"
)

const (
	/**
	 * The default initial capacity for this table,
	 * used when not otherwise specified in a constructor.
	 */
	DEFAULT_INITIAL_CAPACITY int = 16

	/**
	 * The default load factor for this table, used when not
	 * otherwise specified in a constructor.
	 */
	DEFAULT_LOAD_FACTOR float32 = 0.75

	/**
	 * The default concurrency level for this table, used when not
	 * otherwise specified in a constructor.
	 */
	DEFAULT_CONCURRENCY_LEVEL int = 16

	/**
	 * The maximum capacity, used if a higher value is implicitly
	 * specified by either of the constructors with arguments.  MUST
	 * be a power of two <= 1<<30 to ensure that entries are indexable
	 * using ints.
	 */
	MAXIMUM_CAPACITY int = 1 << 30

	/**
	 * The maximum number of segments to allow; used to bound
	 * constructor arguments.
	 */
	MAX_SEGMENTS int = 1 << 16 // slightly conservative

	/**
	 * Number of unsynchronized retries in size and containsValue
	 * methods before resorting to locking. This is used to avoid
	 * unbounded retries if tables undergo continuous modification
	 * which would make it impossible to obtain an accurate result.
	 */
	RETRIES_BEFORE_LOCK int = 2

	DEFAULT_SEGMENT_IMPL int = 0
)

var (
	Debug           = false
	NilKeyError     = errors.New("Do not support nil as key")
	NilValueError   = errors.New("Do not support nil as value")
	NilActionError  = errors.New("Do not support nil as action")
	NonSupportKey   = errors.New("Non support for pointer, interface, channel, slice, map and function ")
	IllegalArgError = errors.New("IllegalArgumentException")
)

type Hashable interface {
	HashBytes() []byte
	Equals(v2 interface{}) bool
}

type hashEnginer struct {
	putFunc func(w io.Writer, v interface{})
}

//segments is read-only, don't need synchronized
type ConcurrentMap struct {
	engChecker *Once
	eng        unsafe.Pointer

	/**
	 * Mask value for indexing into segments. The upper bits of a
	 * key's hash code are used to choose the segment.
	 */
	segmentMask int

	/**
	 * Shift value for indexing within segments.
	 */
	segmentShift uint

	/**
	 * The segments, each of which is a specialized hash table
	 */
	segments []Segment
}

/**
 * Returns the segment that should be used for key with given hash
 * @param hash the hash code for the key
 * @return the segment
 */
func (this *ConcurrentMap) segmentFor(hash uint32) Segment {
	//默认segmentShift是28，segmentMask是（0xFFFFFFF）,hash>>this.segmentShift就是取前面4位
	//&segmentMask似乎没有必要
	//get first four bytes
	return this.segments[(hash>>this.segmentShift)&uint32(this.segmentMask)]
}

/**
 * Returns true if this map contains no key-value mappings.
 */
func (this *ConcurrentMap) IsEmpty() bool {
	segments := this.segments
	for i := 0; i < len(segments); i++ {
		if segments[i].Size() != 0 {
			return false
		}
	}

	return true
}

/**
 * Returns the number of key-value mappings in this map.
 */
func (this *ConcurrentMap) Size() uint32 {
	var sum uint32 = 0
	for i := 0; i < len(this.segments); i++ {
		sum += this.segments[i].Size()
	}

	return sum
}

/**
 * Returns the value to which the specified key is mapped,
 * or nil if this map contains no mapping for the key.
 */
func (this *ConcurrentMap) Get(key interface{}) (value interface{}, err error) {
	if isNil(key) {
		return nil, NilKeyError
	}
	if hash, e := hashKey(key, this, false); e != nil {
		err = e
	} else {
		Printf("Get, %v, %v\n", key, hash)
		value = this.segmentFor(hash).Get(key, hash)
	}
	return
}

/**
 * Tests if the specified object is a key in this table.
 *
 * @param  key   possible key
 * @return true if and only if the specified object is a key in this table,
 * as determined by the == method; false otherwise.
 */
func (this *ConcurrentMap) ContainsKey(key interface{}) (found bool, err error) {
	if isNil(key) {
		return false, NilKeyError
	}

	if hash, e := hashKey(key, this, false); e != nil {
		err = e
	} else {
		Printf("ContainsKey, %v, %v\n", key, hash)
		found = this.segmentFor(hash).ContainsKey(key, hash)
	}
	return
}

/**
 * Maps the specified key to the specified value in this table.
 * Neither the key nor the value can be nil.
 *
 * The value can be retrieved by calling the get method
 * with a key that is equal to the original key.
 *
 * @param key with which the specified value is to be associated
 * @param value to be associated with the specified key
 *
 * @return the previous value associated with key, or
 *         nil if there was no mapping for key
 */
func (this *ConcurrentMap) Put(key interface{}, value interface{}) error {
	if isNil(key) {
		return NilKeyError
	}
	if isNil(value) {
		return NilValueError
	}

	if hash, e := hashKey(key, this, false); e != nil {
		return e
	} else {
		Printf("Put, %v, %v\n", key, hash)
		this.segmentFor(hash).Put(key, hash, value, false)
	}
	return nil
}

/**
 * If mapping exists for the key, then maps the specified key to the specified value in this table.
 * else will ignore.
 * Neither the key nor the value can be nil.
 *
 * The value can be retrieved by calling the get method
 * with a key that is equal to the original key.
 *
 * @return the previous value associated with the specified key,
 *         or nil if there was no mapping for the key
 */
func (this *ConcurrentMap) PutIfAbsent(key interface{}, value interface{}) (bool, error) {
	if isNil(key) {
		return false, NilKeyError
	}
	if isNil(value) {
		return false, NilValueError
	}

	if hash, e := hashKey(key, this, false); e != nil {
		return false, e
	} else {
		Printf("PutIfAbsent, %v, %v\n", key, hash)
		return this.segmentFor(hash).Put(key, hash, value, true), nil
	}
}

/**
 * Copies all of the mappings from the specified map to this one.
 * These mappings replace any mappings that this map had for any of the
 * keys currently in the specified map.
 *
 * @param m mappings to be stored in this map
 */
func (this *ConcurrentMap) PutAll(m map[interface{}]interface{}) (err error) {
	if isNil(m) {
		err = errors.New("Cannot copy nil map")
	}
	for k, v := range m {
		this.Put(k, v)
	}
	return
}

/**
 * Removes the key (and its corresponding value) from this map.
 * This method does nothing if the key is not in the map.
 *
 * @param  key the key that needs to be removed
 * @return the previous value associated with key, or nil if there was no mapping for key
 */
func (this *ConcurrentMap) Remove(key interface{}) error {
	if isNil(key) {
		return NilKeyError
	}

	if hash, e := hashKey(key, this, false); e != nil {
		return e
	} else {
		Printf("Remove, %v, %v\n", key, hash)
		this.segmentFor(hash).Remove(key, hash, nil)
		return nil
	}
}

func (this *ConcurrentMap) Range(fn func(k, v interface{})) {
	for _, segment := range this.segments {
		go segment.Range(fn)
	}
}

func (this *ConcurrentMap) RangeSync(fn func(k, v interface{})) {
	for _, segment := range this.segments {
		segment.Range(fn)
	}
}

/**
 * Removes all of the mappings from this map.
 */
func (this *ConcurrentMap) Clear() {
	for i := 0; i < len(this.segments); i++ {
		this.segments[i].Clear()
	}
}

func (this *ConcurrentMap) parseKey(key interface{}) (err error) {
	this.engChecker.Do(func() {
		var eng *hashEnginer

		val := key

		if _, ok := val.(Hashable); ok {
			eng = hasherEng
		} else {
			switch v := val.(type) {
			case bool:
				_ = v
				eng = boolEng
			case int:
				eng = intEng
			case int8:
				eng = int8Eng
			case int16:
				eng = int16Eng
			case int32:
				eng = int32Eng
			case int64:
				eng = int64Eng
			case uint:
				eng = uintEng
			case uint8:
				eng = uint8Eng
			case uint16:
				eng = uint16Eng
			case uint32:
				eng = uint32Eng
			case uint64:
				eng = uint64Eng
			case uintptr:
				eng = uintptrEng
			case float32:
				eng = float32Eng
			case float64:
				eng = float64Eng
			case complex64:
				eng = complex64Eng
			case complex128:
				eng = complex128Eng
			case string:
				eng = stringEng
			default:
				Printf("key = %v, other case\n", key)
				//some types can be used as key, we can use equals to test
				//_ = val == val

				rv := reflect.ValueOf(val)
				if ki, e := getKeyInfo(rv.Type()); e != nil {
					err = e
					return
				} else {
					putF := getPutFunc(ki)
					eng = &hashEnginer{}
					eng.putFunc = putF
				}
			}
		}

		this.eng = unsafe.Pointer(eng)

		Printf("key = %v, eng=%v, %v\n", key, this.eng, eng)
	})
	return
}

func NewConcurrentMap4(initialCapacity int, loadFactor float32, concurrencyLevel int, segmentImpl int) *ConcurrentMap {
	if initialCapacity < 0 || concurrencyLevel <= 0 || loadFactor <= 0 || loadFactor >= 1 || (segmentImpl != 0 && segmentImpl != 1) {
		panic(IllegalArgError)
	}

	m := &ConcurrentMap{}
	if concurrencyLevel > MAX_SEGMENTS {
		concurrencyLevel = MAX_SEGMENTS
	}

	// Find power-of-two sizes best matching arguments
	sshift := 0
	ssize := 1
	for ssize < concurrencyLevel {
		sshift++
		ssize = ssize << 1
	}

	m.segmentShift = uint(32) - uint(sshift)
	m.segmentMask = ssize - 1

	m.segments = make([]Segment, ssize)

	if initialCapacity > MAXIMUM_CAPACITY {
		initialCapacity = MAXIMUM_CAPACITY
	}

	c := initialCapacity / ssize
	if c*ssize < initialCapacity {
		c++
	}
	cap := 1
	for cap < c {
		cap <<= 1
	}

	for i := 0; i < len(m.segments); i++ {
		if segmentImpl == 1 {
			m.segments[i] = NewRwlockSegment(m, cap, loadFactor)
		} else {
			m.segments[i] = NewHashtableSegment(m, cap, loadFactor)
		}
	}
	m.engChecker = new(Once)
	return m
}

func NewConcurrentMap2(initialCapacity int, concurrencyLevel int) *ConcurrentMap {
	return NewConcurrentMap4(initialCapacity, DEFAULT_LOAD_FACTOR, concurrencyLevel, DEFAULT_SEGMENT_IMPL)
}

func NewConcurrentMap() *ConcurrentMap {
	return NewConcurrentMap2(DEFAULT_INITIAL_CAPACITY, DEFAULT_CONCURRENCY_LEVEL)
}

/**
 * Creates a new map with the same mappings as the given map.
 * The map is created with a capacity of 1.5 times the number
 * of mappings in the given map or 16 (whichever is greater),
 * and a default load factor (0.75) and concurrencyLevel (16).
 *
 * @param m the map
 */
func NewConcurrentMapFromMap(m map[interface{}]interface{}) *ConcurrentMap {
	cm := NewConcurrentMap4(int(math.Max(float64(float32(len(m))/DEFAULT_LOAD_FACTOR+1),
		float64(DEFAULT_INITIAL_CAPACITY))),
		DEFAULT_LOAD_FACTOR, DEFAULT_CONCURRENCY_LEVEL, DEFAULT_SEGMENT_IMPL)
	cm.PutAll(m)
	return cm
}

/**
 * ConcurrentHashMap list entry.
 * Note only value field is variable and must use atomic to read/write it, other three fields are read-only after initializing.
 * so can use unsynchronized reader, the Segment.readValueUnderLock method is used as a
 * backup in case a nil (pre-initialized) value is ever seen in
 * an unsynchronized access method.
 */
type Entry struct {
	key   interface{}
	hash  uint32
	value unsafe.Pointer
	next  *Entry
}

func (this *Entry) Key() interface{} {
	return this.key
}

func (this *Entry) Value() interface{} {
	return *((*interface{})(atomic.LoadPointer(&this.value)))
}

func (this *Entry) fastValue() interface{} {
	return *((*interface{})(this.value))
}

func (this *Entry) storeValue(v *interface{}) {
	atomic.StorePointer(&this.value, unsafe.Pointer(v))
}

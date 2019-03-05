package concurrent

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

type HashtableSegment struct {
	m *ConcurrentMap //point to concurrentMap.eng, so it is **hashEnginer
	/**
	 * The number of elements in this segment's region.
	 * Must use atomic package's LoadInt32 and StoreInt32 functions to read/write this field
	 * otherwise read operation may cannot read latest value
	 */
	count uint32

	/**
	 * Number of updates that alter the size of the table. This is
	 * used during bulk-read methods to make sure they see a
	 * consistent snapshot: If modCounts change during a traversal
	 * of segments computing size or checking containsValue, then
	 * we might have an inconsistent view of state so (usually)
	 * must retry.
	 */
	modCount uint32

	/**
	 * The table is rehashed when its size exceeds this threshold.
	 * (The value of this field is always (int)(capacity *
	 * loadFactor).)
	 */
	threshold uint32

	/**
	 * The per-segment table.
	 * Use unsafe.Pointer because must use atomic.LoadPointer function in read operations.
	 */
	pTable unsafe.Pointer //point to []unsafe.Pointer

	/**
	 * The load factor for the hash table. Even though this value
	 * is same for all segments, it is replicated to avoid needing
	 * links to outer object.
	 */
	loadFactor float32

	lock *sync.Mutex
}

func NewHashtableSegment(cm *ConcurrentMap, initialCapacity int, lf float32) (s *HashtableSegment) {
	s = new(HashtableSegment)
	s.loadFactor = lf
	table := make([]unsafe.Pointer, initialCapacity)
	s.setTable(table)
	s.lock = new(sync.Mutex)
	s.m = cm
	return
}

func (this *HashtableSegment) Size() uint32 {
	return atomic.LoadUint32(&this.count)
}

func (this *HashtableSegment) enginer() *hashEnginer {
	return (*hashEnginer)(atomic.LoadPointer(&this.m.eng))
}

func (this *HashtableSegment) rehash() {
	oldTable := this.table() //*(*[]*Entry)(this.table)
	oldCapacity := len(oldTable)
	if oldCapacity >= MAXIMUM_CAPACITY {
		return
	}

	/*
	 * Reclassify nodes in each list to new Map.  Because we are
	 * using power-of-two expansion, the elements from each bin
	 * must either stay at same index, or move with a power of two
	 * offset. We eliminate unnecessary node creation by catching
	 * cases where old nodes can be reused because their next
	 * fields won't change. Statistically, at the default
	 * threshold, only about one-sixth of them need cloning when
	 * a table doubles. The nodes they replace will be garbage
	 * collectable as soon as they are no longer referenced by any
	 * reader thread that may be in the midst of traversing table
	 * right now.
	 */

	newTable := make([]unsafe.Pointer, oldCapacity<<1)
	atomic.StoreUint32(&this.threshold, uint32(float32(len(newTable))*this.loadFactor))
	sizeMask := uint32(len(newTable) - 1)
	for i := 0; i < oldCapacity; i++ {
		// We need to guarantee that any existing reads of old Map can
		//  proceed. So we cannot yet nil out each bin.
		e := (*Entry)(oldTable[i])

		if e != nil {
			next := e.next
			//计算节点扩容后新的数组下标
			idx := e.hash & sizeMask

			//  Single node on list
			//如果没有后续的碰撞节点，直接复制到新数组即可
			if next == nil {
				newTable[idx] = unsafe.Pointer(e)
			} else {
				/* Reuse trailing consecutive sequence at same slot
				 * 数组扩容后原来数组下标相同（碰撞）的节点可能会计算出不同的新下标
				 * 如果把碰撞链表中所有节点的新下标列出，并将相邻的新下标相同的节点视为一段
				 * 那么下面的代码为了提高效率，会循环碰撞链表，找到链表中最后一段首节点（之后所有节点的新下标相同）
				 * 然后将这个首节点复制到新数组，后续节点因为计算出的新下标相同，所以在扩容后的数组中仍然在同一碰撞链表中
				 * 所以新的首节点的碰撞链表是正确的
				 * 新的首节点之外的其他现存碰撞链表上的节点，则重新复制到新节点（这个重要，可以保持旧节点的不变性）后放入新数组
				 * 这个过程的关键在于维持所有旧节点的next属性不会发生变化，这样才能让无锁的读操作保持线程安全
				 */
				lastRun := e
				lastIdx := idx
				for last := next; last != nil; last = last.next {
					k := last.hash & uint32(sizeMask)
					//发现新下标不同的节点就保存到lastIdx和lastRun中
					//所以lastIdx和lastRun总是对应现有碰撞链表中最后一段新下标相同节点的首节点和其对应的新下标
					if k != lastIdx {
						lastIdx = k
						lastRun = last
					}
				}
				newTable[lastIdx] = unsafe.Pointer(lastRun)

				// Clone all remaining nodes
				for p := e; p != lastRun; p = p.next {
					k := p.hash & sizeMask
					n := newTable[k]
					newTable[k] = unsafe.Pointer(&Entry{p.key, p.hash, p.value, (*Entry)(n)})
				}
			}
		}
	}
	atomic.StorePointer(&this.pTable, unsafe.Pointer(&newTable))
}

/**
 * Sets table to new pointer slice that all item points to HashEntry.
 * Call only while holding lock or in constructor.
 */
func (this *HashtableSegment) setTable(newTable []unsafe.Pointer) {
	this.threshold = (uint32)(float32(len(newTable)) * this.loadFactor)
	this.pTable = unsafe.Pointer(&newTable)
}

/**
 * uses atomic to load table and returns.
 * Call while no lock.
 */
func (this *HashtableSegment) loadTable() (table []unsafe.Pointer) {
	return *(*[]unsafe.Pointer)(atomic.LoadPointer(&this.pTable))
}

/**
 * returns pointer slice that all item points to HashEntry.
 * Call only while holding lock or in constructor.
 */
func (this *HashtableSegment) table() []unsafe.Pointer {
	return *(*[]unsafe.Pointer)(this.pTable)
}

/**
 * Returns properly casted first entry of bin for given hash.
 */
func (this *HashtableSegment) getFirst(hash uint32) *Entry {
	tab := this.loadTable()
	return (*Entry)(atomic.LoadPointer(&tab[hash&uint32(len(tab)-1)]))
}

/**
 * Reads value field of an entry under lock. Called if value
 * field ever appears to be nil. see below code:
 * 		tab[index] = unsafe.Pointer(&Entry{key, hash, unsafe.Pointer(&value), first})
 * go memory model don't explain Entry initialization must be executed before
 * table assignment. So value is nil is possible only if a
 * compiler happens to reorder a HashEntry initialization with
 * its table assignment, which is legal under memory model
 * but is not known to ever occur.
 */
func (this *HashtableSegment) readValueUnderLock(e *Entry) interface{} {
	this.lock.Lock()
	defer this.lock.Unlock()
	return e.fastValue()
}

/* Specialized implementations of map methods */

func (this *HashtableSegment) Get(key interface{}, hash uint32) interface{} {
	if atomic.LoadUint32(&this.count) != 0 { // atomic-read
		e := this.getFirst(hash)
		for e != nil {
			if e.hash == hash && equals(e.key, key) {
				v := e.Value()
				if v != nil {
					//return
					return v
				}
				return this.readValueUnderLock(e) // recheck
			}
			e = e.next
		}
	}
	return nil
}

func (this *HashtableSegment) ContainsKey(key interface{}, hash uint32) bool {
	if atomic.LoadUint32(&this.count) != 0 { // read-volatile
		e := this.getFirst(hash)
		for e != nil {
			if e.hash == hash && equals(e.key, key) {
				return true
			}
			e = e.next
		}
	}
	return false
}

func (this *HashtableSegment) CompareAndReplace(key interface{}, hash uint32, oldVal interface{}, newVal interface{}) bool {
	this.lock.Lock()
	defer this.lock.Unlock()

	e := this.getFirst(hash)
	for e != nil && (e.hash != hash || !equals(e.key, key)) {
		e = e.next
	}

	replaced := false
	if e != nil && oldVal == e.fastValue() {
		replaced = true
		e.storeValue(&newVal)
	}
	return replaced
}

func (this *HashtableSegment) Replace(key interface{}, hash uint32, newVal interface{}) (oldVal interface{}) {
	this.lock.Lock()
	defer this.lock.Unlock()
	e := this.getFirst(hash)
	for e != nil && (e.hash != hash || !equals(e.key, key)) {
		e = e.next
	}

	if e != nil {
		oldVal = e.fastValue()
		e.storeValue(&newVal)
	}
	return
}

/**
 * put方法牵涉到count, modCount, pTable三个共享变量的修改
 * 在Java中count和pTable是volatile字段，而modCount不是
 * 由于IsEmpty和Size等操作会读取count, modCount和pTable并且是无锁的，这里有必要对进行并发安全性的分析
 * 在Java中，volatile的读具有Acquire语义，volatile的写具有release语义，而put的最后会写入count，
 * 其他读操作总是会先读取count，由此保证了put中其他的写入操作不会被reorder到写入count之后，而读操作中其他的读取不会被reorder到读count之前
 * 由此保证了多线程情况下读和写线程中看到的操作次序不会发送混乱，
 * 在Golang中，StorePointer内部使用了xchgl指令，具有内存屏障，但是Load操作似乎并未具有明确的acquire语义
 */
func (this *HashtableSegment) Put(key interface{}, hash uint32, value interface{}, onlyIfAbsent bool) bool {
	this.lock.Lock()
	defer this.lock.Unlock()

	c := this.count
	if c > this.threshold { // ensure capacity
		this.rehash()
	}

	tab := this.table()
	index := hash & uint32(len(tab)-1)
	first := (*Entry)(tab[index])
	e := first

	for e != nil && (e.hash != hash || !equals(e.key, key)) {
		e = e.next
	}

	if e != nil {
		if onlyIfAbsent {
			return false
		}
		e.storeValue(&value)
	} else {
		c++
		this.modCount++
		tab[index] = unsafe.Pointer(&Entry{key, hash, unsafe.Pointer(&value), first})
		atomic.StoreUint32(&this.count, c) // atomic write 这里可以保证对modCount和tab的修改不会被reorder到this.count之后
	}
	return true
}

/**
 * Remove; match on key only if value nil, else match both.
 */
func (this *HashtableSegment) Remove(key interface{}, hash uint32, value interface{}) {
	this.lock.Lock()
	defer this.lock.Unlock()

	c := this.count - 1
	tab := this.table()
	index := hash & uint32(len(tab)-1)
	first := (*Entry)(tab[index])
	e := first

	for e != nil && (e.hash != hash || !equals(e.key, key)) {
		e = e.next
	}

	if e != nil {
		v := e.fastValue()
		if value == nil || value == v {
			// All entries following removed node can stay
			// in list, but all preceding ones need to be
			// cloned.
			this.modCount++
			newFirst := e.next
			for p := first; p != e; p = p.next {
				newFirst = &Entry{p.key, p.hash, p.value, newFirst}
			}
			tab[index] = unsafe.Pointer(newFirst)
			atomic.StoreUint32(&this.count, c) //this.count = c
		}
	}
	return
}

func (this *HashtableSegment) Clear() {
	if atomic.LoadUint32(&this.count) != 0 {
		this.lock.Lock()
		defer this.lock.Unlock()

		tab := this.table()
		for i := 0; i < len(tab); i++ {
			tab[i] = nil
		}
		this.modCount++
		atomic.StoreUint32(&this.count, 0) //this.count = 0 // write-volatile
	}
}

func (this *HashtableSegment) Range(fn func(k, v interface{})) {
	for _, t := range this.loadTable() {
		e := (*Entry)(atomic.LoadPointer(&t))
		for e != nil {
			fn(e.Key(), e.Value())
			e = e.next
		}
	}
}

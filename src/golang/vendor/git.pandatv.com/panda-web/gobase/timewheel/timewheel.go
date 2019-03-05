package timewheel

import (
	"container/list"
	"time"
)

type wheel struct {
	rotation int
	duration time.Duration
	callback func()
}

type TimeWheel struct {
	slotNum  int
	interval time.Duration
	pos      int
	slots    []*list.List
	ticker   *time.Ticker
	closeCh  chan struct{}
	addCh    chan *wheel
}

func NewTimeWheel(size int, interval time.Duration) *TimeWheel {
	if size <= 0 || interval <= 0 {
		return nil
	}

	slots := make([]*list.List, size)
	for i, _ := range slots {
		slots[i] = list.New()
	}

	closeCh := make(chan struct{})
	addCh := make(chan *wheel, size)

	return &TimeWheel{size, interval, 0, slots, nil, closeCh, addCh}
}

func (tw *TimeWheel) Start() {
	ticker := time.NewTicker(tw.interval)
	tw.ticker = ticker

	go tw.loop()
}

func (tw *TimeWheel) Add(duration time.Duration, callback func()) {
	tw.addCh <- &wheel{0, duration, callback}
}

func (tw *TimeWheel) add(w *wheel) {
	if w == nil {
		return
	}

	ticks := int((w.duration + tw.interval - 1) / tw.interval)
	rotation := ticks / tw.slotNum
	pos := (ticks + tw.pos) % tw.slotNum

	w.rotation = rotation
	tw.slots[pos].PushBack(w)
}

func (tw *TimeWheel) Stop() {
	close(tw.closeCh)
}

func (tw *TimeWheel) handle() {
	slot := tw.slots[tw.pos]
	element := slot.Front()

	for element != nil {
		w := element.Value.(*wheel)
		if w.rotation == 0 {
			w.callback()
			next := element.Next()
			slot.Remove(element)
			element = next
		} else {
			w.rotation--
			element = element.Next()
		}
	}
	tw.pos++
	tw.pos = tw.pos % tw.slotNum
}

func (tw *TimeWheel) loop() {
OutLoop:
	for {
		select {
		case <-tw.ticker.C:
			tw.handle()
		case w := <-tw.addCh:
			tw.add(w)
		case <-tw.closeCh:
			tw.ticker.Stop()
			close(tw.addCh)
			break OutLoop
		}
	}
}

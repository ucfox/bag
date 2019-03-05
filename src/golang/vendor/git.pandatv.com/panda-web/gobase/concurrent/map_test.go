package concurrent

import "testing"

func TestRange(t *testing.T) {
	m := NewConcurrentMap()
	m.Put("a", "b")
}

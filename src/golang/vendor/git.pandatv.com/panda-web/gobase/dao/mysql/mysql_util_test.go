package mysql

import (
	"fmt"
	"testing"
)

func TestIncr(t *testing.T) {
	countint32 := new(countInt32)
	*countint32 = 1999999998
	fmt.Println(countint32.Incr())
	fmt.Println(countint32.Incr())
	fmt.Println(countint32.Incr())
}

package timewheel

import (
	"fmt"
	"testing"
	"time"
)

func echo(input string) {
	fmt.Println(time.Now().UTC())
	fmt.Println(input)
}

func TestTW(t *testing.T) {
	tw := NewTimeWheel(60, time.Second)
	tw.Start()
	tw.Add(1*time.Second, func() {
		echo("hehe")
	})
	tw.Add(3*time.Second, func() {
		echo("haha")
	})
	tw.Add(6*time.Second, func() {
		echo("xixi")
	})

	time.Sleep(8 * time.Second)
	tw.Add(2*time.Second, func() {
		echo("zizi")
	})
	time.Sleep(time.Second)
	tw.Stop()
	time.Sleep(time.Second)
}

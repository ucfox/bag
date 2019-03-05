package kafka

import (
	"fmt"
	"testing"
)

func TestProducer(t *testing.T) {
	brokers := ""
	// new instance
	producer, err := NewAsyncProducer(brokers)
	if err != nil {
		fmt.Println(err)
		return
	}
	// handler err
	go func() {
		for err := range producer.Errors() {
			fmt.Println(err)
		}
	}()
	// send data
	err = producer.Send("test_topic", "key", []byte("data"))
	if err != nil {
		fmt.Println(err)
	}
}

package kafka

import (
	"fmt"
	"testing"
)

func TestConsumer(t *testing.T) {
	// new consumer
	consumer, err := NewKafkaConsumer(zookeeper, "test-group", []string{"test_topic"})
	if err != nil {
		fmt.Println(err)
		return
	}

	// handler err
	go func() {
		for err := range consumer.Errors() {
			fmt.Println(err)
		}
	}()

	// read data
	for data := range consumer.Datas() {
		fmt.Println(data)
	}
}

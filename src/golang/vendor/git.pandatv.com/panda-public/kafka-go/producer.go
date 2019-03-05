package kafka

import (
	"fmt"
	"strings"
	"time"

	"github.com/Shopify/sarama"
)

type KafkaProducer struct {
	producer sarama.AsyncProducer
}

func NewAsyncProducer(brokers string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
	config.Producer.Compression = sarama.CompressionSnappy   // Compress messages
	config.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms

	brokerList := strings.Split(brokers, ",")
	producer, err := sarama.NewAsyncProducer(brokerList, config)
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{producer}, nil
}

func (this *KafkaProducer) Close() error {
	return this.producer.Close()
}

func (this *KafkaProducer) Errors() <-chan *sarama.ProducerError {
	return this.producer.Errors()
}

func (this *KafkaProducer) Send(topic, key string, data []byte) error {
	if this.producer == nil {
		return fmt.Errorf("producer is nil")
	}
	m := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: Message(data),
	}
	this.producer.Input() <- m
	return nil
}

type Message []byte

func (this Message) Encode() ([]byte, error) {
	return this, nil
}

func (this Message) Length() int {
	return len(this)
}

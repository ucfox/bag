package kafka

import (
	"errors"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kafka/consumergroup"
	"github.com/wvanbergen/kazoo-go"
)

type KafkaConsumer struct {
	consumer *consumergroup.ConsumerGroup
	config   *consumerConfigs
	ch       chan *sarama.ConsumerMessage
}

func NewKafkaConsumer(zookeeper, consumerGroup string, topics []string, configs ...consumerConfig) (*KafkaConsumer, error) {
	cfg := &consumerConfigs{
		Config:     consumergroup.NewConfig(),
		AutoCommit: true,
	}
	for _, c := range configs {
		c(cfg)
	}

	var zookeeperNodes []string
	zookeeperNodes, cfg.Zookeeper.Chroot = kazoo.ParseConnectionString(zookeeper)

	consumer, err := consumergroup.JoinConsumerGroup(consumerGroup, topics, zookeeperNodes, cfg.Config)
	if err != nil {
		return nil, fmt.Errorf("init kafka consumerr error:%s", err)
	}

	kc := &KafkaConsumer{
		consumer: consumer,
		config:   cfg,
	}
	if cfg.AutoCommit {
		kc.ch = make(chan *sarama.ConsumerMessage, cfg.ChannelBufferSize)
		go kc.run()
	}
	return kc, nil
}

func (this *KafkaConsumer) Errors() <-chan error {
	return this.consumer.Errors()
}

func (this *KafkaConsumer) Datas() <-chan *sarama.ConsumerMessage {
	if this.ch != nil {
		return this.ch
	}
	return this.consumer.Messages()
}

func (this *KafkaConsumer) Close() error {
	err := this.consumer.Close()
	if err != nil && err != consumergroup.AlreadyClosing {
		return fmt.Errorf("close kafka client error: %s", err)
	}
	return nil
}

func (this *KafkaConsumer) CommitOffset(message *sarama.ConsumerMessage) error {
	if this.config.AutoCommit {
		return errors.New("auto commit is open, do not repeat commit offset")
	}
	return this.consumer.CommitUpto(message)
}

func (this *KafkaConsumer) run() {
	defer close(this.ch)

	for message := range this.consumer.Messages() {
		this.consumer.CommitUpto(message)
		this.ch <- message
	}
}

type consumerConfig func(*consumerConfigs)

type consumerConfigs struct {
	*consumergroup.Config
	AutoCommit bool
}

func ConfigResetOffset(b bool) consumerConfig {
	return func(config *consumerConfigs) {
		config.Offsets.ResetOffsets = b
	}
}

func ConfigCommitInterval(s time.Duration) consumerConfig {
	return func(config *consumerConfigs) {
		config.Offsets.CommitInterval = s
	}
}

func ConfigInitial(i int64) consumerConfig {
	return func(config *consumerConfigs) {
		config.Offsets.Initial = i
	}
}

func ConfigProcessingTimeout(s time.Duration) consumerConfig {
	return func(config *consumerConfigs) {
		config.Offsets.ProcessingTimeout = s
	}
}

func ConfigAutoCommit(b bool) consumerConfig {
	return func(config *consumerConfigs) {
		config.AutoCommit = b
	}
}

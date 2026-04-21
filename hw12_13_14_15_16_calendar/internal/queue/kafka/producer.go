package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/queue"
)

type Producer struct {
	producer sarama.SyncProducer
}

// NewProducer создает новый Kafka producer с retry механизмом.
func NewProducer(brokers []string) (queue.Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Compression = sarama.CompressionSnappy
	config.Version = sarama.V2_6_0_0

	var producer sarama.SyncProducer
	var err error

	maxRetries := 10
	backoff := time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		producer, err = sarama.NewSyncProducer(brokers, config)
		if err == nil {
			return &Producer{producer: producer}, nil
		}

		if attempt < maxRetries-1 {
			time.Sleep(backoff)
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
		}
	}

	return nil, fmt.Errorf("failed to create producer after %d attempts: %w", maxRetries, err)
}

func (p *Producer) SendMessage(_ context.Context, topic string, key, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	fmt.Printf("Message sent to partition %d at offset %d\n", partition, offset)
	return nil
}

func (p *Producer) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}

package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/queue"
)

type Consumer struct {
	brokers []string
	groupID string
}

// NewConsumer создает новый Kafka consumer с retry механизмом.
func NewConsumer(brokers []string, groupID string) queue.Consumer {
	return &Consumer{
		brokers: brokers,
		groupID: groupID,
	}
}

func (c *Consumer) Consume(ctx context.Context, topics []string, handler queue.MessageHandler) error {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Version = sarama.V2_6_0_0

	var group sarama.ConsumerGroup
	var err error

	maxRetries := 10
	backoff := time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		group, err = sarama.NewConsumerGroup(c.brokers, c.groupID, config)
		if err == nil {
			break
		}

		if attempt < maxRetries-1 {
			log.Printf("Failed to create consumer group (attempt %d/%d): %v", attempt+1, maxRetries, err)
			time.Sleep(backoff)
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
		}
	}

	if err != nil {
		return fmt.Errorf("failed to create consumer group after %d attempts: %w", maxRetries, err)
	}
	defer group.Close()

	consumerHandler := &consumerGroupHandler{handler: handler}

	for {
		if err := group.Consume(ctx, topics, consumerHandler); err != nil {
			return fmt.Errorf("error from consumer: %w", err)
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (c *Consumer) Close() error {
	return nil
}

type consumerGroupHandler struct {
	handler queue.MessageHandler
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	for msg := range claim.Messages() {
		if err := h.handler(session.Context(), msg.Value); err != nil {
			log.Printf("Error processing message: %v", err)
			continue
		}

		session.MarkMessage(msg, "")
	}
	return nil
}

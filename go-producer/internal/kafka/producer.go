package kafka

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) (*Producer, error) {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		MaxAttempts:  3,
	}
	return &Producer{writer: w}, nil
}

func (p *Producer) Publish(ctx context.Context, topic, key string, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
		Time:  time.Now(),
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

// EnsureTopics creates required Kafka topics if they don't exist.
func EnsureTopics(brokers []string, topics []string) error {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("connect to kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("get controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("connect to controller: %w", err)
	}
	defer controllerConn.Close()

	specs := make([]kafka.TopicConfig, 0, len(topics))
	for _, t := range topics {
		specs = append(specs, kafka.TopicConfig{
			Topic:             t,
			NumPartitions:     3,
			ReplicationFactor: 1,
		})
	}

	err = controllerConn.CreateTopics(specs...)
	if err != nil && err.Error() != "Topic already exists." {
		return fmt.Errorf("create topics: %w", err)
	}
	return nil
}

package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer - продюссер кафки
type Producer struct {
	writer *kafka.Writer
}

// New - контсруктор
func New(brokers []string, topic string) *Producer {
	writer := kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{
		writer: &writer,
	}
}

// WriteMessages - записываем сообщение в кафку
func (p *Producer) WriteMessages(ctx context.Context, key string, value []byte) error {
	write := kafka.Message{
		Key:   []byte(key),
		Value: value,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)

	defer cancel()

	err := p.writer.WriteMessages(ctx, write)
	if err != nil {
		return fmt.Errorf("WriteMessages usecase: error writing message: %v", err)
	}

	return nil
}

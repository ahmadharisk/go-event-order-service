package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type Publisher struct{ w *kafka.Writer }

func NewPublisher(brokers []string, topic string) *Publisher {
	return &Publisher{w: &kafka.Writer{Addr: kafka.TCP(brokers...), Topic: topic, Balancer: &kafka.LeastBytes{}}}
}

func (p *Publisher) Publish(ctx context.Context, key string, payload any) error {
	b, _ := json.Marshal(payload)
	return p.w.WriteMessages(ctx, kafka.Message{Key: []byte(key), Value: b})
}

func (p *Publisher) Close() error { return p.w.Close() }

func NewReader(brokers []string, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{Brokers: brokers, Topic: topic, GroupID: groupID})
}

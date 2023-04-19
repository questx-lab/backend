package kafka

import (
	"context"
	"fmt"

	"github.com/questx-lab/backend/pkg/pubsub"

	"github.com/Shopify/sarama"
)

type publisher struct {
	clientID    string
	brokerAddrs []string
	producer    sarama.SyncProducer
}

func NewPublisher(
	clientID string,
	brokerAddrs []string,
) pubsub.Publisher {
	config := sarama.NewConfig()
	config.ClientID = clientID
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokerAddrs, config)
	if err != nil {
		panic(err)
	}
	return &publisher{
		clientID:    clientID,
		brokerAddrs: brokerAddrs,
		producer:    producer,
	}
}
func (p *publisher) Stop(ctx context.Context) error {
	return p.producer.Close()
}

func (p *publisher) Publish(ctx context.Context, topic string, msg *pubsub.Pack) error {
	m := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msg.Msg),
		Key:   sarama.ByteEncoder(msg.Key),
	}
	if _, _, err := p.producer.SendMessage(m); err != nil {
		return fmt.Errorf("p.producer.SendMessage: %w", err)
	}
	return nil
}

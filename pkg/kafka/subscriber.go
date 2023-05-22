package kafka

import (
	"context"
	"log"
	"time"

	"github.com/questx-lab/backend/pkg/pubsub"

	"github.com/Shopify/sarama"
)

type subscriber struct {
	groupID     string
	brokerAddrs []string
	topics      []string
	client      sarama.ConsumerGroup
	handler     func(context.Context, *pubsub.Pack, time.Time)
}

func NewSubscriber(
	groupID string,
	brokerAddrs []string,
	topics []string,
	handler func(context.Context, *pubsub.Pack, time.Time),
) pubsub.Subscriber {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	client, err := sarama.NewConsumerGroup(brokerAddrs, groupID, config)
	if err != nil {
		panic(err)
	}

	return &subscriber{
		groupID:     groupID,
		brokerAddrs: brokerAddrs,
		topics:      topics,
		client:      client,
		handler:     handler,
	}
}

func (g *subscriber) Stop(ctx context.Context) error {
	return g.client.Close()
}

func (g *subscriber) Subscribe(ctx context.Context) {
	consumer := consumerGroupHandler{
		ready: make(chan bool),
		fn:    g.handler,
	}
	go func() {
		for {
			// TODO: `Consume` should be called inside an infinite loop, when a
			// TODO: server-side rebalance happens, the consumer session will need to be
			// TODO: recreated to get the new claims
			if err := g.client.Consume(ctx, g.topics, &consumer); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()
	<-consumer.ready
}

type consumerGroupHandler struct {
	ready chan bool
	fn    func(context.Context, *pubsub.Pack, time.Time)
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// TODO: ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Printf("%+v", *message)
		session.MarkMessage(message, "")
		h.fn(session.Context(), &pubsub.Pack{
			Key: message.Key,
			Msg: message.Value,
		}, message.Timestamp)
	}
	return nil
}

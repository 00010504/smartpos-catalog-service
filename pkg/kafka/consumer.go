package kafka

import (
	"context"
	"errors"

	"github.com/Invan2/invan_catalog_service/pkg/helper"
	"github.com/Shopify/sarama"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type HandlerFunc func(context.Context, cloudevents.Event)

type Consumer struct {
	consumerName string
	topic        string
	handler      HandlerFunc
}

func (kafka *Kafka) AddConsumer(topic string, handler HandlerFunc) {
	if kafka.consumers[topic] != nil {
		panic(errors.New("consumer with the same name already exists: " + topic))
	}

	kafka.consumers[topic] = &Consumer{
		consumerName: topic,
		topic:        topic,
		handler:      handler,
	}
}

func (kafka *Kafka) Setup(_ sarama.ConsumerGroupSession) error {
	close(kafka.ready)
	return nil
}

func (kafka *Kafka) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (kafka *Kafka) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	consumer := kafka.consumers[claim.Topic()]
	for message := range claim.Messages() {

		event := helper.MessageToEvent(message)

		session.MarkMessage(message, "")
		consumer.handler(kafka.ctx, event)

	}
	return nil
}

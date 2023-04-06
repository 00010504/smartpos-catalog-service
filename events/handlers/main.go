package handlers

import (
	"encoding/json"

	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/Invan2/invan_catalog_service/storage"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type EventHandler struct {
	log      logger.Logger
	strgPG   storage.StoragePg
	strgES   storage.StorageES
	producer *kafka.Producer
}

func NewHandler(log logger.Logger, strgPG storage.StoragePg, strgES storage.StorageES, producer *kafka.Producer) *EventHandler {
	return &EventHandler{
		log:      log,
		strgPG:   strgPG,
		strgES:   strgES,
		producer: producer,
	}
}

func (h *EventHandler) Push(topic string, value interface{}) error {

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	deliveryChan := make(chan kafka.Event, 10000)
	err = h.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: data,
	}, deliveryChan)
	if err != nil {
		return err
	}

	return nil

}

package handlers

import (
	"context"
	"encoding/json"
	"genproto/common"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func (e *EventHandler) UpsertShop(ctx context.Context, event *kafka.Message) error {

	var req common.ShopCreatedModel

	if err := json.Unmarshal(event.Value, &req); err != nil {
		return err

	}

	// e.log.info("shop created", logger.Any("event", req))

	if err := e.strgPG.Shop().Upsert(&req); err != nil {
		return err
	}

	return nil

}

func (e *EventHandler) DeleteShop(ctx context.Context, event *kafka.Message) error {

	var req common.RequestID

	if err := json.Unmarshal(event.Value, &req); err != nil {
		return err

	}

	if err := e.strgPG.Shop().Delete(&req); err != nil {
		return err
	}

	return nil

}

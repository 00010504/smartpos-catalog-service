package handlers

import (
	"context"
	"encoding/json"
	"genproto/common"
	"genproto/inventory_service"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func (e *EventHandler) UpsertSupplier(ctx context.Context, event *kafka.Message) error {

	var req inventory_service.SupplierCreateModel

	if err := json.Unmarshal(event.Value, &req); err != nil {
		return err
	}

	if err := e.strgPG.Supplier().UpsertSupplier(&req); err != nil {
		return err
	}

	return nil

}

func (e *EventHandler) DeleteSupplier(ctx context.Context, event *kafka.Message) error {

	var req common.RequestID

	if err := json.Unmarshal(event.Value, &req); err != nil {
		return err
	}

	if err := e.strgPG.Supplier().Delete(&req); err != nil {
		return err
	}

	return nil

}

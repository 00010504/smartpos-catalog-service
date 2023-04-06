package handlers

import (
	"context"
	"encoding/json"
	"genproto/catalog_service"

	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func (e *EventHandler) UpsertMeasurementValue(ctx context.Context, event *kafka.Message) error {

	var request catalog_service.UpsertShopMeasurmentValueRequest

	tr, err := e.strgPG.WithTransaction()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tr.Rollback()
		} else {
			_ = tr.Commit()
		}
	}()

	if err := json.Unmarshal(event.Value, &request); err != nil {
		return err
	}
	e.log.Info("UpsertMeasurementValue", logger.Any("event", request))

	if err := e.strgPG.Product().UpsertShopMeasurmentValue(&request); err != nil {
		return err
	}

	if err := e.strgES.Product().UpsertShopMeasurmentValue(&request); err != nil {
		return err
	}

	return nil
}

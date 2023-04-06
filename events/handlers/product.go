package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"genproto/catalog_service"
	"genproto/common"

	"github.com/Invan2/invan_catalog_service/pkg/telegram"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
)

func (e *EventHandler) CreateMultipleProducts(ctx context.Context, event *kafka.Message) error {

	var req common.CreateImportProductsModel

	if err := json.Unmarshal(event.Value, &req); err != nil {
		return errors.Wrap(err, "error while unmarshal req")
	}

	go telegram.SendNewMessage(fmt.Sprintf("productlar soni: %d", len(req.Products)))

	fmt.Println("products count", len(req.Products))

	if len(req.Products) <= 0 {
		return nil
	}

	tr, err := e.strgPG.WithTransaction()
	if err != nil {
		return errors.Wrap(err, "error while run transaction")
	}

	defer func() {
		if err != nil {
			_ = tr.Rollback()
		} else {
			_ = tr.Commit()
		}
	}()

	err = tr.Product().InsertMany(req.Products)
	if err != nil {
		return err
	}

	err = e.strgES.Product().InsertMany(req.Products)
	if err != nil {
		return err
	}

	err = e.Push("v1.inventory_service.create_multiple_products_on_order_service", req)
	if err != nil {
		return err
	}

	return nil
}

func (e *EventHandler) UpdateShopPrice(ctx context.Context, event *kafka.Message) error {
	var req catalog_service.UpsertShopPriceRequest

	if err := json.Unmarshal(event.Value, &req); err != nil {
		return errors.Wrap(err, "error while unmarshal req")
	}

	tr, err := e.strgPG.WithTransaction()
	if err != nil {
		return errors.Wrap(err, "error while run transaction")
	}

	defer func() {
		if err != nil {
			_ = tr.Rollback()
		} else {
			_ = tr.Commit()
		}
	}()

	err = tr.Product().UpsertShopRetailPrice(&req)
	if err != nil {
		return err
	}

	err = e.strgES.Product().UpsertShopPrice(&req)
	if err != nil {
		return err
	}

	return nil
}

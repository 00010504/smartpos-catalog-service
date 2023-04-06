package handlers

import (
	"context"
	"encoding/json"
	"genproto/catalog_service"
	"genproto/common"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
)

func (e *EventHandler) CreateCompany(ctx context.Context, event *kafka.Message) error {

	var (
		req                     common.CompanyCreatedModel
		kafka_measurement_units []*common.MeasurementUnitCopyRequest
	)

	if err := json.Unmarshal(event.Value, &req); err != nil {
		return errors.Wrap(err, "error while unmarshaling company")
	}

	// e.log.info("create company event", logger.Any("event", req))

	if err := e.strgPG.Company().Upsert(&req); err != nil {
		// e.log.info(err.Error(), logger.Any("event", req))

		return err
	}

	// push measurement_unit

	userReq := &common.Request{CompanyId: req.Id, UserId: req.CreatedBy}

	// e.log.info("getting measurementUnits")

	measurement_units, err := e.strgPG.MeasurementUnit().GetAll(&catalog_service.GetAllMeasurementUnitsRequest{Limit: 10, Page: 1, Request: userReq})
	if err != nil {
		// e.log.info("error measurement_units getAll", logger.Error(err))
		return err
	}

	for _, measurement_unit := range measurement_units.Data {
		kafka_measurement_units = append(kafka_measurement_units, &common.MeasurementUnitCopyRequest{
			Id:                   measurement_unit.Id,
			CompanyId:            req.Id,
			IsDeletable:          measurement_unit.IsDeletable,
			ShortName:            measurement_unit.ShortName,
			LongName:             measurement_unit.LongName,
			Precision:            measurement_unit.Precision.Value,
			CreatedBy:            req.CreatedBy,
			Request:              userReq,
			LongNameTranslation:  measurement_unit.LongNameTranslation,
			ShortNameTranslation: measurement_unit.ShortNameTranslation,
		})
	}

	if err := e.Push("v1.catalog_service.measurement_units.created.success", common.MeasurementUnitsCopyRequest{
		MeasurementUnits: kafka_measurement_units,
		Request:          userReq,
	}); err != nil {
		return err
	}

	// e.log.info("company shop is about to create", logger.Any("event", req))

	if req.Shop == nil {
		return errors.New("error while create copmany shop. shop == nil")
	}

	if err := e.strgPG.Shop().Upsert(req.Shop); err != nil {
		return err
	}

	return nil

}

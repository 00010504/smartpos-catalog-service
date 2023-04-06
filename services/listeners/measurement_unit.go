package listeners

import (
	"context"
	"genproto/common"

	"genproto/catalog_service"
)

func (c *catalogService) CreateMeasurementUnit(ctx context.Context, req *catalog_service.CreateMeasurementUnitRequest) (*common.ResponseID, error) {

	tr, err := c.strg.WithTransaction()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tr.Rollback()
		} else {
			_ = tr.Commit()
		}
	}()

	id, err := tr.MeasurementUnit().Create(req)
	if err != nil {
		return nil, err
	}

	measurementUnit, err := tr.MeasurementUnit().GetByID(&common.RequestID{Id: id, Request: req.Request})
	if err != nil {
		return nil, err
	}

	err = c.kafka.Push("v1.catalog_service.measurement_unit.created.success", common.MeasurementUnitCopyRequest{
		Id:                   measurementUnit.Id,
		CompanyId:            req.Request.CompanyId,
		IsDeletable:          measurementUnit.IsDeletable,
		ShortName:            measurementUnit.ShortName,
		LongName:             measurementUnit.LongName,
		Precision:            measurementUnit.Precision.Value,
		CreatedBy:            measurementUnit.CreatedBy.Id,
		Request:              req.Request,
		LongNameTranslation:  measurementUnit.LongNameTranslation,
		ShortNameTranslation: measurementUnit.ShortNameTranslation,
	})
	if err != nil {
		return nil, err
	}

	return &common.ResponseID{Id: id}, nil
}

func (c *catalogService) GetMeasurementUnitByID(ctx context.Context, req *common.RequestID) (*catalog_service.MeasurementUnit, error) {

	measurementUnit, err := c.strg.MeasurementUnit().GetByID(req)
	if err != nil {
		return nil, err
	}

	return measurementUnit, nil
}

func (c *catalogService) UpdateMeasurementUnit(ctx context.Context, req *catalog_service.UpdateMeasurementUnitRequest) (*common.ResponseID, error) {

	tr, err := c.strg.WithTransaction()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tr.Rollback()
		} else {
			_ = tr.Commit()
		}
	}()

	id, err := tr.MeasurementUnit().Update(req)
	if err != nil {
		return nil, err
	}

	measurementUnit, err := tr.MeasurementUnit().GetByID(&common.RequestID{Id: req.Id, Request: req.Request})
	if err != nil {
		return nil, err
	}

	err = c.kafka.Push("v1.catalog_service.measurement_unit.created.success", common.MeasurementUnitCopyRequest{
		Id:                   measurementUnit.Id,
		CompanyId:            req.Request.CompanyId,
		IsDeletable:          measurementUnit.IsDeletable,
		ShortName:            measurementUnit.ShortName,
		LongName:             measurementUnit.LongName,
		Precision:            measurementUnit.Precision.Value,
		CreatedBy:            measurementUnit.CreatedBy.Id,
		Request:              req.Request,
		LongNameTranslation:  measurementUnit.LongNameTranslation,
		ShortNameTranslation: measurementUnit.ShortNameTranslation,
	})
	if err != nil {
		return nil, err
	}

	return &common.ResponseID{Id: id}, nil
}

func (c *catalogService) GetAllMeasurementUnits(ctx context.Context, req *catalog_service.GetAllMeasurementUnitsRequest) (*catalog_service.GetAllMeasurementUnitsResponse, error) {
	res, err := c.strg.MeasurementUnit().GetAll(req)
	if err != nil {
		return nil, err
	}

	return res, nil

}

func (c *catalogService) DeleteMeasurementUnitById(ctx context.Context, req *common.RequestID) (*common.ResponseID, error) {

	_, err := c.strg.MeasurementUnit().Delete(req)
	if err != nil {
		return nil, err
	}

	return &common.ResponseID{Id: req.Id}, nil
}

func (c *catalogService) GetAllDefaultUnits(ctx context.Context, req *common.SearchRequest) (*catalog_service.GetAllDefaultUnitsResponse, error) {
	return c.strg.MeasurementUnit().GetAllDefaultUnits(ctx, req)
}

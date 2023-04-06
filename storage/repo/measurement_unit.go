package repo

import (
	"context"
	"genproto/catalog_service"
	"genproto/common"
)

type MeasurementUnitPgI interface {
	Create(entity *catalog_service.CreateMeasurementUnitRequest) (string, error)
	GetByID(entity *common.RequestID) (*catalog_service.MeasurementUnit, error)
	Update(entity *catalog_service.UpdateMeasurementUnitRequest) (string, error)
	Delete(req *common.RequestID) (*common.ResponseID, error)
	GetAll(req *catalog_service.GetAllMeasurementUnitsRequest) (*catalog_service.GetAllMeasurementUnitsResponse, error)
	GetAllDefaultUnits(ctx context.Context, req *common.SearchRequest) (*catalog_service.GetAllDefaultUnitsResponse, error)
}

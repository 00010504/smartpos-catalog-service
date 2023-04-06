package repo

import (
	"genproto/catalog_service"
	"genproto/common"
	"genproto/inventory_service"
)

type SupplierI interface {
	UpsertSupplier(entity *inventory_service.SupplierCreateModel) error
	GetById(req *common.RequestID) (*catalog_service.ShortSupplier, error)
	Delete(req *common.RequestID) error
}

package repo

import (
	"context"
	"genproto/catalog_service"
	"genproto/common"
)

type VatI interface {
	Create(req *catalog_service.CreateVatRequest) (*common.ResponseID, error)
	GetById(ctx context.Context, req *common.RequestID) (*catalog_service.GetVatByIdResponse, error)
	Update(ctx context.Context, req *catalog_service.UpdateVatRequest) (*common.ResponseID, error)
	GetAll(ctx context.Context, req *common.SearchRequest) (*catalog_service.GetAllVatsResponse, error)
	Delete(req *common.RequestID) (*common.ResponseID, error)
}

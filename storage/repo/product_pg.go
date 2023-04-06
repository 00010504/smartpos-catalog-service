package repo

import (
	"genproto/catalog_service"
	"genproto/common"

	"github.com/Invan2/invan_catalog_service/models"
)

type ProductPgI interface {
	Create(entity *catalog_service.CreateProductRequest) (productId string, productDetailId string, err error)
	InsertMany([]*common.CreateProductCopyRequest) error
	GetByID(req *common.RequestID) (*catalog_service.Product, error)
	Update(entity *catalog_service.UpdateProductRequest) (*common.ResponseID, error)
	UpsertShopMeasurmentValue(req *catalog_service.UpsertShopMeasurmentValueRequest) error
	// GetProductCategories(productDetailId string) ([]*catalog_service.ShortCategory, error)
	Delete(req *common.RequestID) (*common.ResponseID, error)
	DeleteProducts(entity *common.RequestIDs) (*common.Empty, error)
	GetProductCustomFields(req *common.Request) ([]*models.GetProductCustomFieldResponse, error)
	UpsertShopRetailPrice(req *catalog_service.UpsertShopPriceRequest) error
	ProductBulkEdit(req *catalog_service.ProductBulkOperationRequest) (*common.ResponseID, error)
}

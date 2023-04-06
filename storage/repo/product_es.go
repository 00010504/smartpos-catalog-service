package repo

import (
	"genproto/catalog_service"
	"genproto/common"

	"github.com/Invan2/invan_catalog_service/models"
)

type ProductESI interface {
	Create(product *catalog_service.ProductES) error
	InsertMany([]*common.CreateProductCopyRequest) error
	Update(product *catalog_service.ProductES) error
	UpsertShopMeasurmentValue(supplierOrder *catalog_service.UpsertShopMeasurmentValueRequest) error
	GetAll(req *catalog_service.GetAllProductsRequest) (*catalog_service.GetAllProductsResponse, error)
	GetForLabel(req *catalog_service.GetProductLabelsRequest) (*catalog_service.GetAllProductsResponse, error)
	SearchProducts(entity *catalog_service.GetAllProductsRequest) (*catalog_service.SearchProductsResponse, error)
	DeleteProduct(*common.RequestID) (*common.Empty, error)
	DeleteProducts(*common.RequestIDs) (*common.Empty, error)
	GetAllForExcel(req *catalog_service.GetAllProductsRequest) (*models.GetAllForExcelResponse, error)
	GetAllForCSV(req *catalog_service.GetAllProductsRequest) (*models.GetAllForCsvResponse, error)
	UpsertShopPrice(req *catalog_service.UpsertShopPriceRequest) error
	BulkUpdateProduct(req *catalog_service.ProductBulkOperationRequest, productMap map[string]*catalog_service.ProductES) error
}

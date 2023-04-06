package listeners

import (
	"context"
	"genproto/common"

	"genproto/catalog_service"

	pdfmaker "github.com/Invan2/invan_catalog_service/pkg/pdf"

	"github.com/Invan2/invan_catalog_service/config"
	"github.com/Invan2/invan_catalog_service/events"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/Invan2/invan_catalog_service/storage"
	"github.com/minio/minio-go/v7"
)

type catalogService struct {
	log     logger.Logger
	kafka   events.PubSubServer
	strg    storage.StoragePg
	elastic storage.StorageES
	minio   *minio.Client
	cfg     *config.Config
	pdf     pdfmaker.PdFMakekerI
}

type CatalogService interface {
	Ping(ctx context.Context, message *common.PingPong) (*common.PingPong, error)

	//product
	CreateProduct(ctx context.Context, req *catalog_service.CreateProductRequest) (*common.ResponseID, error)
	GetProductByID(ctx context.Context, req *common.RequestID) (*catalog_service.Product, error)
	UpdateProduct(ctx context.Context, req *catalog_service.UpdateProductRequest) (*common.ResponseID, error)
	GetAllProducts(ctx context.Context, req *catalog_service.GetAllProductsRequest) (*catalog_service.GetAllProductsResponse, error)
	DeleteProductById(ctx context.Context, req *common.RequestID) (*common.ResponseID, error)
	SearchProducts(ctx context.Context, req *catalog_service.GetAllProductsRequest) (*catalog_service.SearchProductsResponse, error)
	DeleteProductsByIds(ctx context.Context, req *common.RequestIDs) (*common.Empty, error)
	BulkUpdateProduct(ctx context.Context, req *catalog_service.ProductBulkOperationRequest) (*common.ResponseID, error)

	BulkGenerateProductLabels(ctx context.Context, req *catalog_service.GetProductLabelsRequest) (*common.ResponseID, error)

	// measurementUnit
	CreateMeasurementUnit(ctx context.Context, req *catalog_service.CreateMeasurementUnitRequest) (*common.ResponseID, error)
	GetMeasurementUnitByID(ctx context.Context, req *common.RequestID) (*catalog_service.MeasurementUnit, error)
	UpdateMeasurementUnit(ctx context.Context, req *catalog_service.UpdateMeasurementUnitRequest) (*common.ResponseID, error)
	GetAllMeasurementUnits(ctx context.Context, req *catalog_service.GetAllMeasurementUnitsRequest) (*catalog_service.GetAllMeasurementUnitsResponse, error)
	DeleteMeasurementUnitById(ctx context.Context, req *common.RequestID) (*common.ResponseID, error)
	GetAllDefaultUnits(ctx context.Context, req *common.SearchRequest) (*catalog_service.GetAllDefaultUnitsResponse, error)

	//category
	CreateCategory(ctx context.Context, req *catalog_service.CreateCategoryRequest) (*common.ResponseID, error)
	GetCategoryByID(ctx context.Context, req *common.RequestID) (*catalog_service.GetCategoryByIDResponse, error)
	UpdateCategory(ctx context.Context, req *catalog_service.UpdateCategoryRequest) (*common.ResponseID, error)
	GetAllCategories(ctx context.Context, req *catalog_service.GetAllCategoriesRequest) (*catalog_service.GetAllCategoriesResponse, error)
	DeleteCategoryById(ctx context.Context, req *common.RequestID) (*common.ResponseID, error)

	// label
	CreateLabel(context.Context, *catalog_service.CreateLabelRequest) (*common.ResponseID, error)
	GetLabelById(context.Context, *common.RequestID) (*catalog_service.GetLabelResponse, error)
	UpdateLabelById(context.Context, *catalog_service.UpdateLabelRequest) (*common.ResponseID, error)
	GetAllLabels(context.Context, *common.SearchRequest) (*catalog_service.GetAllLabelsResponse, error)
	DeleteLabelById(context.Context, *common.RequestID) (*common.ResponseID, error)
	DeleteLabelsByIds(context.Context, *common.RequestIDs) (*common.Empty, error)
	GetProductFields(ctx context.Context, req *catalog_service.GetProductFieldsRequest) (*catalog_service.GetProductFieldsResponse, error)

	// exel_template
	CreateExelTemplate(ctx context.Context, req *common.Request) (*common.ResponseID, error)
	CreateProductExelTemplate(ctx context.Context, req *catalog_service.GetProductExcelDownloadRequest) (*common.ResponseID, error)
	CreateProductCsvTemplate(ctx context.Context, req *catalog_service.GetProductCsvDownloadRequest) (*common.ResponseID, error)

	// Scales_template
	CreateScalesTemplates(context.Context, *catalog_service.CreateScalesTemplateRequest) (*common.ResponseID, error)
	GetScalesTemplateByID(context.Context, *catalog_service.GetScalesTemplateByIDRequest) (*catalog_service.ScalesTemplate, error)
	GetAllScalesTemplates(context.Context, *catalog_service.GetAllScalesTemplatesRequest) (*catalog_service.GetAllScalesTemplatesResponse, error)

	// VAT
	CreateVat(ctx context.Context, req *catalog_service.CreateVatRequest) (*common.ResponseID, error)
	GetVatById(ctx context.Context, req *common.RequestID) (*catalog_service.GetVatByIdResponse, error)
	UpdateVatById(ctx context.Context, req *catalog_service.UpdateVatRequest) (*common.ResponseID, error)
	GetAllVats(ctx context.Context, req *common.SearchRequest) (*catalog_service.GetAllVatsResponse, error)
	DeleteVat(ctx context.Context, req *common.RequestID) (*common.ResponseID, error)
}

func NewCatalogService(log logger.Logger, kafka events.PubSubServer, strg storage.StoragePg, elastic storage.StorageES, minio *minio.Client, cfg *config.Config) CatalogService {
	return &catalogService{
		log:     log,
		kafka:   kafka,
		strg:    strg,
		elastic: elastic,
		minio:   minio,
		cfg:     cfg,
		pdf:     pdfmaker.NewPdfMaker(log, minio, cfg),
	}
}

package pdfmaker

import (
	"genproto/catalog_service"

	"github.com/Invan2/invan_catalog_service/config"
	"github.com/Invan2/invan_catalog_service/pkg/logger"

	"github.com/minio/minio-go/v7"
)

type PdFmaker struct {
	log         logger.Logger
	cfg         *config.Config
	minioClient *minio.Client
}

type PdFMakekerI interface {
	HTMLtoPDF(htmlPath string, pdfName string) (string, error)
	MakeProductsLabel(products []map[string]interface{}, label *catalog_service.GetLabelResponse) (string, error)
	MakeProductsPriceTag(products []map[string]interface{}, label *catalog_service.GetLabelResponse) (string, error)
	MakeProductsPriceReceipt(products []map[string]string, label *catalog_service.GetLabelResponse) (string, error)
}

func NewPdfMaker(log logger.Logger, minioClient *minio.Client, config *config.Config) PdFMakekerI {

	return &PdFmaker{
		log:         log,
		minioClient: minioClient,
		cfg:         config,
	}
}

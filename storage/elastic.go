package storage

import (
	"github.com/Invan2/invan_catalog_service/config"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/Invan2/invan_catalog_service/storage/elastic"
	"github.com/Invan2/invan_catalog_service/storage/repo"
	"github.com/elastic/go-elasticsearch/v8"
)

type storageES struct {
	db          *elasticsearch.Client
	log         logger.Logger
	productRepo repo.ProductESI
}

type StorageES interface {
	Product() repo.ProductESI
}

func NewStorageES(log logger.Logger, db *elasticsearch.Client, cfg config.Config) StorageES {
	return &storageES{
		db:          db,
		log:         log,
		productRepo: elastic.NewProductRepo(log, db, cfg),
	}
}

func (s *storageES) Product() repo.ProductESI {
	return s.productRepo
}

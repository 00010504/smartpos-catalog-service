package services

import (
	"github.com/Invan2/invan_catalog_service/config"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
)

type gRPCServices struct {
	log logger.Logger
	cfg *config.Config
}

type GRPCServices interface {
}

func NewGrpcServices(log logger.Logger, cfg *config.Config) (GRPCServices, error) {

	res := &gRPCServices{
		log: log,
		cfg: cfg,
	}

	return res, nil
}

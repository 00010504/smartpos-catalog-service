package repo

import (
	"genproto/catalog_service"
	"genproto/common"
)

type LabelI interface {
	Create(*catalog_service.CreateLabelRequest) (*common.ResponseID, error)
	GetById(req *common.RequestID) (*catalog_service.GetLabelResponse, error)
	UpdateById(req *catalog_service.UpdateLabelRequest) (*common.ResponseID, error)
	GetAll(req *common.SearchRequest) (*catalog_service.GetAllLabelsResponse, error)
	DeleteLabelById(*common.RequestID) (*common.ResponseID, error)
	DeleteLabelsByIds(*common.RequestIDs) (*common.Empty, error)
}

package repo

import (
	"genproto/catalog_service"
	"genproto/common"
)

type ScalesTemplateI interface {
	CreateScalesTemplates(*catalog_service.CreateScalesTemplateRequest) (res *common.ResponseID, err error)
	GetScalesTemplateByID(*catalog_service.GetScalesTemplateByIDRequest) (res *catalog_service.ScalesTemplate, err error)
	GetAllScalesTemplates(*catalog_service.GetAllScalesTemplatesRequest) (res *catalog_service.GetAllScalesTemplatesResponse, err error)
}

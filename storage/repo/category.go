package repo

import (
	"genproto/catalog_service"
	"genproto/common"
)

type CategoryPgI interface {
	Create(entity *catalog_service.CreateCategoryRequest) (string, error)
	GetByID(entiy *common.RequestID) (*catalog_service.GetCategoryByIDResponse, error)
	Update(entity *catalog_service.UpdateCategoryRequest) (string, error)
	Delete(req *common.RequestID) (*common.ResponseID, error)
	GetAll(req *catalog_service.GetAllCategoriesRequest) (*catalog_service.GetAllCategoriesResponse, error)
	GetShortCategoriesByIds(ids []string) ([]*catalog_service.ShortCategory, error)
}

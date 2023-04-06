package repo

import (
	"genproto/common"

	"github.com/Invan2/invan_catalog_service/models"
)

type ShopI interface {
	Upsert(*common.ShopCreatedModel) error
	Delete(*common.RequestID) error
	GetAll(req *models.GetShopsReq) ([]*models.GetShopNameRespone, error)
	GetCompanyAllShopNames(req *common.Request) (map[string]string, error)
}

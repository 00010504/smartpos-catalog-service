package repo

import (
	"genproto/common"
)

type CompanyPgI interface {
	Upsert(entity *common.CompanyCreatedModel) error
	Delete(req *common.RequestID) (*common.ResponseID, error)
}

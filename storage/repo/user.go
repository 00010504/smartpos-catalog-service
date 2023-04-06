package repo

import (
	"genproto/common"
)

type UserPgI interface {
	Upsert(entity *common.UserCreatedModel) error
	Delete(req *common.RequestID) (*common.ResponseID, error)
}

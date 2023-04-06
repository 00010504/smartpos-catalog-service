package postgres

import (
	"genproto/common"

	"github.com/Invan2/invan_catalog_service/models"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/Invan2/invan_catalog_service/storage/repo"
	"github.com/pkg/errors"
)

type userRepo struct {
	db  models.DB
	log logger.Logger
}

func NewUserRepo(log logger.Logger, db models.DB) repo.UserPgI {
	return &userRepo{
		db:  db,
		log: log,
	}
}

func (u *userRepo) Upsert(entity *common.UserCreatedModel) error {

	query := `
		INSERT INTO
			"user"
		(
			id,
			user_type_id,
			first_name,
			last_name,
			phone_number,
			image
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6
		) ON CONFLICT (id) DO
		UPDATE
			SET
			user_type_id = $2,
			first_name = $3,
			last_name = $4,
			phone_number = $5,
			image = $6;
	`

	_, err := u.db.Exec(
		query,
		entity.Id,
		entity.UserTypeId,
		entity.FirstName,
		entity.LastName,
		entity.PhoneNumber,
		entity.Image,
	)
	if err != nil {
		return errors.Wrap(err, "error while insert user")
	}

	return nil
}

func (u *userRepo) Delete(req *common.RequestID) (*common.ResponseID, error) {

	query := `
	  	UPDATE
			"user"
	  	SET
			deleted_at = extract(epoch from now())::bigint
	  	WHERE
			id = $1 AND deleted_at = 0
	`

	res, err := u.db.Exec(
		query,
		req.Id,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error while delete user")
	}

	i, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if i == 0 {
		return nil, errors.New("user not found")
	}

	return &common.ResponseID{Id: req.Id}, nil
}

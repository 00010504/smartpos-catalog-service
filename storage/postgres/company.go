package postgres

import (
	"genproto/common"

	"github.com/Invan2/invan_catalog_service/models"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/Invan2/invan_catalog_service/storage/repo"
	"github.com/pkg/errors"
)

type companyRepo struct {
	db  models.DB
	log logger.Logger
}

func NewCompanyRepo(log logger.Logger, db models.DB) repo.CompanyPgI {
	return &companyRepo{
		db:  db,
		log: log,
	}
}

func (c *companyRepo) Upsert(entity *common.CompanyCreatedModel) error {

	query := `
		INSERT INTO
			"company"
		(
			id,
			name,
			created_by
		)
		VALUES (
			$1,
			$2,
			$3
		) ON CONFLICT (id) DO
		UPDATE
			SET
			name = $2;
	`

	_, err := c.db.Exec(
		query,
		entity.Id,
		entity.Name,
		entity.CreatedBy,
	)

	if err != nil {
		return errors.Wrap(err, "error while insert company")
	}

	return nil
}

func (c *companyRepo) Delete(req *common.RequestID) (*common.ResponseID, error) {

	query := `
	  	UPDATE
			"company"
	  	SET
			deleted_at = extract(epoch from now())::bigint
	  	WHERE
			id = $1 AND deleted_at = 0
	`

	res, err := c.db.Exec(
		query,
		req.Id,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error while delete company")
	}

	i, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if i == 0 {
		return nil, errors.New("company not found")
	}

	return &common.ResponseID{Id: req.Id}, nil
}

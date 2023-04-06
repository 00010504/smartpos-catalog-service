package postgres

import (
	"genproto/common"

	"github.com/Invan2/invan_catalog_service/models"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/Invan2/invan_catalog_service/storage/repo"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type shopRepo struct {
	db  models.DB
	log logger.Logger
}

func NewShopRepo(log logger.Logger, db models.DB) repo.ShopI {
	return &shopRepo{
		db:  db,
		log: log,
	}
}

func (s *shopRepo) Upsert(entity *common.ShopCreatedModel) error {

	query := `
		INSERT INTO
			"shop"
		(
			id,
			name,
			company_id,
			created_by
		)
		VALUES (
			$1,
			$2,
			$3,
			$4
		) ON CONFLICT (id) DO
		UPDATE
			SET
			"name" = $2,
			"company_id" = $3
			;
	`

	_, err := s.db.Exec(
		query,
		entity.Id,
		entity.Name,
		entity.Request.CompanyId,
		entity.Request.UserId,
	)
	if err != nil {
		return errors.Wrap(err, "error while insert shop")
	}

	return nil
}

func (s *shopRepo) Delete(req *common.RequestID) error {
	query := `
		UPDATE "shop" SET deleted_at=extract(epoch from now())::bigint
		WHERE id=$1 AND deleted_at=0
	`

	_, err := s.db.Exec(query, req.Id)
	if err != nil {
		return err
	}
	return nil
}

func (c *shopRepo) GetAll(req *models.GetShopsReq) ([]*models.GetShopNameRespone, error) {

	var res []*models.GetShopNameRespone

	query := `
		SELECT
			id,
			name
		FROM "shop"
		WHERE
			id = ANY($1) AND company_id = $2 AND
			deleted_at = 0
	`
	rows, err := c.db.Query(query, pq.Array(req.ShopIds), req.CompanyId)
	if err != nil {
		return nil, errors.Wrap(err, "error while get shops. Query")
	}

	defer rows.Close()

	for rows.Next() {

		var shop models.GetShopNameRespone

		err = rows.Scan(&shop.Id, &shop.Name)
		if err != nil {
			return nil, errors.Wrap(err, "error while get shops. Scan")
		}

		res = append(res, &shop)
	}

	return res, nil
}

func (c *shopRepo) GetCompanyAllShopNames(req *common.Request) (map[string]string, error) {

	var res = make(map[string]string)

	query := `
		SELECT id, name
		FROM "shop"
		WHERE company_id = $1 AND deleted_at = 0
	`
	rows, err := c.db.Query(query, req.CompanyId)
	if err != nil {
		return nil, errors.Wrap(err, "error while get shops. Query")
	}

	defer rows.Close()

	for rows.Next() {

		var shop models.GetShopNameRespone

		err = rows.Scan(&shop.Id, &shop.Name)
		if err != nil {
			return nil, errors.Wrap(err, "error while get shops. Scan")
		}

		res[shop.Id] = shop.Name
	}

	return res, nil
}

package postgres

import (
	"genproto/catalog_service"
	"genproto/common"
	"genproto/inventory_service"

	"github.com/Invan2/invan_catalog_service/config"
	"github.com/Invan2/invan_catalog_service/models"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/Invan2/invan_catalog_service/storage/repo"
	"github.com/pkg/errors"
)

type SupplierPgI struct {
	db  models.DB
	log logger.Logger
	cfg config.Config
}

func NewSupplierRepo(log logger.Logger, db models.DB, cfg config.Config) repo.SupplierI {
	return &SupplierPgI{
		db:  db,
		log: log,
		cfg: cfg,
	}
}

func (s *SupplierPgI) UpsertSupplier(entity *inventory_service.SupplierCreateModel) error {

	query := `
		INSERT INTO
			"supplier"
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
			"company_id" = $3;
	`

	_, err := s.db.Exec(
		query,
		entity.Id,
		entity.SupplierCompanyName,
		entity.Request.CompanyId,
		entity.Request.UserId,
	)
	if err != nil {
		return errors.Wrap(err, "error while supplier")
	}

	return nil
}

func (s *SupplierPgI) GetById(req *common.RequestID) (*catalog_service.ShortSupplier, error) {

	var (
		res catalog_service.ShortSupplier
	)

	query := `
		SELECT
			id,
			name
		FROM "supplier"
		WHERE id = $1 AND deleted_at = 0 AND company_id = $2
	`

	err := s.db.QueryRow(query, req.Id, req.Request.CompanyId).Scan(
		&res.Id,
		&res.Name,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error while getting supplier")
	}
	return &res, nil
}

func (s *SupplierPgI) Delete(req *common.RequestID) error {

	query := `
		UPDATE 
			"supplier"
		SET
			deleted_at = extract(epoch from now())::bigint
		WHERE 
			id = $1 AND deleted_at = 0
		`
	_, err := s.db.Exec(query, req.Id)
	if err != nil {
		return errors.Wrap(err, "error while delete supplier")
	}

	return nil
}

package postgres

import (
	"context"
	"genproto/catalog_service"
	"genproto/common"

	"github.com/Invan2/invan_catalog_service/config"
	"github.com/Invan2/invan_catalog_service/models"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/Invan2/invan_catalog_service/storage/repo"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type vatRepo struct {
	db  models.DB
	log logger.Logger
	cfg config.Config
}

func NewVatRepo(log logger.Logger, db models.DB, cfg config.Config) repo.VatI {
	return &vatRepo{
		db:  db,
		log: log,
		cfg: cfg,
	}
}

func (v *vatRepo) Create(req *catalog_service.CreateVatRequest) (*common.ResponseID, error) {
	var (
		vatId = uuid.NewString()
	)

	query := `
		INSERT INTO "vat"
			(
				"id",
				"name",
				"percentage",
				"company_id",
				"created_by"
			)
		VALUES
			(
				$1,
				$2,
				$3,
				$4,
				$5
			)
	`
	_, err := v.db.Exec(
		query,
		vatId,
		req.Name,
		req.Percentage,
		req.Request.CompanyId,
		req.Request.UserId,
	)

	if err != nil {
		return nil, errors.Wrap(err, "error while create vat")
	}
	return &common.ResponseID{Id: vatId}, nil
}

func (v *vatRepo) GetById(ctx context.Context, req *common.RequestID) (*catalog_service.GetVatByIdResponse, error) {

	var (
		res catalog_service.GetVatByIdResponse
	)

	query := `
		SELECT
			id,
			name,
			percentage
		FROM "vat"
		WHERE id = $1 AND company_id = $2 AND deleted_at = 0
	`
	err := v.db.QueryRow(query, req.Id, req.Request.CompanyId).Scan(
		&res.Id,
		&res.Name,
		&res.Percentage,
	)   

	if err != nil {
		return nil, errors.Wrap(err, "error while getting vat")
	}
	return &res, nil
}

func (v *vatRepo) Update(ctx context.Context, req *catalog_service.UpdateVatRequest) (*common.ResponseID, error) {

	query := `
		UPDATE
			"vat"
		SET
			name = $2,
			percentage = $3
		WHERE id = $1 AND company_id = $4 AND deleted_at = 0
	`
	_, err := v.db.Exec(
		query,
		req.Id,
		req.Name,
		req.Percentage,
		req.Request.CompanyId,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error while update vat")

	}
	return &common.ResponseID{Id: req.Id}, nil
}

func (v *vatRepo) GetAll(ctx context.Context, req *common.SearchRequest) (*catalog_service.GetAllVatsResponse, error) {

	var (
		res = catalog_service.GetAllVatsResponse{
			Data:  make([]*catalog_service.GetVatResponse, 0),
			Total: 0,
		}
		values = map[string]interface{}{
			"limit":      req.Limit,
			"offset":     req.Limit * (req.Page - 1),
			"search":     req.Search,
			"company_id": req.Request.CompanyId,
		}
	)

	query := `
		SELECT 
			v.id,
			v.name,
			v.percentage
		FROM "vat" v
	`
	filter := ` WHERE v.company_id=:company_id AND v.deleted_at = 0 `
	if req.Search != "" {
		filter += `  AND (
		v."name" ILIKE '%' || :search || '%'
		) 
		`
	}

	query += filter + `
		ORDER BY  v.created_at DESC
		LIMIT :limit
		OFFSET :offset
	`

	rows, err := v.db.NamedQuery(query, values)
	if err != nil {
		return nil, errors.Wrap(err, "error while search")
	}

	defer rows.Close()

	for rows.Next() {

		var (
			vat = catalog_service.GetVatResponse{}
		)

		err = rows.Scan(&vat.Id, &vat.Name, &vat.Percentage)
		if err != nil {
			return nil, errors.Wrap(err, "error while scanning all vats")
		}

		res.Data = append(res.Data, &vat)
	}

	query = `
		SELECT
			count(v.id)
		from "vat" v

	` + filter

	stmt, err := v.db.PrepareNamed(query)
	if err != nil {
		return nil, errors.Wrap(err, "error while prepareName")
	}

	defer stmt.Close()

	err = stmt.QueryRow(values).Scan(&res.Total)
	if err != nil {
		return nil, errors.Wrap(err, "error while scanning queryRow")
	}

	return &res, nil
}

func (v *vatRepo) Delete(req *common.RequestID) (*common.ResponseID, error) {

	query := `
	  	UPDATE
			"vat"
	  	SET
			deleted_at = extract(epoch from now())::bigint
	  	WHERE
			id = $1 AND deleted_at = 0 AND company_id = $2
	`

	res, err := v.db.Exec(
		query,
		req.Id,
		req.Request.CompanyId,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error while delete vat")
	}

	i, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if i == 0 {
		return nil, errors.Wrap(errors.New("vat not found"), "error while delete vat rowsAffected = 0")
	}

	return &common.ResponseID{Id: req.Id}, nil
}

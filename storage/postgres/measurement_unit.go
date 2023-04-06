package postgres

import (
	"context"
	"encoding/json"
	"genproto/catalog_service"
	"genproto/common"

	"github.com/Invan2/invan_catalog_service/models"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/Invan2/invan_catalog_service/storage/repo"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type measurementUnitRepo struct {
	db  models.DB
	log logger.Logger
}

func NewMeasurementUnitRepo(log logger.Logger, db models.DB) repo.MeasurementUnitPgI {
	return &measurementUnitRepo{
		db:  db,
		log: log,
	}
}

func (p *measurementUnitRepo) Create(entity *catalog_service.CreateMeasurementUnitRequest) (string, error) {

	id := uuid.New().String()

	query := `
		INSERT INTO
			"measurement_unit"
		(
			id,
			unit_id,
			precision_id,
			company_id,
			created_by
		)
		VALUES
		(
			$1,
			$2,
			$3,
			$4,
			$5
	  	);
	`

	_, err := p.db.Exec(
		query,
		id,
		entity.UnitId,
		entity.PrecisionId,
		entity.Request.CompanyId,
		entity.Request.UserId,
	)
	if err != nil {
		return "", errors.Wrap(err, "error while insert measurement_unit")
	}

	return id, nil
}

func (p *measurementUnitRepo) GetByID(entity *common.RequestID) (*catalog_service.MeasurementUnit, error) {

	var (
		measurementUnit = catalog_service.MeasurementUnit{
			Precision: &catalog_service.Precision{},
		}
		shortUser            models.NullShortUser
		nameTranslation      []byte
		shortNameTranslation []byte
	)

	query := `
		SELECT 
			mu.id,
			dmu.id,
			dmu.long_name,
			dmu.short_name,
			mp.id,
			mp.value,
			mu.is_deletable,
			dmu.long_name_translation,
			dmu.short_name_translation,
			mu.created_at,
			u.id,
			u.first_name,
			u.last_name
		FROM 
			"measurement_unit" mu
		LEFT JOIN "user" u ON u.id = mu.created_by AND u.deleted_at = 0
		LEFT JOIN "default_measurement_unit" dmu ON mu.unit_id = dmu.id AND dmu.deleted_at = 0
		LEFT JOIN "measurement_precision" mp ON mp.id = mu.precision_id 
		WHERE
			mu.id = $1 AND mu.deleted_at = 0 AND mu.company_id=$2
	`

	err := p.db.QueryRow(query, entity.Id, entity.Request.CompanyId).Scan(
		&measurementUnit.Id,
		&measurementUnit.UnitId,
		&measurementUnit.LongName,
		&measurementUnit.ShortName,
		&measurementUnit.Precision.Id,
		&measurementUnit.Precision.Value,
		&measurementUnit.IsDeletable,
		&nameTranslation,
		&shortNameTranslation,
		&measurementUnit.CreatedAt,
		&shortUser.ID,
		&shortUser.FirstName,
		&shortUser.LastName,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error while getting measurement_unit")
	}

	if err = json.Unmarshal(nameTranslation, &measurementUnit.LongNameTranslation); err != nil {
		return nil, errors.Wrap(err, "error while Unmarshal NameTrasnlation")
	}

	if err = json.Unmarshal(nameTranslation, &measurementUnit.ShortNameTranslation); err != nil {
		return nil, errors.Wrap(err, "error while Unmarshal ShortNameTranslation")
	}

	if shortUser.ID.Valid {
		measurementUnit.CreatedBy = &common.ShortUser{
			Id:        shortUser.ID.String,
			FirstName: shortUser.FirstName.String,
			LastName:  shortUser.LastName.String,
			Image:     shortUser.Image.String,
		}
	}

	return &measurementUnit, nil
}

func (p *measurementUnitRepo) Update(entity *catalog_service.UpdateMeasurementUnitRequest) (string, error) {

	query := `
		UPDATE
			"measurement_unit"
		SET

			precision_id = $2,
			unit_id = $3
		WHERE
			id = $1 AND deleted_at=0 AND company_id=$4
	`

	_, err := p.db.Exec(
		query,
		entity.Id,
		entity.PrecisionId,
		entity.UnitId,
		entity.Request.CompanyId,
	)
	if err != nil {
		return "", errors.Wrap(err, "error while update measurement_unit")
	}

	return entity.Id, nil
}

func (p *measurementUnitRepo) Delete(req *common.RequestID) (*common.ResponseID, error) {

	query := `
	  	UPDATE
			"measurement_unit"
	  	SET
			deleted_at = extract(epoch from now())::bigint
	  	WHERE
			id = $1 AND deleted_at = 0 AND company_id = $2
	`

	res, err := p.db.Exec(
		query,
		req.Id,
		req.Request.CompanyId,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error while delete measurement_unit")
	}

	i, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if i == 0 {
		return nil, errors.New("measurement_unit not found")
	}

	return &common.ResponseID{Id: req.Id}, nil
}

func (p *measurementUnitRepo) GetAll(req *catalog_service.GetAllMeasurementUnitsRequest) (*catalog_service.GetAllMeasurementUnitsResponse, error) {

	var (
		res = catalog_service.GetAllMeasurementUnitsResponse{
			Data: make([]*catalog_service.ShortMeasurementUnit, 0),
		}

		searchFields = map[string]interface{}{
			"company_id": req.Request.CompanyId,
			"limit":      req.Limit,
			"offset":     req.Limit * (req.Page - 1),
			"search":     req.Search,
		}
	)

	namedQuery := `
		SELECT
			mu.id,
			dmu.long_name,
			dmu.short_name,
			mp.id,
			mp.value,
			mu.is_deletable,
			mu.created_at,
			u.id,
			u.first_name,
			u.last_name,
			u.image,
			dmu.long_name_translation,
			dmu.short_name_translation
		FROM
			"measurement_unit" mu
		LEFT JOIN "user" u ON u.id = mu.created_by AND u.deleted_at = 0
		LEFT JOIN "default_measurement_unit" dmu ON mu.unit_id = dmu.id AND dmu.deleted_at = 0
		LEFT JOIN "measurement_precision" mp ON mp.id=mu.precision_id 
	`

	filter := `
		WHERE
			mu.company_id = :company_id 
			AND
			mu.deleted_at = 0
	`

	if req.Search != "" {
		filter += `
		AND
		(
			dmu.long_name ILIKE '%' || :search || '%'
			OR
			dmu.short_name ILIKE '%' || :search || '%'
		)
	`
	}

	namedQuery += filter + `
		LIMIT :limit
		OFFSET :offset
	`

	rows, err := p.db.NamedQuery(namedQuery, searchFields)
	if err != nil {
		return nil, errors.Wrap(err, "error while select measurement_unit")
	}

	defer rows.Close()

	for rows.Next() {

		var (
			measurementUnit = catalog_service.ShortMeasurementUnit{
				Precision: &catalog_service.Precision{},
			}
			shortUser       models.NullShortUser
			shortTr, longTr []byte
		)

		err = rows.Scan(
			&measurementUnit.Id,
			&measurementUnit.LongName,
			&measurementUnit.ShortName,
			&measurementUnit.Precision.Id,
			&measurementUnit.Precision.Value,
			&measurementUnit.IsDeletable,
			&measurementUnit.CreatedAt,
			&shortUser.ID,
			&shortUser.FirstName,
			&shortUser.LastName,
			&shortUser.Image,
			&longTr,
			&shortTr,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error while getting measurement_unit pagin rows.Scan")
		}

		if err = json.Unmarshal(longTr, &measurementUnit.LongNameTranslation); err != nil {
			return nil, errors.Wrap(err, "error while Unmarshal NameTrasnlation")
		}

		if err = json.Unmarshal(shortTr, &measurementUnit.ShortNameTranslation); err != nil {
			return nil, errors.Wrap(err, "error while Unmarshal ShortNameTranslation")
		}

		if shortUser.ID.Valid {
			measurementUnit.CreatedBy = &common.ShortUser{
				Id:        shortUser.ID.String,
				FirstName: shortUser.FirstName.String,
				LastName:  shortUser.LastName.String,
				Image:     shortUser.Image.String,
			}
		}

		res.Data = append(res.Data, &measurementUnit)
	}

	countQuery := `
		SELECT 
			count(*) AS total
		FROM 
			"measurement_unit" mu
		LEFT JOIN "user" u ON u.id = mu.created_by AND u.deleted_at = 0
		LEFT JOIN "default_measurement_unit" dmu ON mu.unit_id = dmu.id AND dmu.deleted_at = 0
		LEFT JOIN "measurement_precision" mp ON mp.id=mu.precision_id
	` +
		filter

	queryStmt, err := p.db.PrepareNamed(countQuery)
	if err != nil {
		return nil, errors.Wrap(err, "error while PrepareNamed")
	}

	defer queryStmt.Close()

	if err != queryStmt.QueryRow(searchFields).Scan(&res.Total) {
		return nil, errors.Wrap(err, "error while scanning total")
	}

	return &res, nil
}

func (m *measurementUnitRepo) GetAllDefaultUnits(ctx context.Context, req *common.SearchRequest) (*catalog_service.GetAllDefaultUnitsResponse, error) {

	var (
		res = catalog_service.GetAllDefaultUnitsResponse{
			Precisions: make([]*catalog_service.Precision, 0),
			Units:      make([]*catalog_service.Unit, 0),
		}
	)

	query := `
	SELECT
		id,
		long_name,
		short_name,
		long_name_translation,
		short_name_translation
	FROM
		"default_measurement_unit"
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			unit                    catalog_service.Unit
			shortNameTr, longNameTr []byte
		)

		err := rows.Scan(&unit.Id, &unit.LongName, &unit.ShortName, &longNameTr, &shortNameTr)
		if err != nil {
			return nil, err
		}

		if len(shortNameTr) > 0 {
			err := json.Unmarshal(shortNameTr, &unit.ShortNameTranslation)
			if err != nil {
				return nil, err
			}
		}

		if len(longNameTr) > 0 {
			err := json.Unmarshal(longNameTr, &unit.LongNameTranslation)
			if err != nil {
				return nil, err
			}
		}

		res.Units = append(res.Units, &unit)
	}

	query = `
	SELECT
		id,
		value
	FROM
		"measurement_precision"
	`

	rows, err = m.db.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "error while getting measurement_precisions query")
	}

	defer rows.Close()

	for rows.Next() {
		var precision catalog_service.Precision

		err := rows.Scan(&precision.Id, &precision.Value)
		if err != nil {
			return nil, err
		}

		res.Precisions = append(res.Precisions, &precision)
	}

	return &res, err
}

package postgres

import (
	"genproto/catalog_service"
	"genproto/common"
	"strings"

	"github.com/Invan2/invan_catalog_service/config"
	"github.com/Invan2/invan_catalog_service/models"
	"github.com/Invan2/invan_catalog_service/pkg/helper"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/Invan2/invan_catalog_service/storage/repo"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type scalesTemplateRepo struct {
	db  models.DB
	log logger.Logger
	cfg config.Config
}

func NewScalesTemplateRepo(log logger.Logger, db models.DB, cfg config.Config) repo.ScalesTemplateI {
	return &scalesTemplateRepo{
		db:  db,
		log: log,
		cfg: cfg,
	}
}
func (st *scalesTemplateRepo) CreateScalesTemplates(req *catalog_service.CreateScalesTemplateRequest) (*common.ResponseID, error) {
	id := uuid.New().String()
	// insert scales_template
	query := `
		INSERT INTO
			"scales_template"
			(
				id,
				name,
				value,
				product_unit_ids,
				company_id,
				created_by
			)
		VALUES ( $1, $2, $3, $4, $5, $6 );
		`

	_, err := st.db.Exec(
		query,
		id,
		req.GetName(),
		req.GetValues(),
		req.GetMeasurementUnitIds(),
		req.GetRequest().GetCompanyId(),
		req.GetRequest().GetUserId(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error while insert scales_template")
	}

	return &common.ResponseID{Id: id}, nil
}

func (st *scalesTemplateRepo) GetScalesTemplateByID(req *catalog_service.GetScalesTemplateByIDRequest) (*catalog_service.ScalesTemplate, error) {
	var (
		scales             = &catalog_service.ScalesTemplate{}
		measurementUnitIds string
	)

	query := `
		SELECT id, name, product_unit_ids, value 
		FROM scales_template 
		WHERE company_id = $1 AND id = $2 AND deleted_at = 0
	`

	err := st.db.QueryRow(query, req.GetRequest().GetCompanyId(), req.GetId()).Scan(&scales.Id, &scales.Name, &measurementUnitIds, &scales.Values)
	if err != nil {
		return nil, errors.Wrap(err, "error while GetAllScalesTemplates. Scan")
	}
	scales.MeasurementUnitId = strings.Split(measurementUnitIds, ",")

	return scales, nil
}

func (st *scalesTemplateRepo) GetAllScalesTemplates(req *catalog_service.GetAllScalesTemplatesRequest) (*catalog_service.GetAllScalesTemplatesResponse, error) {
	var (
		count  int
		filter string
		args   = make(map[string]interface{})
		res    = &catalog_service.GetAllScalesTemplatesResponse{}
	)

	//Get total scales templates
	countQuery := `SELECT count(1) FROM scales_template WHERE company_id = $1 AND deleted_at = 0`
	err := st.db.QueryRow(countQuery, req.GetRequest().GetCompanyId()).Scan(&count)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, name, product_unit_ids 
		FROM scales_template 
		WHERE company_id = :company_id AND deleted_at = 0
	`

	// Get all scales
	args["company_id"] = req.GetRequest().GetCompanyId()
	args["limit"] = req.GetLimit()
	args["offset"] = req.GetLimit() * (req.GetPage() - 1)
	filter += ` LIMIT :limit OFFSET :offset`

	query += ` ORDER BY created_at desc ` + filter
	query, arrArgs := helper.ReplaceQueryParams(query, args)

	rows, err := st.db.Query(query, arrArgs...)
	if err != nil {
		return nil, errors.Wrap(err, "error while GetAllScalesTemplates. GetRows")
	}
	defer rows.Close()

	for rows.Next() {
		scales := catalog_service.ScalesTemplate{}
		var measurementUnitIds string
		err = rows.Scan(
			&scales.Id,
			&scales.Name,
			&measurementUnitIds,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error while GetAllScalesTemplates. Scan")
		}
		scales.MeasurementUnitId = strings.Split(measurementUnitIds, ",")
		res.ScalesTemplates = append(res.ScalesTemplates, &scales)
	}
	res.Total = int32(count)

	return res, nil
}

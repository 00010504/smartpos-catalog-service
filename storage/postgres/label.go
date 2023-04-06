package postgres

import (
	"genproto/catalog_service"
	"genproto/common"
	"strings"

	"github.com/Invan2/invan_catalog_service/models"
	"github.com/Invan2/invan_catalog_service/pkg/helper"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/Invan2/invan_catalog_service/storage/repo"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type labelRepo struct {
	db  models.DB
	log logger.Logger
}

func NewLabelRepo(log logger.Logger, db models.DB) repo.LabelI {
	return &labelRepo{
		db:  db,
		log: log,
	}
}

func (l *labelRepo) createLabelContent(labelId string, contents map[string]*catalog_service.LabelContent, createdBy string) error {

	var values = []interface{}{}

	query := `
		INSERT INTO
			"label_content"
		(
			id,
			label_id,
			position_x,
			position_y,
			width,
			height,
			type,
			product_image,
			field_name,
			font_family,
			font_style,
			font_size,
			font_weight,
			text_align,
			created_by
		)
		VALUES 
	`

	for _, content := range contents {
		query += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"

		values = append(
			values,
			uuid.NewString(),
			labelId,
			content.Position.X,
			content.Position.Y,
			content.Width,
			content.Height,
			content.Type,
			content.ProductImage,
			content.FieldName,
			content.Format.FontFamily,
			content.Format.FontStyle,
			content.Format.FontSize,
			content.Format.FontWeight,
			content.Format.TextAlign,
			createdBy,
		)
	}

	query = strings.TrimSuffix(query, ",")
	query = helper.ReplaceSQL(query, "?")

	stmt, err := l.db.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "error while insert label_content. Prepare")
	}

	defer stmt.Close()

	_, err = stmt.Exec(values...)
	if err != nil {
		return errors.Wrap(err, "error while insert label_content. Exec")
	}

	return nil
}

func (l *labelRepo) Create(entity *catalog_service.CreateLabelRequest) (*common.ResponseID, error) {

	var (
		id = uuid.NewString()
	)

	query := `
		INSERT INTO
			"label"
		(
			id,
			name,
			width,
			height,
			company_id,
			created_by
		)
		VALUES ($1, $2, $3, $4, $5, $6);
	`

	_, err := l.db.Exec(
		query,
		id,
		entity.Parameters.Name,
		entity.Parameters.Width,
		entity.Parameters.Height,
		entity.Request.CompanyId,
		entity.Request.UserId,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error while insert label")
	}

	err = l.createLabelContent(id, entity.Content, entity.Request.UserId)
	if err != nil {
		return nil, err
	}

	return &common.ResponseID{Id: id}, nil
}

func (l *labelRepo) GetById(req *common.RequestID) (*catalog_service.GetLabelResponse, error) {

	var res = catalog_service.GetLabelResponse{
		Parameters: &catalog_service.LabelParametrs{},
		Content:    make(map[string]*catalog_service.LabelContent),
	}

	query := `
		SELECT
			id,
			name,
			width,
			height
		FROM "label"
		WHERE id = $1 AND company_id = $2 AND deleted_at = 0
	`

	err := l.db.QueryRow(query, req.Id, req.Request.CompanyId).Scan(
		&res.Id,
		&res.Parameters.Name,
		&res.Parameters.Width,
		&res.Parameters.Height,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error while get label by id. Scan")
	}

	content, err := l.getContent([]string{res.Id})
	if err != nil {
		return nil, err
	}

	res.Content = content[res.Id]

	return &res, nil
}

func (l *labelRepo) UpdateById(req *catalog_service.UpdateLabelRequest) (*common.ResponseID, error) {

	query := `
		UPDATE "label"
		SET
			name = $2,
			width = $3,
			height = $4
		WHERE id = $1 AND deleted_at = 0
	`

	res, err := l.db.Exec(query, req.Id, req.Parameters.Name, req.Parameters.Width, req.Parameters.Height)
	if err != nil {
		return nil, errors.Wrap(err, "error while update labelById. Exec")
	}

	i, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if i == 0 {
		return nil, errors.Wrap(errors.New("label not found"), "i == 0")
	}

	// delete old contents
	query = `
		UPDATE "label_content"
		SET
			deleted_by = $2,
			deleted_at = extract(epoch from now())::bigint
		WHERE label_id = $1 AND deleted_at = 0
	`

	_, err = l.db.Exec(query, req.Id, req.Request.UserId)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	// insert content
	err = l.createLabelContent(req.Id, req.Content, req.Request.UserId)
	if err != nil {
		return nil, err
	}

	return &common.ResponseID{Id: req.Id}, nil
}

func (l *labelRepo) GetAll(req *common.SearchRequest) (*catalog_service.GetAllLabelsResponse, error) {

	var (
		res = catalog_service.GetAllLabelsResponse{
			Data: []*catalog_service.GetLabelResponse{},
		}
		searchFields = map[string]interface{}{
			"company_id": req.Request.CompanyId,
			"limit":      req.Limit,
			"offset":     req.Limit * (req.Page - 1),
			"search":     req.Search,
		}
		ids []string
	)

	namedQuery := `
		SELECT
			id,
			name,
			width,
			height
		FROM "label" l
	`

	filter := `
		WHERE l.company_id = :company_id AND l.deleted_at = 0
	`

	if req.Search != "" {
		filter += `
			AND l.name ILIKE '%' || :search || '%'
		`
	}

	namedQuery += filter + `
		LIMIT :limit
		OFFSET :offset
	`

	rows, err := l.db.NamedQuery(namedQuery, searchFields)
	if err != nil {
		return nil, errors.Wrap(err, "error while get all labels. Scan")
	}

	for rows.Next() {

		var label = catalog_service.GetLabelResponse{Parameters: &catalog_service.LabelParametrs{}}

		err := rows.Scan(
			&label.Id,
			&label.Parameters.Name,
			&label.Parameters.Width,
			&label.Parameters.Height,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error while get labels. rows.Scan")
		}

		ids = append(ids, label.Id)
		res.Data = append(res.Data, &label)
	}

	countQuery := `
		SELECT count(*) AS total
		FROM "label" l
	` + filter

	resStmt, err := l.db.PrepareNamed(countQuery)
	if err != nil {
		return nil, errors.Wrap(err, "error while scanning total PrepareNamed")
	}

	defer resStmt.Close()

	if err != resStmt.QueryRow(searchFields).Scan(&res.Total) {
		return nil, errors.Wrap(err, "error while scanning total")
	}

	content, err := l.getContent(ids)
	if err != nil {
		return nil, err
	}

	for _, label := range res.Data {
		label.Content = content[label.Id]
	}

	return &res, nil
}

func (l *labelRepo) getContent(ids []string) (map[string]map[string]*catalog_service.LabelContent, error) {

	var res = make(map[string]map[string]*catalog_service.LabelContent)

	query := `
		SELECT
			id,
			label_id,
			position_x,
			position_y,
			width,
			height,
			type,
			product_image,
			field_name,
			font_family,
			font_style,
			font_size,
			font_weight,
			text_align
		FROM "label_content"
		WHERE label_id = ANY($1) AND deleted_at = 0
	`

	rows, err := l.db.Query(query, pq.Array(ids))
	if err != nil {
		return nil, errors.Wrap(err, "error while get label content. Scan")
	}

	defer rows.Close()

	for rows.Next() {

		var (
			content = catalog_service.LabelContent{
				Position: &catalog_service.LabelPosition{},
				Format:   &catalog_service.TextFormat{},
			}
			labelId string
		)

		err = rows.Scan(
			&content.Id,
			&labelId,
			&content.Position.X,
			&content.Position.Y,
			&content.Width,
			&content.Height,
			&content.Type,
			&content.ProductImage,
			&content.FieldName,
			&content.Format.FontFamily,
			&content.Format.FontStyle,
			&content.Format.FontSize,
			&content.Format.FontWeight,
			&content.Format.TextAlign,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error while getting label_content. rows.Scan")
		}

		_, ok := res[labelId]
		if ok {
			res[labelId][content.FieldName] = &content
		} else {
			res[labelId] = map[string]*catalog_service.LabelContent{content.FieldName: &content}
		}
	}

	return res, nil
}

func (l *labelRepo) DeleteLabelById(req *common.RequestID) (*common.ResponseID, error) {

	query := `
	  	UPDATE
			"label"
	  	SET
			deleted_at = extract(epoch from now())::bigint
	  	WHERE
			id = $1 AND deleted_at = 0 AND company_id = $2
	`

	res, err := l.db.Exec(
		query,
		req.Id,
		req.Request.CompanyId,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error while delete label")
	}

	i, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if i == 0 {
		return nil, errors.Wrap(errors.New("label not found"), "error while delete label rowsAffected = 0")
	}

	return &common.ResponseID{Id: req.Id}, nil
}

func (l *labelRepo) DeleteLabelsByIds(req *common.RequestIDs) (*common.Empty, error) {

	query := `
	  	UPDATE
			"label"
	  	SET
			deleted_at = extract(epoch from now())::bigint
	  	WHERE
			deleted_at = 0 AND id = ANY($1) AND company_id = $2
	`

	res, err := l.db.Exec(query, pq.Array(req.Ids), req.Request.CompanyId)
	if err != nil {
		return nil, errors.Wrap(err, "error while delete labels")
	}

	i, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if int(i) != len(req.Ids) {
		return nil, errors.Wrap(errors.New("label not found"), "error while delete labels rowsAffected = 0")
	}

	return &common.Empty{}, nil
}

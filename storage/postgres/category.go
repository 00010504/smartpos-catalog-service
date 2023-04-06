package postgres

import (
	"database/sql"
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

type categoryRepo struct {
	db  models.DB
	log logger.Logger
}

func NewCategoryRepo(log logger.Logger, db models.DB) repo.CategoryPgI {
	return &categoryRepo{
		db:  db,
		log: log,
	}
}

func (c *categoryRepo) Create(entity *catalog_service.CreateCategoryRequest) (string, error) {

	var values = []interface{}{}
	id := uuid.New().String()

	query := `
		INSERT INTO
		"category"
		(
			id,
			name,
			parent_id,
			company_id,
			created_by
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5
		);
	`

	_, err := c.db.Exec(
		query,
		id,
		entity.Name,
		helper.NullString(entity.ParentId),
		entity.Request.CompanyId,
		entity.Request.UserId,
	)
	if err != nil {
		return "", errors.Wrap(err, "error while insert category")
	}

	// create child categories
	if len(entity.Child) > 0 {

		query = `
		INSERT INTO
			"category"
		(
			id,
			name,
			parent_id,
			company_id,
			created_by
		)
		VALUES
	`
		for _, childName := range entity.Child {
			query += "(?, ?, ?, ?, ?),"
			values = append(values,
				uuid.New().String(),
				childName,
				id,
				entity.Request.CompanyId,
				entity.Request.UserId,
			)
		}

		query = strings.TrimSuffix(query, ",")
		query = helper.ReplaceSQL(query, "?")

		stmt, err := c.db.Prepare(query)
		if err != nil {
			return "", errors.Wrap(err, "category.create. error while insert category child. Prepare")
		}
		_, err = stmt.Exec(values...)
		if err != nil {
			return "", errors.Wrap(err, "category.create. error while insert category child. Exec")
		}

		stmt.Close()
	}

	return id, nil
}

func (c *categoryRepo) GetByID(entity *common.RequestID) (*catalog_service.GetCategoryByIDResponse, error) {

	var (
		category  catalog_service.Category
		shortUser models.NullShortUser
		parentId  sql.NullString
	)

	query := `
		SELECT 
			cat.id,
			cat.name,
			CAST (cat.parent_id AS VARCHAR(64)),
			u.id,
			u.first_name,
			u.last_name,
			u.image
		FROM 
			"category" cat
		LEFT JOIN
			"user" u
		ON
			u.id = cat.created_by AND u.deleted_at = 0
		WHERE
			cat.id = $1 AND
			cat.deleted_at = 0 AND
			cat.company_id = $2
	`

	err := c.db.QueryRow(query, entity.Id, entity.Request.CompanyId).Scan(
		&category.Id,
		&category.Name,
		&parentId,
		&shortUser.ID,
		&shortUser.FirstName,
		&shortUser.LastName,
		&shortUser.Image,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error while getting category")
	}

	category.ParentId = parentId.String

	if shortUser.ID.Valid {
		category.ShortUser = &common.ShortUser{
			Id:        shortUser.ID.String,
			FirstName: shortUser.FirstName.String,
			LastName:  shortUser.LastName.String,
			Image:     shortUser.Image.String,
		}
	}

	childs, err := getCategoriesChilds(c.db, []string{category.Id})
	if err != nil {
		return nil, errors.Wrap(err, "error while getting category childs")
	}

	category.Children = childs[category.Id]

	return &catalog_service.GetCategoryByIDResponse{
		Category: &category,
	}, nil
}

func (c *categoryRepo) Update(entity *catalog_service.UpdateCategoryRequest) (string, error) {

	var (
		parentId sql.NullString
		values   = []interface{}{}
	)

	if entity.ParentId != "" {
		parentId.String = entity.ParentId
		parentId.Valid = true
	}

	query := `
	  	UPDATE "category"
	  	SET
			name = $2,
			parent_id = $3
	  	WHERE id = $1 AND deleted_at = 0
	`

	_, err := c.db.Exec(query, entity.Id, entity.Name, parentId)
	if err != nil {
		return "", errors.Wrap(err, "error while update category")
	}

	// delete category childs
	query = `
		UPDATE "category"
		SET deleted_at = extract(epoch from now())::bigint
		WHERE parent_id = $1 AND company_id = $2 AND deleted_at = 0
	`

	res, err := c.db.Exec(query, entity.Id, entity.Request.CompanyId)
	if err != nil {
		return "", errors.Wrap(err, "error while delete category childs")
	}

	_, err = res.RowsAffected()
	if err != nil {
		return "", errors.Wrap(err, "error while delete category childs. RowsAffected")
	}

	// insert childs
	if len(entity.Children) > 0 {

		values = []interface{}{}

		query = `
			INSERT INTO
				"category"
			(
				id,
				name,
				parent_id,
				company_id,
				created_by
			)
			VALUES
		`
		for _, childName := range entity.Children {
			query += "(?, ?, ?, ?, ?),"
			values = append(values,
				uuid.New().String(),
				childName.Name,
				entity.Id,
				entity.Request.CompanyId,
				entity.Request.UserId,
			)
		}

		query = strings.TrimSuffix(query, ",")
		query = helper.ReplaceSQL(query, "?")

		stmt, err := c.db.Prepare(query)
		if err != nil {
			return "", errors.Wrap(err, "category.update. error while insert category child. Prepare")
		}
		_, err = stmt.Exec(values...)
		if err != nil {
			return "", errors.Wrap(err, "category.update. error while insert category child. Exec")
		}

		stmt.Close()
	}

	return entity.Id, nil
}

func (c *categoryRepo) Delete(req *common.RequestID) (*common.ResponseID, error) {

	query := `
		UPDATE
			"category"
		SET
			deleted_at = extract(epoch from now())::bigint
		WHERE
			(id = $1 OR parent_id = $1) AND deleted_at = 0 AND company_id = $2
	`

	res, err := c.db.Exec(
		query,
		req.Id,
		req.Request.CompanyId,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error while delete category")
	}

	i, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if i == 0 {
		return nil, errors.New("category not found")
	}

	return &common.ResponseID{Id: req.Id}, nil
}

func (c *categoryRepo) GetAll(req *catalog_service.GetAllCategoriesRequest) (*catalog_service.GetAllCategoriesResponse, error) {

	var (
		res          catalog_service.GetAllCategoriesResponse
		searchFields = map[string]interface{}{
			"company_id": req.Request.CompanyId,
			"limit":      req.Limit,
			"offset":     req.Limit * (req.Page - 1),
			"name":       req.Search,
		}
		categoryIds = make([]string, 0)
	)

	namedQuery := `
		SELECT
			cat.id,
			cat.name,
			cat.created_at,
			u.id,
			u.first_name,
			u.last_name,
			u.image
		FROM 
			"category" cat
		LEFT JOIN
			"user" u
		ON
			u.id = cat.created_by AND u.deleted_at = 0
	`

	filter := `
		WHERE
			cat.deleted_at = 0 AND
			cat.company_id = :company_id AND
			cat.parent_id IS NULL
	`

	if req.Search != "" {
		filter += `
		AND cat.name ILIKE '%' || :name || '%'
	`
	}

	namedQuery += filter + `
		LIMIT :limit
		OFFSET :offset
	`

	rows, err := c.db.NamedQuery(namedQuery, searchFields)
	if err != nil {
		return nil, errors.Wrap(err, "error while getting category pagin")
	}

	defer rows.Close()

	res.Data = make([]*catalog_service.Category, 0)

	for rows.Next() {
		var (
			shortUser models.NullShortUser
			category  catalog_service.Category
		)
		err = rows.Scan(
			&category.Id,
			&category.Name,
			&category.CreatedAt,
			&shortUser.ID,
			&shortUser.FirstName,
			&shortUser.LastName,
			&shortUser.Image,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error while getting category pagin rows.Scan")
		}

		categoryIds = append(categoryIds, category.Id)

		if shortUser.ID.Valid {
			category.ShortUser = &common.ShortUser{
				Id:        shortUser.ID.String,
				FirstName: shortUser.FirstName.String,
				LastName:  shortUser.LastName.String,
				Image:     shortUser.Image.String,
			}
		}

		res.Data = append(res.Data, &category)
	}

	childs, err := getCategoriesChilds(c.db, categoryIds)
	if err != nil {
		return nil, errors.Wrap(err, "error while getting category childs")
	}

	for index, category := range res.Data {
		res.Data[index].Children = childs[category.Id]
	}

	countQuery := `
		SELECT 
			count(*) AS total
		FROM 
			"category" cat
	` +
		filter

	resStmt, err := c.db.PrepareNamed(countQuery)
	if err != nil {
		return nil, errors.Wrap(err, "error while scanning total PrepareNamed")
	}

	defer resStmt.Close()

	if err != resStmt.QueryRow(searchFields).Scan(&res.Total) {
		return nil, errors.Wrap(err, "error while scanning total")
	}

	return &res, nil
}

func getCategoriesChilds(db models.DB, ids []string) (map[string][]*catalog_service.Category, error) {

	var (
		res = make(map[string][]*catalog_service.Category)
	)

	if len(ids) == 0 {
		return res, nil
	}

	query := `
		SELECT 
			cat.id,
			cat.name,
			cat.parent_id,
			cat.created_at,
			u.id,
			u.first_name,
			u.last_name,
			u.image
		FROM 
			"category" cat
		LEFT JOIN
			"user" u
		ON
			u.id = cat.created_by AND u.deleted_at = 0
		WHERE
			cat.deleted_at = 0 AND
			cat.parent_id = ANY ($1)
	`

	rows, err := db.Query(query, pq.Array(ids))
	if err != nil {
		return nil, errors.Wrap(err, "error while getting categories childs")
	}

	defer rows.Close()

	for rows.Next() {

		var (
			shortUser models.NullShortUser
			category  catalog_service.Category
		)

		err = rows.Scan(
			&category.Id,
			&category.Name,
			&category.ParentId,
			&category.CreatedAt,
			&shortUser.ID,
			&shortUser.FirstName,
			&shortUser.LastName,
			&shortUser.Image,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error while getting categories childs rows.Scan")
		}

		if shortUser.ID.Valid {
			category.ShortUser = &common.ShortUser{
				Id:        shortUser.ID.String,
				FirstName: shortUser.FirstName.String,
				LastName:  shortUser.LastName.String,
				Image:     shortUser.Image.String,
			}
		}

		res[category.ParentId] = append(res[category.ParentId], &category)
	}

	return res, nil
}

func (c *categoryRepo) GetShortCategoriesByIds(ids []string) ([]*catalog_service.ShortCategory, error) {

	var (
		res []*catalog_service.ShortCategory
	)

	if len(ids) == 0 {
		return res, nil
	}

	namedQuery := `
		SELECT
			cat.parent_id,
			cat.id,
			cat.name
		FROM 
			"category" cat
		WHERE
			cat.deleted_at = 0 AND
			cat.id = ANY ($1)
	`

	rows, err := c.db.Query(namedQuery, pq.Array(ids))
	if err != nil {
		return nil, errors.Wrap(err, "error while getting category pagin")
	}

	defer rows.Close()

	for rows.Next() {
		var (
			category catalog_service.ShortCategory
			parentId sql.NullString
		)
		err = rows.Scan(
			&parentId,
			&category.Id,
			&category.Name,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error while getting category pagin rows.Scan")
		}

		category.ParentId = parentId.String

		res = append(res, &category)
	}

	return res, nil
}

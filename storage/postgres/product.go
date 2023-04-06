package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"genproto/catalog_service"
	"genproto/common"
	"strconv"
	"strings"

	"github.com/Invan2/invan_catalog_service/config"
	"github.com/Invan2/invan_catalog_service/models"
	"github.com/Invan2/invan_catalog_service/pkg/helper"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/Invan2/invan_catalog_service/storage/repo"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type productRepo struct {
	db  models.DB
	log logger.Logger
	cfg config.Config
}

func NewProductRepo(log logger.Logger, db models.DB, cfg config.Config) repo.ProductPgI {
	return &productRepo{
		db:  db,
		log: log,
		cfg: cfg,
	}
}

func (p *productRepo) Create(product *catalog_service.CreateProductRequest) (productId string, productDetailId string, err error) {

	// p.log.info("create product", logger.Any("product", product))

	productId = uuid.New().String()

	// insert product
	query := `
		INSERT INTO
			"product"
		(
			id,
			company_id,
			product_type_id,
			created_by
		)
		VALUES (
			$1,
			$2,
			$3,
			$4
		);
	`

	_, err = p.db.Exec(
		query,
		productId,
		product.Request.CompanyId,
		product.ProductTypeId,
		product.Request.UserId,
	)
	if err != nil {
		return "", "", errors.Wrap(err, "error while insert product")
	}

	// insert product_detail
	productDetailId, err = p.createProductDetail(product, productId)
	if err != nil {
		return "", "", err
	}

	return productId, productDetailId, nil
}

func (p *productRepo) createProductDetail(product *catalog_service.CreateProductRequest, productId string) (string, error) {

	var (
		values          = []interface{}{}
		productDetailId = uuid.New().String()
	)

	// insert product_detail
	query := `
		INSERT INTO
			"product_detail"
		(
			version,
			id,
			product_id,
			sku,
			name,
			mxik_code,
			is_marking,
			brand_id,
			description,
			measurement_unit_id,
			created_by,
			supplier_id,
			vat_id
		)
		VALUES (
			(
				SELECT last_version
				FROM "product"
				WHERE deleted_at = 0  AND id = $2
			),
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12
		);
	`

	_, err := p.db.Exec(
		query,
		productDetailId,
		productId,
		product.Sku,
		product.Name,
		product.MxikCode,
		product.IsMarking,
		helper.NullString(product.BrandId),
		product.Description,
		product.MeasurementUnitId,
		product.Request.UserId,
		product.SupplierId,
		product.VatId,
	)
	if err != nil {
		return "", errors.Wrap(err, "error while insert product_detail")
	}

	values = []interface{}{}
	// // insert barcode
	if len(product.Barcodes) > 0 {

		query = `
			INSERT INTO
				"product_barcode"
			(
				barcode,
				product_detail_id
			)
			VALUES
		`
		for _, barcode := range product.Barcodes {
			query += "(?, ?),"
			values = append(values,
				barcode,
				productDetailId,
			)
		}

		query = strings.TrimSuffix(query, ",")
		query = helper.ReplaceSQL(query, "?")

		stmt, err := p.db.Prepare(query)
		if err != nil {
			return "", errors.Wrap(err, "error while insert product_barcode. Prepare")
		}

		defer stmt.Close()

		_, err = stmt.Exec(values...)
		if err != nil {
			return "", errors.Wrap(err, "error while insert product_barcode. Exec")
		}
	}

	values = []interface{}{}
	// insert product tags
	if len(product.TagIds) > 0 {

		query = `
			INSERT INTO
				"product_tag"
			(
				tag_id,
				product_detail_id
			)
			VALUES 
		`
		for _, tag := range product.TagIds {
			query += "(?, ?),"
			values = append(values,
				tag,
				productDetailId,
			)
		}

		query = strings.TrimSuffix(query, ",")
		query = helper.ReplaceSQL(query, "?")

		stmt, err := p.db.Prepare(query)
		if err != nil {
			return "", errors.Wrap(err, "error while insert product_tag. Prepare")
		}

		defer stmt.Close()

		_, err = stmt.Exec(values...)
		if err != nil {
			return "", errors.Wrap(err, "error while insert product_tag. Exec")
		}

	}
	values = []interface{}{}

	// insert product categories
	if len(product.Images) > 0 {

		query = `
			INSERT INTO
				"product_image"
			(
				id,
				sequence_number,
				product_detail_id,
				file_name
			)
			VALUES 
		`
		for _, image := range product.Images {

			query += "(?, ?, ?, ?),"
			values = append(values,
				uuid.New().String(),
				image.SequenceNumber,
				productDetailId,
				image.ImageUrl,
			)
		}

		query = strings.TrimSuffix(query, ",")
		query = helper.ReplaceSQL(query, "?")

		stmt, err := p.db.Prepare(query)
		if err != nil {
			return "", errors.Wrap(err, "error while insert product_image. Prepare")
		}

		defer stmt.Close()

		_, err = stmt.Exec(values...)
		if err != nil {
			return "", errors.Wrap(err, "error while insert product_image. Exec")
		}
	}

	values = []interface{}{}
	// upsert product measurement_values
	if len(product.ShopMeasurementValues) > 0 {

		// upsert measurement values
		query = `
			INSERT INTO
				"measurement_values"
			(
				shop_id,
				product_id,
				is_available,
				has_trigger,
				amount,
				small_left
			)
			VALUES 
		`

		for _, value := range product.ShopMeasurementValues {
			query += "(?, ?, ?, ?, ?, ?),"
			values = append(values,
				value.ShopId,
				productId,
				value.IsAvailable,
				value.HasTrigger,
				value.Amount,
				value.SmallLeft,
			)
		}

		query = strings.TrimSuffix(query, ",")
		query = helper.ReplaceSQL(query, "?")

		query += `
			ON CONFLICT (product_id, shop_id) 
			DO UPDATE SET
				is_available = EXCLUDED.is_available,
				has_trigger = EXCLUDED.has_trigger,
				amount = EXCLUDED.amount,
				small_left = EXCLUDED.small_left
		`

		stmt, err := p.db.Prepare(query)
		if err != nil {
			return "", errors.Wrap(err, "error while insert product measurement_values. Prepare")
		}

		defer stmt.Close()

		_, err = stmt.Exec(values...)
		if err != nil {
			return "", errors.Wrap(err, "error while insert product measurement_values. Exec")
		}
	}

	// upsert shop_price
	values = []interface{}{}
	if len(product.ShopPrices) > 0 {
		query = `
			INSERT INTO
				"shop_price"
			(
				id,
				product_id,
				shop_id,
				min_price,
				max_price,
				supply_price,
				retail_price,
				whole_sale_price
			)
			VALUES 
		`

		for _, value := range product.ShopPrices {
			query += "(?, ?, ?, ?, ?, ?, ?, ?),"

			values = append(values,
				uuid.New().String(),
				productId,
				value.ShopId,
				value.MinPrice,
				value.MaxPrice,
				value.SupplyPrice,
				value.RetailPrice,
				value.WholeSalePrice,
			)
		}

		query = strings.TrimSuffix(query, ",")
		query = helper.ReplaceSQL(query, "?")

		query += `
			ON CONFLICT (product_id, shop_id) 
			DO UPDATE SET
				min_price = EXCLUDED.min_price,
				max_price = EXCLUDED.max_price,
				supply_price = EXCLUDED.supply_price,
				retail_price = EXCLUDED.retail_price,
				whole_sale_price = EXCLUDED.whole_sale_price
		`

		stmt, err := p.db.Prepare(query)
		if err != nil {
			return "", errors.Wrap(err, "error while insert product shop_price. Prepare")
		}

		defer stmt.Close()

		_, err = stmt.Exec(values...)
		if err != nil {
			return "", errors.Wrap(err, "error while insert product shop_price. Exec")
		}
	}

	values = []interface{}{}
	// insert product categories
	if len(product.CategoryIds) > 0 {

		query = `
				INSERT INTO
					"product_category"
				(
					product_detail_id,
					category_id
				)
				VALUES 
			`
		for _, category := range product.CategoryIds {

			query += "(?, ?),"

			values = append(values, productDetailId, category)
		}

		query = strings.TrimSuffix(query, ",")
		query = helper.ReplaceSQL(query, "?")

		stmt, err := p.db.Prepare(query)
		if err != nil {
			return "", errors.Wrap(err, "error while insert product_category. Prepare")
		}

		defer stmt.Close()

		_, err = stmt.Exec(values...)
		if err != nil {
			return "", errors.Wrap(err, "error while insert product_category. Exec")
		}
	}

	return productDetailId, nil
}

func (p *productRepo) GetByID(req *common.RequestID) (*catalog_service.Product, error) {

	var (
		product         catalog_service.Product
		shortUser       models.NullShortUser
		brand           models.ProductNullBrand
		supplier        models.ProductNullSupplier
		vat             models.VatNullSupplier
		measurementUnit = models.ProductNullMeasurementUnit{
			Precision: models.ProductNullPrecision{},
		}
		longNameTranslation  []byte
		shortNameTranslation []byte
		productDetailId      string
	)

	query := `
		SELECT 
			p.id,
			p.product_type_id,
			CAST (p.created_at AS VARCHAR(64)),
			pd.id,
			pd.name,
			pd.sku,
			pd.mxik_code,
			pd.is_marking,
			pd.description,
			br.id,
			br.name,
			s.id,
			s.name,
			v.id,
			v.name,
			v.percentage,
			mu.id,
			dmu.short_name,
			dmu.long_name,
			dmu.short_name_translation,
			dmu.long_name_translation,
			mp.id,
			mp.value,
			u.id,
			u.first_name,
			u.last_name
		FROM 
			"product" p
		JOIN  "product_detail" pd ON p.id = pd.product_id AND p.last_version = pd.version
		LEFT JOIN "brand" br ON br.id = pd.brand_id
		LEFT JOIN "supplier" s ON s.id = pd.supplier_id AND s.deleted_at = 0
		LEFT JOIN "vat" v ON v.id = pd.vat_id AND v.deleted_at = 0
		LEFT JOIN "measurement_unit" mu ON mu.id = pd.measurement_unit_id
		LEFT JOIN "default_measurement_unit" dmu ON mu.unit_id = dmu.id AND dmu.deleted_at = 0
		LEFT JOIN "measurement_precision" mp ON mp.id = mu.precision_id 
		LEFT JOIN "user" u ON u.id = p.created_by AND u.deleted_at = 0
		WHERE
			p.id = $1 AND p.deleted_at = 0 AND p.company_id = $2
	`

	err := p.db.QueryRow(query, req.Id, req.Request.CompanyId).Scan(
		&product.Id,
		&product.ProductTypeId,
		&product.CreatedAt,
		&productDetailId,
		&product.Name,
		&product.Sku,
		&product.MxikCode,
		&product.IsMarking,
		&product.Description,
		&brand.Id,
		&brand.Name,
		&supplier.Id,
		&supplier.Name,
		&vat.Id,
		&vat.Name,
		&vat.Percentage,
		&measurementUnit.Id,
		&measurementUnit.ShortName,
		&measurementUnit.LongName,
		&shortNameTranslation,
		&longNameTranslation,
		&measurementUnit.Precision.Id,
		&measurementUnit.Precision.Value,
		&shortUser.ID,
		&shortUser.FirstName,
		&shortUser.LastName,
	)

	if err != nil {
		return nil, errors.Wrap(err, "error while getting product")
	}

	if shortUser.ID.Valid {
		product.CreatedBy = &common.ShortUser{
			Id:        shortUser.ID.String,
			FirstName: shortUser.FirstName.String,
			LastName:  shortUser.LastName.String,
			Image:     shortUser.Image.String,
		}
	}

	if supplier.Id.Valid {
		product.Supplier = &catalog_service.ShortSupplier{
			Id:   supplier.Id.String,
			Name: supplier.Name.String,
		}
	}

	if vat.Id.Valid {
		product.Vat = &catalog_service.ShortVat{
			Id:   vat.Id.String,
			Name: vat.Name.String,
		}
	}

	if measurementUnit.Id.Valid {
		product.MeasurementUnit = &catalog_service.ShortMeasurementUnit{
			Id:        measurementUnit.Id.String,
			ShortName: measurementUnit.ShortName.String,
			LongName:  measurementUnit.LongName.String,
		}
		if measurementUnit.Precision.Id.Valid {
			product.MeasurementUnit.Precision = &catalog_service.Precision{
				Id:    measurementUnit.Precision.Id.String,
				Value: measurementUnit.Precision.Value.String,
			}
		}

		if err = json.Unmarshal(longNameTranslation, &measurementUnit.LongNameTranslation); err != nil {
			return nil, errors.Wrap(err, "error while Unmarshal NameTrasnlation")
		}

		if err = json.Unmarshal(shortNameTranslation, &measurementUnit.ShortNameTranslation); err != nil {
			return nil, errors.Wrap(err, "error while Unmarshal ShortNameTranslation")
		}
	}

	product.Barcodes, err = p.getProductBarcodes(productDetailId)
	if err != nil {
		return nil, err
	}

	product.Categories, err = p.getProductCategories(productDetailId)
	if err != nil {
		return nil, err
	}

	product.Images, err = p.getProductImages(productDetailId)
	if err != nil {
		return nil, err
	}

	product.MeasurementValues, err = p.getProductMeasurementValues(req.Id, productDetailId)
	if err != nil {
		return nil, err
	}

	shopPrices, err := p.getProductShopPrices([]string{req.Id})
	if err != nil {
		return nil, err
	}

	product.ShopPrices = shopPrices[req.Id]

	return &product, nil
}

func (p *productRepo) Update(entity *catalog_service.UpdateProductRequest) (*common.ResponseID, error) {

	query := `
	  	UPDATE
				"product"
	  	SET
				last_version = last_version + 1,
				parent_id = $2,
				product_type_id = $3
	  	WHERE id = $1
	`

	res, err := p.db.Exec(query, entity.Id, helper.NullString(entity.ParentId), entity.ProductTypeId)
	if err != nil {
		return nil, errors.Wrap(err, "error while update product")
	}

	if i, _ := res.RowsAffected(); i == 0 {
		err = sql.ErrNoRows
		return nil, errors.Wrap(err, "error while update product")
	}

	// insert product_detail
	_, err = p.createProductDetail(
		&catalog_service.CreateProductRequest{
			Request:               entity.Request,
			Barcodes:              entity.Barcodes,
			BrandId:               entity.BrandId,
			Name:                  entity.Name,
			ParentId:              entity.ParentId,
			Sku:                   entity.Sku,
			CategoryIds:           entity.CategoryIds,
			Description:           entity.Description,
			Images:                entity.Images,
			IsMarking:             entity.IsMarking,
			MeasurementUnitId:     entity.MeasurementUnitId,
			SupplierId:            entity.SupplierId,
			VatId:                 entity.VatId,
			MxikCode:              entity.MxikCode,
			ProductTypeId:         entity.ProductTypeId,
			ShopMeasurementValues: entity.MeasurementValues,
			TagIds:                entity.TagIds,
			ShopPrices:            entity.ShopPrices,
		},
		entity.Id,
	)
	if err != nil {
		return nil, err
	}

	return &common.ResponseID{Id: entity.Id}, nil
}

func (p *productRepo) Delete(req *common.RequestID) (*common.ResponseID, error) {

	query := `
	  	UPDATE
				"product"
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
		return nil, errors.Wrap(err, "error while delete product")
	}

	i, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if i == 0 {
		return nil, errors.New("product not found")
	}

	return &common.ResponseID{Id: req.Id}, nil
}

func (p *productRepo) DeleteProducts(req *common.RequestIDs) (*common.Empty, error) {

	query := `
	  	UPDATE
				"product"
	  	SET
				deleted_at = extract(epoch from now())::bigint
	  	WHERE
				deleted_at = 0 AND id = ANY($1) AND company_id = $2
	`

	res, err := p.db.Exec(query, pq.Array(req.Ids), req.Request.CompanyId)
	if err != nil {
		return nil, errors.Wrap(err, "error while delete products")
	}

	i, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if int(i) != len(req.Ids) {
		return nil, errors.New("products not found")
	}

	return &common.Empty{}, nil
}

func (p *productRepo) getProductBarcodes(productDetailId string) ([]string, error) {

	var (
		barcodes = make([]string, 0)
	)

	query := `
		SELECT 
			barcode
		FROM 
			"product_barcode"
		WHERE
			product_detail_id = $1
			
	`

	rows, err := p.db.Query(query, productDetailId)
	if err != nil {
		return nil, errors.Wrap(err, "error while getting product barcodes")
	}

	defer rows.Close()

	for rows.Next() {

		var barcode string

		err = rows.Scan(&barcode)
		if err != nil {
			return nil, errors.Wrap(err, "error while scanning product barcodes")
		}

		barcodes = append(barcodes, barcode)
	}

	return barcodes, nil
}

func (p *productRepo) getProductCategories(productDetailId string) ([]*catalog_service.ShortCategory, error) {

	var (
		categories = make([]*catalog_service.ShortCategory, 0)
	)

	query := `
		SELECT 
			cat.id,
			cat.name
		FROM 
			"product_category" pc
		JOIN "category" cat ON cat.id = pc.category_id AND cat.deleted_at = 0
		WHERE
			pc.product_detail_id = $1
			
	`

	rows, err := p.db.Query(query, productDetailId)
	if err != nil {
		return nil, errors.Wrap(err, "error while getting product categories")
	}

	defer rows.Close()

	for rows.Next() {

		var category catalog_service.ShortCategory

		err = rows.Scan(&category.Id, &category.Name)
		if err != nil {
			return nil, errors.Wrap(err, "error while scanning product categories")
		}

		categories = append(categories, &category)
	}

	return categories, nil
}

func (p *productRepo) getProductImages(productDetailId string) ([]*catalog_service.ProductImage, error) {

	var (
		images = make([]*catalog_service.ProductImage, 0)
	)

	query := `
		SELECT
			sequence_number,
			file_name
		FROM 
			"product_image"
		WHERE
			product_detail_id = $1
	`

	rows, err := p.db.Query(query, productDetailId)
	if err != nil {
		return nil, errors.Wrap(err, "error while getting product images")
	}

	defer rows.Close()

	for rows.Next() {

		var image catalog_service.ProductImage

		err = rows.Scan(&image.SequenceNumber, &image.ImageUrl)
		if err != nil {
			return nil, errors.Wrap(err, "error while getting product images")
		}

		image.ImageUrl = fmt.Sprintf("https://%s/%s/%s", p.cfg.MinioEndpoint, config.FileBucketName, image.ImageUrl)

		images = append(images, &image)
	}

	return images, nil
}

func (p *productRepo) getProductMeasurementValues(productId string, productDetailId string) ([]*catalog_service.ShopMeasurementValue, error) {

	var (
		measurementValues = make([]*catalog_service.ShopMeasurementValue, 0)
		shopName          sql.NullString
	)

	query := `
		SELECT
			mv.shop_id,
			mv.amount,
			mv.has_trigger,
			mv.is_available,
			mv.small_left,
			sh.name
		FROM 
			"measurement_values" mv
		LEFT JOIN "shop" sh ON sh.deleted_at = 0 AND mv.shop_id = sh.id 
		WHERE
			mv.product_id = $1
	`

	rows, err := p.db.Query(query, productId)
	if err != nil {
		return nil, errors.Wrap(err, "error while getting product measurementValues. Query")
	}

	defer rows.Close()

	for rows.Next() {

		var measurementValue catalog_service.ShopMeasurementValue

		err = rows.Scan(
			&measurementValue.ShopId,
			&measurementValue.Amount,
			&measurementValue.HasTrigger,
			&measurementValue.IsAvailable,
			&measurementValue.SmallLeft,
			&shopName,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error while getting product measurementValues. Scan")
		}

		measurementValue.ShopName = shopName.String

		// measurementValues[measurementValue.ShopId] = &measurementValue
		measurementValues = append(measurementValues, &measurementValue)
	}

	return measurementValues, nil
}

func (p *productRepo) getProductShopPrices(productIds []string) (map[string][]*catalog_service.ShopPrice, error) {

	var (
		res      = make(map[string][]*catalog_service.ShopPrice, 0)
		shopName sql.NullString
	)

	query := `
		SELECT
			shp.product_id,
			shp.min_price,
			shp.max_price,
			shp.retail_price,
			shp.supply_price,
			shp.whole_sale_price,
			shp.shop_id,
			sh.name
		FROM 
			"shop_price" shp
		LEFT JOIN "shop" sh ON sh.deleted_at = 0 AND shp.shop_id = sh.id 
		WHERE
			shp.product_id = ANY($1)
	`

	rows, err := p.db.Query(query, pq.Array(productIds))
	if err != nil {
		return nil, errors.Wrap(err, "error while getting product shop_price. Query")
	}

	defer rows.Close()

	for rows.Next() {

		var (
			shopPrice       catalog_service.ShopPrice
			productDetailId string
		)

		err = rows.Scan(
			&productDetailId,
			&shopPrice.MinPrice,
			&shopPrice.MaxPrice,
			&shopPrice.RetailPrice,
			&shopPrice.SupplyPrice,
			&shopPrice.WholeSalePrice,
			&shopPrice.ShopId,
			&shopName,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error while getting product shop_price. Scan")
		}

		shopPrice.ShopName = shopName.String

		shopPrices, ok := res[productDetailId]
		if ok {
			res[productDetailId] = append(res[productDetailId], &shopPrice)
		} else {
			shopPrices = append(shopPrices, &shopPrice)

			res[productDetailId] = shopPrices
		}
	}

	return res, nil
}

func (p *productRepo) getProductsShopPriceForLabel(shopId string, productDetailIds []string) (map[string]*models.ProductShopPrice, error) {

	var (
		res      = make(map[string]*models.ProductShopPrice, 0)
		shopName sql.NullString
	)

	query := `
		SELECT
			shp.product_id,
			shp.min_price,
			shp.max_price,
			shp.retail_price,
			shp.whole_sale_price,
			shp.shop_id,
			sh.name
		FROM 
			"shop_price" shp
		LEFT JOIN "shop" sh ON sh.deleted_at = 0 AND shp.shop_id = sh.id 
		WHERE
			shp.shop_id = $1 AND shp.product_detail_id = ANY($2)
	`

	rows, err := p.db.Query(query, shopId, pq.Array(productDetailIds))
	if err != nil {
		return nil, errors.Wrap(err, "error while getting product shop_price. Query")
	}

	defer rows.Close()

	for rows.Next() {

		var (
			shopPrice       models.ProductShopPrice
			productDetailId string
		)

		err = rows.Scan(
			&productDetailId,
			&shopPrice.MinPrice,
			&shopPrice.MaxPrice,
			&shopPrice.RetailPrice,
			&shopPrice.WholeSalePrice,
			&shopPrice.ShopId,
			&shopName,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error while getting product shop_price. Scan")
		}

		shopPrice.ShopName = shopName.String

		res[productDetailId] = &shopPrice
	}

	return res, nil
}

func (p *productRepo) getProductBarcodesForLabel(productDetailIds []string) (map[string][]string, error) {

	var (
		res = make(map[string][]string, 0)
	)

	query := `
		SELECT 
			barcode,
			product_detail_id
		FROM 
			"product_barcode"
		WHERE
			product_detail_id = ANY($1)
	`

	rows, err := p.db.Query(query, pq.Array(productDetailIds))
	if err != nil {
		return nil, errors.Wrap(err, "error while getting product barcodes")
	}

	defer rows.Close()

	for rows.Next() {

		var (
			barcode         string
			productDetailId string
		)

		err = rows.Scan(&barcode, &productDetailId)
		if err != nil {
			return nil, errors.Wrap(err, "error while scanning product barcodes")
		}

		barcodes, ok := res[productDetailId]
		if ok {
			barcodes = append(barcodes, barcode)

			res[productDetailId] = barcodes
		} else {
			res[productDetailId] = []string{barcode}
		}
	}

	return res, nil
}

func (p *productRepo) UpsertShopMeasurmentValue(req *catalog_service.UpsertShopMeasurmentValueRequest) error {

	var (
		values []interface{}
	)

	query := `
		INSERT INTO
			"measurement_values"
		(
			shop_id,
			amount,
			product_id
		)
		VALUES
	`

	for _, v := range req.ProductsValues {

		query += `(?, ?, ?),`

		values = append(values, req.ShopId, v.Amount, v.ProductId)

	}

	query = strings.TrimSuffix(query, ",")
	query = helper.ReplaceSQL(query, "?")

	query += `
		ON CONFLICT (product_id, shop_id) DO UPDATE SET amount = EXCLUDED.amount
	`

	_, err := p.db.Exec(query, values...)
	if err != nil {
		return errors.Wrap(err, "error while upsertShopMeasurementValues")
	}

	return nil
}

func (p *productRepo) UpsertShopRetailPrice(req *catalog_service.UpsertShopPriceRequest) error {

	var (
		values []interface{}
	)

	query := `
		INSERT INTO
			"shop_price"
		(
			id,
			shop_id,
			retail_price,
			supply_price,
			product_id
		)
		VALUES
	`

	for _, v := range req.ProductsValues {

		query += `(?, ?, ?, ?, ?),`

		values = append(values, uuid.NewString(), v.Price.ShopId, v.Price.RetailPrice, v.Price.SupplyPrice, v.ProductId)

	}

	query = strings.TrimSuffix(query, ",")
	query = helper.ReplaceSQL(query, "?")

	query += `
		ON CONFLICT (product_id, shop_id) DO UPDATE SET retail_price = EXCLUDED.retail_price
	`

	_, err := p.db.Exec(query, values...)
	if err != nil {
		return errors.Wrap(err, "error while upsert shop retail price")
	}

	return nil
}

func (p *productRepo) getProductDetailByProductIds(ids []string) (map[string]int, error) {

	var res = make(map[string]int)

	// set default version
	for _, id := range ids {
		res[id] = 1
	}

	query := `
		SELECT "product_id", MAX("version") version
		FROM "product_detail"
		WHERE product_id = ANY($1)
		GROUP BY product_id
	`

	rows, err := p.db.Query(query, pq.Array(ids))
	if err != nil {
		return nil, errors.Wrap(err, "error while getting productDetails. Query")
	}

	defer rows.Close()

	for rows.Next() {

		var (
			productId string
			version   int
		)

		err = rows.Scan(&productId, &version)
		if err != nil {
			return nil, errors.Wrap(err, "error while getting productDetails. rows.Scan")
		}
		res[productId] = version + 1
	}

	return res, nil
}

func (p *productRepo) InsertMany(products []*common.CreateProductCopyRequest) error {

	var (
		values            = []interface{}{}
		productBarcodes   = []interface{}{}
		measurementValues = []interface{}{}
		shopPrices        = []interface{}{}
		productIds        = make([]string, 0)
	)

	if len(products) <= 0 {
		return nil
	}

	query := `
		INSERT INTO
			"product"
		(
			id,
			company_id,
			product_type_id,
			created_by
		)
		VALUES
	`

	for _, product := range products {

		query += "(?, ?, ?, ?),"
		values = append(values,
			product.Id,
			product.Request.CompanyId,
			product.ProductTypeId,
			product.Request.UserId,
		)

		productIds = append(productIds, product.Id)
	}

	query = strings.TrimSuffix(query, ",")
	query = helper.ReplaceSQL(query, "?")

	query += `
		ON CONFLICT (id) DO NOTHING
	`
	productDetails, err := p.getProductDetailByProductIds(productIds)
	if err != nil {
		return err
	}

	stmt, err := p.db.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "error while insertMany products. Prepare")
	}
	defer stmt.Close()

	_, err = stmt.Exec(values...)
	if err != nil {
		return errors.Wrap(err, "error while insertMany products. Exec")
	}

	// insert product details
	queryProductDetails := `
		INSERT INTO
			"product_detail"
		(
			version,
			id,
			product_id,
			sku,
			name,
			mxik_code,
			is_marking,
			brand_id,
			description,
			measurement_unit_id,
			created_by,
			supplier_id,
			vat_id
		)
		VALUES
	`

	// insert product barcodes query
	productBarcodesQuery := `
		INSERT INTO
			"product_barcode"
		(
			barcode,
			product_detail_id
		)
		VALUES 
	`

	// insert measurement values query
	measurementValuesQuery := `
		INSERT INTO
			"measurement_values"
		(
			shop_id,
		 	product_id,
			is_available,
			has_trigger,
			amount,
			small_left
		)
		VALUES 
	`

	// insert shop_price query
	shopPricesQuery := `
		INSERT INTO
			"shop_price"
		(
			id,
			product_id,
			shop_id,
			min_price,
			max_price,
			supply_price,
			retail_price,
			whole_sale_price
		)
		VALUES 
	`

	values = []interface{}{}
	for _, product := range products {

		var (
			productDetailId = uuid.NewString()
			brandId         sql.NullString
			supplierId      sql.NullString
			vatId           sql.NullString
			version         = productDetails[product.Id]
		)

		if product.BrandId != "" {
			brandId.String = product.BrandId
			brandId.Valid = true
		}

		queryProductDetails += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"
		values = append(values,
			version,
			productDetailId,
			product.Id,
			product.Sku,
			product.Name,
			product.MxikCode,
			product.IsMarking,
			brandId,
			product.Description,
			product.MeasurementUnitId,
			product.Request.UserId,
			supplierId,
			vatId,
		)

		// collect product barcodes
		for _, barcode := range product.Barcode {
			productBarcodesQuery += "(?, ?),"
			productBarcodes = append(productBarcodes, barcode, productDetailId)
		}

		// collect product measurement values and shopPrices
		if len(product.ShopMeasurementValues) > 0 {

			for _, value := range product.ShopMeasurementValues {
				measurementValuesQuery += "(?, ?, ?, ?, ?, ?),"
				measurementValues = append(measurementValues,
					value.ShopId,
					product.Id,
					value.IsAvailable,
					false,
					value.InStock,
					0,
				)

				shopPricesQuery += "(?, ?, ?, ?, ?, ?, ?, ?),"
				shopPrices = append(shopPrices,
					uuid.New().String(),
					product.Id,
					value.ShopId,
					value.MinPrice,
					value.MaxPrice,
					value.SupplyPrice,
					value.RetailPrice,
					value.WholeSalePrice,
				)
			}
		}
	}

	queryProductDetails = strings.TrimSuffix(queryProductDetails, ",")
	queryProductDetails = helper.ReplaceSQL(queryProductDetails, "?")

	queryProductDetails += `
		ON CONFLICT (product_id, version) DO NOTHING;
	`

	stmt2, err := p.db.Prepare(queryProductDetails)
	if err != nil {
		return errors.Wrap(err, "error while insertMany product_details. Prepare")
	}

	defer stmt2.Close()

	_, err = stmt2.Exec(values...)
	if err != nil {
		return errors.Wrap(err, "error while insertMany product_details. Exec")
	}

	// insert barcodes
	if len(productBarcodes) > 0 {

		productBarcodesQuery = strings.TrimSuffix(productBarcodesQuery, ",")
		productBarcodesQuery = helper.ReplaceSQL(productBarcodesQuery, "?")

		productBarcodesQuery += `
		ON CONFLICT (product_detail_id, barcode) DO NOTHING;
		`

		stmt, err = p.db.Prepare(productBarcodesQuery)
		if err != nil {
			return errors.Wrap(err, "error while insert product_barcodes. Prepare")
		}
		defer stmt.Close()

		_, err = stmt.Exec(productBarcodes...)
		if err != nil {
			return errors.Wrap(err, "error while insert product_barcodes. Exec")
		}
	}

	// insert measurementValues
	if len(measurementValues) > 0 {

		measurementValuesQuery = strings.TrimSuffix(measurementValuesQuery, ",")
		measurementValuesQuery = helper.ReplaceSQL(measurementValuesQuery, "?")

		measurementValuesQuery += `
			ON CONFLICT (product_id, shop_id)
			DO UPDATE SET
				is_available = EXCLUDED.is_available,
				has_trigger = EXCLUDED.has_trigger,
				amount = EXCLUDED.amount,
				small_left = EXCLUDED.small_left
		`

		stmt3, err := p.db.Prepare(measurementValuesQuery)
		if err != nil {
			return errors.Wrap(err, "error while insert product measurement_values. Prepare")
		}
		defer stmt3.Close()

		_, err = stmt3.Exec(measurementValues...)
		if err != nil {
			return errors.Wrap(err, "error while insert product measurement_values. Exec")
		}
	}

	// insert shopPrices
	if len(shopPrices) > 0 {
		shopPricesQuery = strings.TrimSuffix(shopPricesQuery, ",")
		shopPricesQuery = helper.ReplaceSQL(shopPricesQuery, "?")

		shopPricesQuery += `
			ON CONFLICT (product_id, shop_id)
			DO UPDATE SET
				min_price = EXCLUDED.min_price,
				max_price = EXCLUDED.max_price,
				supply_price = EXCLUDED.supply_price,
				retail_price = EXCLUDED.retail_price,
				whole_sale_price = EXCLUDED.whole_sale_price
		`

		stmt4, err := p.db.Prepare(shopPricesQuery)
		if err != nil {
			return errors.Wrap(err, "error while insert product shop prices. Prepare")
		}
		defer stmt4.Close()

		_, err = stmt4.Exec(shopPrices...)
		if err != nil {
			return errors.Wrap(err, "error while insert product shop prices. Exec")
		}
	}

	return nil
}

func (p *productRepo) GetProductCustomFields(req *common.Request) ([]*models.GetProductCustomFieldResponse, error) {

	var res []*models.GetProductCustomFieldResponse

	query := `
		SELECT
			id,
			name,
			type
		FROM "custom_field"
		WHERE company_id = $1 AND deleted_at = 0
	`

	rows, err := p.db.Query(query, req.CompanyId)
	if err != nil {
		return nil, errors.Wrap(err, "error while getting product custom fields")
	}

	defer rows.Close()

	for rows.Next() {

		var customField models.GetProductCustomFieldResponse

		err := rows.Scan(customField.Id, customField.Name, customField.Type)
		if err != nil {
			return nil, errors.Wrap(err, "error while getting product custom fields. Scan")
		}

		res = append(res, &customField)
	}

	return res, nil
}

func (p *productRepo) GetProductsForLabel(req *catalog_service.GetAllProductsRequest) ([]map[string]string, error) {

	var (
		res              = make([]map[string]string, 0)
		products         = make([]*models.GetProductForLabel, 0)
		productDetailIds = make([]string, 0)
	)

	query := `
		SELECT 
			p.id,
			pd.id,
			pd.name,
			pd.sku,
			pd.mxik_code,
			pd.is_marking,
			pd.description
		FROM 
			"product" p
		JOIN  "product_detail" pd ON p.id = pd.product_id AND p.last_version = pd.version
		WHERE
			p.id = $1 AND p.deleted_at = 0 AND p.company_id = $2
	`

	rows, err := p.db.Query(query, req.Request.CompanyId)
	if err != nil {
		return nil, errors.Wrap(err, "error while getting product")
	}

	for rows.Next() {

		var (
			product         models.GetProductForLabel
			productDetailId string
		)

		err = rows.Scan(
			&product.Id,
			&productDetailId,
			&product.Name,
			&product.Sku,
			&product.MxikCode,
			&product.IsMarking,
			&product.Description,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error while getting product")
		}

		products = append(products, &product)
		productDetailIds = append(productDetailIds, productDetailId)
	}

	shopPrices, err := p.getProductsShopPriceForLabel(req.Search, productDetailIds)
	if err != nil {
		return nil, errors.Wrap(err, "error while getting getProductsShopPriceForLabel")
	}

	barcodes, err := p.getProductBarcodesForLabel(productDetailIds)
	if err != nil {
		return nil, errors.Wrap(err, "error while getting getProductsShopPriceForLabel")
	}

	for _, product := range products {
		product.ShopPrice = shopPrices[product.Id]
		product.Barcodes = barcodes[product.Id]
	}

	for _, product := range products {
		r := map[string]string{
			"id":           product.Id,
			"name":         product.Name,
			"barcode":      "",
			"sku":          product.Sku,
			"mxik_code":    product.MxikCode,
			"date":         "01.01.2000",
			"shop_name":    product.ShopPrice.ShopName,
			"retail_price": strconv.FormatFloat(product.ShopPrice.RetailPrice, 'E', -1, 32),
		}

		if len(product.Barcodes) > 0 {
			r["barcode"] = product.Barcodes[0]
		}

		res = append(res, r)
	}

	return res, nil
}

func (p *productRepo) ProductBulkEdit(req *catalog_service.ProductBulkOperationRequest) (*common.ResponseID, error) {

	var (
		resposeID = uuid.NewString()
	)

	if req.ProductField == "name" {
		nameQuery := `
			UPDATE
				"product_detail" AS pd
			SET
				name = $2
			WHERE
				(pd.product_id, pd.version) = (select id, last_version FROM product WHERE id = pd.product_id AND deleted_at = 0) AND pd.product_id = ANY($1)
	`

		res, err := p.db.Exec(nameQuery, pq.Array(req.ProductIds), req.Value)
		if err != nil {
			return nil, errors.Wrap(err, "error while update product_detail name")
		}

		if i, _ := res.RowsAffected(); i == 0 {
			return nil, sql.ErrNoRows
		}
	}

	if req.ProductField == "measurement_value" {
		measurementValueQuery := `
			UPDATE
				"product_detail" AS pd
			SET
				measurement_unit_id = $2
			WHERE
			(pd.product_id, pd.version) = (select id, last_version FROM product WHERE id = pd.product_id AND deleted_at = 0) AND pd.product_id = ANY($1)
	`

		res, err := p.db.Exec(measurementValueQuery, pq.Array(req.ProductIds), req.Value)
		if err != nil {
			return nil, errors.Wrap(err, "error while update product_detail measurementValue")
		}

		if i, _ := res.RowsAffected(); i == 0 {
			return nil, sql.ErrNoRows
		}
	}

	if req.ProductField == "category" {

		var productCategory = make([]interface{}, 0)

		categoryValueQuery := `
			INSERT INTO "product_category"
				("product_detail_id", "category_id")
			VALUES
		`

		for _, productId := range req.ProductIds {

			categoryValueQuery += `
			(
				(
					SELECT pd.id FROM "product" p
					JOIN "product_detail" pd ON pd.product_id = p.id AND pd.version = p.last_version
					WHERE p.id = ?
				), 
				?
			),`
			productCategory = append(productCategory,
				productId,
				req.Value,
			)
		}

		categoryValueQuery = strings.TrimSuffix(categoryValueQuery, ",")
		categoryValueQuery = helper.ReplaceSQL(categoryValueQuery, "?")

		categoryValueQuery += `
			ON CONFLICT ("product_detail_id", "category_id")
			DO NOTHING
	`

		stmt4, err := p.db.Prepare(categoryValueQuery)
		if err != nil {
			return nil, errors.Wrap(err, "error while insert product category. Prepare")
		}
		defer stmt4.Close()

		_, err = stmt4.Exec(productCategory...)
		if err != nil {
			return nil, errors.Wrap(err, "error while insert product category. Exec")
		}
	}

	if req.ProductField == "low_stock" {

		stockValueQuery := `
			UPDATE
				"measurement_values" AS mv
			SET
				small_left = $2
			WHERE
				mv.product_id = ANY($1) AND mv.shop_id = ANY($3)
	`

		res, err := p.db.Exec(stockValueQuery, pq.Array(req.ProductIds), req.Value, pq.Array(req.ShopIds))
		if err != nil {
			return nil, errors.Wrap(err, "error while update measurement_values")
		}

		if i, _ := res.RowsAffected(); i == 0 {
			return nil, sql.ErrNoRows
		}
	}

	return &common.ResponseID{Id: resposeID}, nil
}

package listeners

import (
	"context"
	"genproto/common"
	"time"

	"genproto/catalog_service"

	"github.com/Invan2/invan_catalog_service/config"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/pkg/errors"
)

func (c *catalogService) CreateProduct(ctx context.Context, req *catalog_service.CreateProductRequest) (*common.ResponseID, error) {

	var (
		measurementValues      = make(map[string]*catalog_service.ShopMeasurementValue, 0)
		shopPrices             = make(map[string]*catalog_service.ShopPrice, 0)
		kafkaMeasurementValues = make([]*common.CommonShopMeasurementValue, 0)
	)

	measurementUnit, err := c.strg.MeasurementUnit().GetByID(&common.RequestID{Id: req.MeasurementUnitId, Request: req.Request})
	if err != nil {
		return nil, err
	}

	supplier, err := c.strg.Supplier().GetById(&common.RequestID{Id: req.SupplierId, Request: req.Request})
	if err != nil {
		return nil, err
	}

	vat, err := c.strg.Vat().GetById(ctx, &common.RequestID{Id: req.VatId, Request: req.Request})
	if err != nil {
		return nil, err
	}

	tr, err := c.strg.WithTransaction()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tr.Rollback()
		} else {
			_ = tr.Commit()
		}
	}()

	productId, _, err := tr.Product().Create(req)
	if err != nil {
		return nil, err
	}

	shopNames, err := c.strg.Shop().GetCompanyAllShopNames(req.Request)
	if err != nil {
		return nil, err
	}

	categories, err := c.strg.Category().GetShortCategoriesByIds(req.CategoryIds)
	if err != nil {
		return nil, err
	}

	for _, value := range req.ShopMeasurementValues {
		measurementValues[value.ShopId] = &catalog_service.ShopMeasurementValue{
			ShopId:      value.ShopId,
			ShopName:    shopNames[value.ShopId],
			HasTrigger:  value.HasTrigger,
			SmallLeft:   float32(value.SmallLeft),
			IsAvailable: value.IsAvailable,
			Amount:      value.Amount,
		}
	}

	for _, shopPrice := range req.ShopPrices {
		shopPrices[shopPrice.ShopId] = &catalog_service.ShopPrice{
			ShopId:         shopPrice.ShopId,
			ShopName:       shopNames[shopPrice.ShopId],
			SupplyPrice:    float32(shopPrice.SupplyPrice),
			RetailPrice:    float32(shopPrice.RetailPrice),
			WholeSalePrice: float32(shopPrice.WholeSalePrice),
			MinPrice:       float32(shopPrice.MinPrice),
			MaxPrice:       float32(shopPrice.MaxPrice),
		}
	}

	productEs := &catalog_service.ProductES{
		Id:            productId,
		ParentId:      req.ParentId,
		Name:          req.Name,
		Barcodes:      req.Barcodes,
		Sku:           req.Sku,
		MxikCode:      req.MxikCode,
		Description:   req.Description,
		IsMarking:     req.IsMarking,
		ProductTypeId: req.ProductTypeId,
		CompanyId:     req.Request.CompanyId,
		MeasurementUnit: &catalog_service.ShortMeasurementUnit{
			Id:          measurementUnit.Id,
			ShortName:   measurementUnit.ShortName,
			LongName:    measurementUnit.LongName,
			Precision:   measurementUnit.Precision,
			IsDeletable: measurementUnit.IsDeletable,
		},
		Supplier: &catalog_service.ShortSupplier{
			Id:   supplier.Id,
			Name: supplier.Name,
		},
		Vat: &catalog_service.ShortVat{
			Id:         vat.Id,
			Name:       vat.Name,
			Percentage: vat.Percentage,
		},
		Image:             "",
		MeasurementValues: measurementValues,
		Categories:        categories,
		ShopPrices:        shopPrices,
		CreatedAt:         time.Now().Format(config.DateTimeFormat),
		UpdatedAt:         float64(time.Now().UnixMilli()),
	}

	if len(req.Images) > 0 {
		productEs.Image = req.Images[0].ImageUrl
	}

	for _, value := range measurementValues {
		kafkaMeasurementValues = append(kafkaMeasurementValues, &common.CommonShopMeasurementValue{
			IsAvailable:    value.IsAvailable,
			InStock:        value.Amount,
			ShopId:         value.ShopId,
			RetailPrice:    shopPrices[value.ShopId].RetailPrice,
			SupplyPrice:    shopPrices[value.ShopId].SupplyPrice,
			MinPrice:       shopPrices[value.ShopId].MinPrice,
			MaxPrice:       shopPrices[value.ShopId].MaxPrice,
			WholeSalePrice: shopPrices[value.ShopId].WholeSalePrice,
		})
	}

	err = c.kafka.Push("v1.catalog_service.product.created.success", common.CreateProductCopyRequest{
		Id:                    productId,
		IsMarking:             req.IsMarking,
		Sku:                   req.Sku,
		Name:                  req.Name,
		MeasurementUnitId:     req.MeasurementUnitId,
		SupplierId:            req.SupplierId,
		VatId:                 req.VatId,
		MxikCode:              req.MxikCode,
		BrandId:               req.BrandId,
		Description:           req.Description,
		ProductTypeId:         req.ProductTypeId,
		ParentId:              req.ParentId,
		ShopMeasurementValues: kafkaMeasurementValues,
		Image:                 productEs.Image,
		Barcode:               req.Barcodes,
		Request:               req.Request,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error while creating product")
	}

	err = c.elastic.Product().Create(productEs)
	if err != nil {
		return nil, errors.Wrap(err, "error while creating product. Elastic")
	}

	return &common.ResponseID{Id: productId}, nil
}

func (c *catalogService) GetProductByID(ctx context.Context, req *common.RequestID) (*catalog_service.Product, error) {
	return c.strg.Product().GetByID(req)
}

func (c *catalogService) UpdateProduct(ctx context.Context, req *catalog_service.UpdateProductRequest) (*common.ResponseID, error) {

	var (
		shopMeasurementValues  = make(map[string]*catalog_service.ShopMeasurementValue)
		shopPrices             = make(map[string]*catalog_service.ShopPrice)
		kafkaMeasurementValues = make([]*common.CommonShopMeasurementValue, 0)
	)
	tr, err := c.strg.WithTransaction()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tr.Rollback()
		} else {
			_ = tr.Commit()
		}
	}()

	res, err := tr.Product().Update(req)
	if err != nil {
		return nil, err
	}

	measurementUnit, err := c.strg.MeasurementUnit().GetByID(&common.RequestID{Id: req.MeasurementUnitId, Request: req.Request})
	if err != nil {
		return nil, err
	}

	supplier, err := c.strg.Supplier().GetById(&common.RequestID{Id: req.SupplierId, Request: req.Request})
	if err != nil {
		return nil, err
	}

	vat, err := c.strg.Vat().GetById(ctx, &common.RequestID{Id: req.VatId, Request: req.Request})
	if err != nil {
		return nil, err
	}

	for _, measurementValues := range req.MeasurementValues {
		shopMeasurementValues[measurementValues.ShopId] = &catalog_service.ShopMeasurementValue{
			ShopId:      measurementValues.ShopId,
			ShopName:    measurementValues.ShopName,
			Amount:      float32(measurementValues.Amount),
			SmallLeft:   float32(measurementValues.SmallLeft),
			HasTrigger:  measurementValues.HasTrigger,
			IsAvailable: measurementValues.IsAvailable,
		}
	}

	for _, shopPrice := range req.ShopPrices {
		shopPrices[shopPrice.ShopId] = &catalog_service.ShopPrice{
			ShopId:         shopPrice.ShopId,
			ShopName:       shopPrice.ShopName,
			SupplyPrice:    float32(shopPrice.SupplyPrice),
			RetailPrice:    float32(shopPrice.RetailPrice),
			WholeSalePrice: float32(shopPrice.WholeSalePrice),
			MinPrice:       float32(shopPrice.MinPrice),
			MaxPrice:       float32(shopPrice.MaxPrice),
		}
	}

	categories, err := tr.Category().GetShortCategoriesByIds(req.CategoryIds)
	if err != nil {
		return nil, err
	}

	productEs := &catalog_service.ProductES{
		Id:            req.Id,
		ParentId:      req.ParentId,
		Name:          req.Name,
		Barcodes:      req.Barcodes,
		Sku:           req.Sku,
		MxikCode:      req.MxikCode,
		Description:   req.Description,
		IsMarking:     req.IsMarking,
		ProductTypeId: req.ProductTypeId,
		CompanyId:     req.Request.CompanyId,
		MeasurementUnit: &catalog_service.ShortMeasurementUnit{
			Id:          measurementUnit.Id,
			ShortName:   measurementUnit.ShortName,
			LongName:    measurementUnit.LongName,
			Precision:   measurementUnit.Precision,
			IsDeletable: measurementUnit.IsDeletable,
		},
		Supplier: &catalog_service.ShortSupplier{
			Id:   supplier.Id,
			Name: supplier.Name,
		},
		Vat: &catalog_service.ShortVat{
			Id:         vat.Id,
			Name:       vat.Name,
			Percentage: vat.Percentage,
		},
		Image:             "",
		MeasurementValues: shopMeasurementValues,
		Categories:        categories,
		ShopPrices:        shopPrices,
		// CreatedAt:         time.Now().Format(config.DateTimeFormat),
		UpdatedAt: float64(time.Now().UnixMilli()),
	}

	if len(req.Images) > 0 {
		productEs.Image = req.Images[0].ImageUrl
	}

	c.log.Info("product", logger.Any("data", productEs))

	for _, value := range shopMeasurementValues {
		kafkaMeasurementValues = append(kafkaMeasurementValues, &common.CommonShopMeasurementValue{
			IsAvailable:    value.IsAvailable,
			InStock:        value.Amount,
			RetailPrice:    shopPrices[value.ShopId].RetailPrice,
			SupplyPrice:    shopPrices[value.ShopId].SupplyPrice,
			MinPrice:       shopPrices[value.ShopId].MinPrice,
			MaxPrice:       shopPrices[value.ShopId].MaxPrice,
			WholeSalePrice: shopPrices[value.ShopId].WholeSalePrice,
			ShopId:         value.ShopId,
		})
	}

	err = c.kafka.Push("v1.catalog_service.product.created.success", &common.CreateProductCopyRequest{
		Id:                    req.Id,
		Sku:                   req.Sku,
		Name:                  req.Name,
		Image:                 productEs.Image,
		BrandId:               req.BrandId,
		MxikCode:              req.MxikCode,
		ParentId:              req.ParentId,
		Description:           req.Description,
		ProductTypeId:         req.ProductTypeId,
		MeasurementUnitId:     req.MeasurementUnitId,
		SupplierId:            req.SupplierId,
		VatId:                 req.VatId,
		IsMarking:             req.IsMarking,
		Barcode:               req.Barcodes,
		ShopMeasurementValues: kafkaMeasurementValues,
		Request:               req.Request,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error while updating product")
	}

	err = c.elastic.Product().Update(productEs)
	if err != nil {
		return nil, errors.Wrap(err, "error while updating product Elastic")
	}

	return res, nil
}

func (c *catalogService) GetAllProducts(ctx context.Context, req *catalog_service.GetAllProductsRequest) (*catalog_service.GetAllProductsResponse, error) {
	c.log.Info("GetAllProducts", logger.Any("request", req))
	return c.elastic.Product().GetAll(req)
}

func (c *catalogService) DeleteProductsByIds(ctx context.Context, req *common.RequestIDs) (*common.Empty, error) {

	tr, err := c.strg.WithTransaction()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tr.Rollback()
		} else {
			_ = tr.Commit()
		}
	}()

	res, err := tr.Product().DeleteProducts(req)
	if err != nil {
		return nil, err
	}

	_, err = c.elastic.Product().DeleteProducts(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *catalogService) DeleteProductById(ctx context.Context, req *common.RequestID) (*common.ResponseID, error) {

	tr, err := c.strg.WithTransaction()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tr.Rollback()
		} else {
			_ = tr.Commit()
		}
	}()

	_, err = tr.Product().Delete(req)
	if err != nil {
		return nil, err
	}

	_, err = c.elastic.Product().DeleteProduct(req)
	if err != nil {
		return nil, err
	}

	return &common.ResponseID{Id: req.Id}, nil
}

func (c *catalogService) SearchProducts(ctx context.Context, req *catalog_service.GetAllProductsRequest) (*catalog_service.SearchProductsResponse, error) {

	return c.elastic.Product().SearchProducts(req)
}

func (c *catalogService) BulkUpdateProduct(ctx context.Context, req *catalog_service.ProductBulkOperationRequest) (*common.ResponseID, error) {

	var (
		productMap      = make(map[string]*catalog_service.ProductES)
		measurementUnit catalog_service.ShortMeasurementUnit
		categories      = make([]*catalog_service.ShortCategory, 0)
	)

	if req.ProductField == "measurement_value" {
		mu, err := c.strg.MeasurementUnit().GetByID(&common.RequestID{Id: req.Value, Request: req.Request})
		if err != nil {
			return nil, err
		}

		measurementUnit = catalog_service.ShortMeasurementUnit{
			Id:                   mu.Id,
			ShortName:            mu.ShortName,
			LongName:             mu.LongName,
			IsDeletable:          mu.IsDeletable,
			Precision:            mu.Precision,
			LongNameTranslation:  mu.LongNameTranslation,
			ShortNameTranslation: mu.ShortNameTranslation,
			CreatedAt:            mu.CreatedAt,
			CreatedBy:            mu.CreatedBy,
		}

	}

	if req.ProductField == "category" {

		cat, err := c.strg.Category().GetByID(&common.RequestID{Id: req.Value, Request: req.Request})
		if err != nil {
			return nil, errors.Wrap(err, "error while getting categories")
		}

		categories = append(categories, &catalog_service.ShortCategory{
			Id:       cat.Category.Id,
			Name:     cat.Category.Name,
			ParentId: cat.Category.ParentId,
		})

	}

	for _, val := range req.ProductIds {

		productMap[val] = &catalog_service.ProductES{
			Name:            req.Value,
			MeasurementUnit: &measurementUnit,
			Categories:      categories,
		}

	}

	tr, err := c.strg.WithTransaction()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tr.Rollback()
		} else {
			_ = tr.Commit()
		}
	}()

	err = c.kafka.Push("v1.catalog_service.product.bulk_updated.success", catalog_service.ProductBulkOperationRequest{
		ProductIds:   req.ProductIds,
		ShopIds:      req.ShopIds,
		ProductField: req.ProductField,
		Value:        req.Value,
		Request:      req.Request,
	})

	if err != nil {
		return nil, err
	}

	res, err := tr.Product().ProductBulkEdit(req)
	if err != nil {
		return nil, err
	}

	err = c.elastic.Product().BulkUpdateProduct(req, productMap)
	if err != nil {
		return nil, err
	}
	return res, nil
}

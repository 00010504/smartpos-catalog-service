package elastic

import (
	"bytes"
	"context"
	"fmt"
	"genproto/catalog_service"
	"genproto/common"
	"io"
	"strings"
	"time"

	"github.com/clarketm/json"

	"github.com/Invan2/invan_catalog_service/config"
	"github.com/Invan2/invan_catalog_service/models"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/Invan2/invan_catalog_service/storage/repo"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

type productRepo struct {
	db  *elasticsearch.Client
	log logger.Logger
	cfg config.Config
}

func NewProductRepo(log logger.Logger, db *elasticsearch.Client, cfg config.Config) repo.ProductESI {
	return &productRepo{
		db:  db,
		log: log,
		cfg: cfg,
	}
}

type H map[string]interface{}

func (p *productRepo) Create(product *catalog_service.ProductES) error {

	p.log.Info("create product on elastic", logger.Any("data", product))

	if !exists(p.db, config.ElasticProductIndex) {

		query := H{
			"mappings": H{
				"properties": H{
					"measurement_values": H{
						"type": "flattened",
					},
					"shop_prices": H{
						"type": "flattened",
					},
					"updated_at": H{
						"type": "text",
						"fields": H{
							"keyword": H{
								"type": "keyword",
							},
						},
					},
				},
			},
		}

		body, err := json.Marshal(query)
		if err != nil {
			return errors.Wrap(err, "error while marshaling mapping query")
		}

		esReq := esapi.IndicesCreateRequest{
			Index: config.ElasticProductIndex,
			Body:  bytes.NewReader(body),
		}

		res, err := esReq.Do(context.Background(), p.db)
		if err != nil {
			return errors.Wrap(err, "error while create index")
		}

		if res.IsError() {
			data, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			p.log.Error("errror while create product index", logger.Any("res", string(data)))
			return errors.New("error while create products index on elastic")
		}
	}

	var (
		body bytes.Buffer
	)

	err := config.JSONPBMarshaler.Marshal(&body, product)
	if err != nil {
		return errors.Wrap(err, "error while marshaling, jsonpb")
	}

	res, err := p.db.Create(config.ElasticProductIndex, product.Id, bytes.NewReader(body.Bytes()), p.db.Create.WithRefresh("true"))
	if err != nil {
		return errors.Wrap(err, "Failed to bulk insert products")
	}
	defer res.Body.Close()

	if res.IsError() {
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		p.log.Error("errror while create product", logger.Any("res", string(data)))
		return errors.New("error while create product on elastic")
	}

	return nil
}

func (p *productRepo) Update(product *catalog_service.ProductES) error {
	if !exists(p.db, config.ElasticProductIndex) {
		res, err := p.db.Indices.Create(config.ElasticProductIndex)
		if err != nil {
			return errors.Wrap(err, "error while create index")
		}

		if res.IsError() {
			data, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			p.log.Error("errror while get  products for supplier orders ", logger.Any("res", string(data)))
			return errors.New("error while  products for supplier orders" + string(data))
		}
	}

	var (
		updateReq = catalog_service.UpdateProductES{Doc: product}
		body      bytes.Buffer
	)

	err := config.JSONPBMarshaler.Marshal(&body, &updateReq)
	if err != nil {
		return errors.Wrap(err, "error while marshaling jsonpb")
	}

	res, err := p.db.Update(config.ElasticProductIndex, product.Id, bytes.NewReader(body.Bytes()), p.db.Update.WithRefresh("true"))
	if err != nil {
		return errors.Wrap(err, "error while update document on elastic")
	}

	if res.IsError() {
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		p.log.Error("errror while update ", logger.Any("res", string(data)))
		return errors.New("error while update product on elastic")
	}

	return nil
}

func makeGettAllSearchRequest(req *catalog_service.GetAllProductsRequest, statistics bool) (H, error) {

	must := make([]H, 0)
	mustNot := make([]H, 0)
	bool := make(H)
	query := make(H)

	must = append(must, H{
		"term": H{
			"company_id.keyword": H{
				"value": req.GetRequest().GetCompanyId(),
			},
		},
	})

	if len(req.CategoryIds) > 0 {
		must = append(must, H{
			"terms": H{
				"categories.id.keyword": req.GetCategoryIds(),
			},
		})
	}

	if len(req.MeasurementIds) > 0 {
		must = append(must, H{
			"terms": H{
				"measurement_unit.id.keyword": req.GetMeasurementIds(),
			},
		})
	}

	if len(req.ProductIds) > 0 {
		must = append(must, H{
			"terms": H{
				"id.keyword": req.GetProductIds(),
			},
		})
	}

	if req.Search != "" {
		must = append(must, H{
			"query_string": H{
				"query":            getSearchString(req.GetSearch()),
				"default_operator": "AND",
			},
		})
	}

	for _, field := range req.Filters {

		filterFunction, ok := filterFunctionMap[field.Key]
		if !ok {
			return nil, ErrFilterNotFound
		}

		query, err := filterFunction(field)
		if err != nil {
			return nil, errors.Wrap(err, "error while filterFunction")
		}

		if field.Relation == common.Relation_EQUAL ||
			field.Relation == common.Relation_GREATER_THAN ||
			field.Relation == common.Relation_INCLUDE {
			must = append(must, query)
		} else {
			mustNot = append(mustNot, query)
		}

	}

	if len(must) > 0 {
		bool["must"] = must
	}

	if len(bool) > 0 {
		query["bool"] = bool
	}

	// Sort
	sort := make([]H, 0)

	if req.SortBy == "" {
		sort = append(sort, H{
			"updated_at.keyword": H{
				"order": "desc",
			},
		})
	}

	var res = H{
		"query": query,
		"sort":  sort,
	}

	if statistics {
		aggs := H{
			"total_retail_price": H{
				"sum": H{
					"script": H{
						"lang": "painless",
						"params": H{
							"shop_measurement_values": "doc['shop_measurement_values']",
							"shop_prices":             "doc['shop_prices']",
						},
						"source": `
							double sum = 0;
							if (
								params == null ||
								params._source == null ||
								params._source['shop_measurement_values'] == null ||
								params._source['shop_prices'] == null
							) {
								return sum;
							}
							for (key in params._source['shop_measurement_values'].entrySet()) {
								if (
									key.getValue() != null &&
									key.getValue().amount != null &&
									params._source['shop_prices'][key.getKey()] != null &&
									params._source['shop_prices'][key.getKey()].retail_price != null
								)
									sum += key.getValue().amount * params._source['shop_prices'][key.getKey()].retail_price;
							}
							return sum
						`,
					},
					"missing": 0,
				},
			},
			"total_supply_price": H{
				"sum": H{
					"script": H{
						"lang": "painless",
						"params": H{
							"shop_measurement_values": "doc['shop_measurement_values']",
							"shop_prices":             "doc['shop_prices']",
						},
						"source": `
							double sum = 0;
							if (
								params == null ||
								params._source == null ||
								params._source['shop_measurement_values'] == null ||
								params._source['shop_prices'] == null
							) {
								return sum;
							}
							for (key in params._source['shop_measurement_values'].entrySet()) {
								if (
									key.getValue() != null &&
									key.getValue().amount != null &&
									params._source['shop_prices'][key.getKey()] != null &&
									params._source['shop_prices'][key.getKey()].supply_price != null
								)
									sum += key.getValue().amount * params._source['shop_prices'][key.getKey()].supply_price;
							}
							return sum;
						`,
					},
					"missing": 0,
				},
			},
		}

		res["aggs"] = aggs
	}

	return res, nil
}

func (p *productRepo) GetAll(entity *catalog_service.GetAllProductsRequest) (*catalog_service.GetAllProductsResponse, error) {

	var (
		res = catalog_service.GetAllProductsResponse{
			Data:       make([]*catalog_service.ProductES, 0),
			Statistics: &catalog_service.Statistics{},
			Total:      0,
		}
		r    map[string]interface{}
		buf  bytes.Buffer
		size = int(entity.Limit)
		from = int((entity.Page - 1) * entity.Limit)
	)

	req, err := makeGettAllSearchRequest(entity, true)
	if err != nil {
		return nil, err
	}

	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return nil, err
	}

	response, err := p.db.Search(
		p.db.Search.WithContext(context.Background()),
		p.db.Search.WithIndex(config.ElasticProductIndex),
		p.db.Search.WithBody(&buf),
		p.db.Search.WithTrackTotalHits(true),
		p.db.Search.WithFrom(from),
		p.db.Search.WithSize(size),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error while get documents on elastic")
	}

	if response.IsError() {
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		p.log.Error("errror while get all products ", logger.Any("res", string(data)))
		return nil, errors.New("error while get  products on elastic " + string(data))
	}

	err = json.NewDecoder(response.Body).Decode(&r)
	if err != nil {
		return nil, errors.Wrap(err, "error while json.decode elastic res.Body")
	}

	for _, source := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {

		product := catalog_service.ProductES{}

		jsonString, _ := json.Marshal(source.(map[string]interface{})["_source"])

		err = json.Unmarshal(jsonString, &product)
		if err != nil {
			return nil, errors.Wrap(err, "error while json.Unmarshal jsonString &product")
		}

		if product.Image != "" {
			product.Image = fmt.Sprintf("https://%s/%s/%s", p.cfg.MinioEndpoint, config.FileBucketName, product.Image)
		}

		res.Data = append(res.Data, &catalog_service.ProductES{
			Id:                product.Id,
			CompanyId:         product.CompanyId,
			Sku:               product.Sku,
			Name:              product.Name,
			MeasurementUnit:   product.MeasurementUnit,
			ParentId:          product.ParentId,
			Barcodes:          product.Barcodes,
			ProductTypeId:     product.ProductTypeId,
			Image:             product.Image,
			MxikCode:          product.MxikCode,
			IsMarking:         product.IsMarking,
			MeasurementValues: product.MeasurementValues,
			Supplier:          product.Supplier,
			Vat:               product.Vat,
			Description:       product.Description,
			CreatedAt:         product.CreatedAt,
			ShopPrices:        product.ShopPrices,
			Categories:        product.Categories,
		})

	}

	res.Total = int64(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))

	res.Statistics.TotalRetailPrice = cast.ToUint64(r["aggregations"].(map[string]interface{})["total_retail_price"].(map[string]interface{})["value"])
	res.Statistics.TotalSupplyPrice = cast.ToUint64(r["aggregations"].(map[string]interface{})["total_supply_price"].(map[string]interface{})["value"])
	res.Statistics.NumberOfProducts = uint64(res.Total)

	return &res, nil
}

func makeSearchRequest(req *catalog_service.GetAllProductsRequest) H {

	must := make([]H, 0)
	bool := make(H)
	query := make(H)

	must = append(must, H{
		"term": H{
			"company_id.keyword": H{
				"value": req.Request.CompanyId,
			},
		},
	})

	if req.Search != "" {
		must = append(must, H{
			"query_string": H{
				"query":            getSearchString(req.Search),
				"default_operator": "AND",
				"fields":           []string{"sku", "name", "barcodes"},
			},
		})
	}

	if len(must) > 0 {
		bool["must"] = must
	}

	if len(bool) > 0 {
		query["bool"] = bool
	}

	sort := make([]H, 0)

	if req.SortBy == "" {
		sort = append(sort, H{
			"created_at.keyword": H{
				"order": "desc",
			},
		})
	}

	return H{
		"query": query,
		"sort":  sort,
	}
}

func (p *productRepo) SearchProducts(entity *catalog_service.GetAllProductsRequest) (*catalog_service.SearchProductsResponse, error) {

	var (
		res = catalog_service.SearchProductsResponse{
			Data:  make([]*catalog_service.ProductES, 0),
			Total: 0,
		}
		r    map[string]interface{}
		buf  bytes.Buffer
		size = int(entity.Limit)
		from = int((entity.Page - 1) * entity.Limit)
	)

	req := makeSearchRequest(entity)

	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return nil, err
	}

	response, err := p.db.Search(
		p.db.Search.WithContext(context.Background()),
		p.db.Search.WithIndex(config.ElasticProductIndex),
		p.db.Search.WithBody(&buf),
		p.db.Search.WithTrackTotalHits(true),
		p.db.Search.WithFrom(from),
		p.db.Search.WithSize(size),
	)
	if err != nil || (response != nil && response.IsError()) {
		return nil, errors.New("error while get documents on elastic")
	}

	if response.IsError() {
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		p.log.Error("errror while get all products ", logger.Any("res", string(data)))
		return nil, errors.New("error while get  products on elastic " + string(data))
	}

	err = json.NewDecoder(response.Body).Decode(&r)
	if err != nil {
		return nil, errors.Wrap(err, "error while json.decode elastic res.Body")
	}

	for _, source := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {

		product := catalog_service.ProductES{}

		jsonString, _ := json.Marshal(source.(map[string]interface{})["_source"])

		err = json.Unmarshal(jsonString, &product)
		if err != nil {
			return nil, errors.Wrap(err, "error while json.Unmarshal jsonString &product")
		}

		if product.Image != "" {
			product.Image = fmt.Sprintf("https://%s/file/%s", p.cfg.MinioEndpoint, product.Image)
		}

		res.Data = append(res.Data, &catalog_service.ProductES{
			Id:                product.Id,
			CompanyId:         product.CompanyId,
			Sku:               product.Sku,
			Name:              product.Name,
			MeasurementUnit:   product.MeasurementUnit,
			Supplier:          product.Supplier,
			Vat:               product.Vat,
			ParentId:          product.ParentId,
			Barcodes:          product.Barcodes,
			ProductTypeId:     product.ProductTypeId,
			Image:             product.Image,
			MxikCode:          product.MxikCode,
			IsMarking:         product.IsMarking,
			MeasurementValues: product.MeasurementValues,
			Description:       product.Description,
			ShopPrices:        product.ShopPrices,
			Categories:        product.Categories,
			CreatedAt:         product.CreatedAt,
		})

	}

	res.Total = int64(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))

	return &res, nil
}

func (p *productRepo) DeleteProduct(req *common.RequestID) (*common.Empty, error) {

	res, err := p.db.Delete(
		config.ElasticProductIndex,
		req.Id,
		p.db.Delete.WithRefresh("true"),
	)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		p.log.Error("errror while delete products ", logger.Any("res", string(data)))
		return nil, errors.New("error while delete on elastic " + string(data))
	}

	return &common.Empty{}, nil
}

func (p *productRepo) DeleteProducts(req *common.RequestIDs) (*common.Empty, error) {

	var (
		buf bytes.Buffer
	)

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"ids": map[string]interface{}{
				"values": req.Ids,
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, errors.Wrap(err, "Error encoding query")
	}

	res, err := p.db.DeleteByQuery(
		[]string{config.ElasticProductIndex},
		&buf,
		p.db.DeleteByQuery.WithRefresh(true),
	)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		p.log.Error("errror while delete by query products ", logger.Any("res", string(data)))
		return nil, errors.New("error while delete by query on elastic " + string(data))
	}

	return &common.Empty{}, nil
}

func (p *productRepo) getOrderProducts(order *common.OrderCopyRequest) (map[string]*catalog_service.ProductES, error) {

	var (
		res    = make(map[string]*catalog_service.ProductES)
		r      map[string]interface{}
		should []H = make([]H, 0)
		buf    bytes.Buffer
	)

	for _, orderItem := range order.Items {
		should = append(should, H{"term": H{"_id": orderItem.ProductId}})
	}

	req := H{
		"query": H{
			"bool": H{
				"should": should,
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return nil, err
	}

	response, err := p.db.Search(
		p.db.Search.WithContext(context.Background()),
		p.db.Search.WithIndex(config.ElasticProductIndex),
		p.db.Search.WithBody(&buf),
		p.db.Search.WithTrackTotalHits(true),
	)
	if err != nil || (response != nil && response.IsError()) {
		return nil, errors.New("error while get documents on elastic")
	}

	if response.IsError() {
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		p.log.Error("errror while get  products for orders ", logger.Any("res", string(data)))
		return nil, errors.New("error while get producs for order on elastic " + string(data))
	}

	err = json.NewDecoder(response.Body).Decode(&r)
	if err != nil {
		return nil, errors.Wrap(err, "error while json.decode elastic res.Body")
	}

	for _, source := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		product := catalog_service.ProductES{}

		jsonString, _ := json.Marshal(source.(map[string]interface{})["_source"])

		err = json.Unmarshal(jsonString, &product)
		if err != nil {
			return nil, errors.Wrap(err, "error while json.Unmarshal jsonString &product")
		}

		res[product.Id] = &catalog_service.ProductES{
			Id:                product.Id,
			CompanyId:         product.CompanyId,
			Sku:               product.Sku,
			Name:              product.Name,
			MeasurementUnit:   product.MeasurementUnit,
			Supplier:          product.Supplier,
			Vat:               product.Vat,
			ParentId:          product.ParentId,
			Barcodes:          product.Barcodes,
			ProductTypeId:     product.ProductTypeId,
			Image:             product.Image,
			MxikCode:          product.MxikCode,
			IsMarking:         product.IsMarking,
			MeasurementValues: product.MeasurementValues,
			Description:       product.Description,
			CreatedAt:         product.CreatedAt,
			ShopPrices:        product.ShopPrices,
			Categories:        product.Categories,
			CreatedBy:         product.CreatedBy,
		}
	}

	return res, nil
}

func (p *productRepo) getSupplierOrderProducts(order *catalog_service.UpsertShopMeasurmentValueRequest) (map[string]*catalog_service.ProductES, error) {

	var (
		res    = make(map[string]*catalog_service.ProductES)
		r      map[string]interface{}
		should []H = make([]H, 0)
		buf    bytes.Buffer
	)

	for _, supplierOrderItem := range order.ProductsValues {
		should = append(should, H{"term": H{"_id": supplierOrderItem.ProductId}})
	}

	req := H{
		"query": H{
			"bool": H{
				"should": should,
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return nil, err
	}

	response, err := p.db.Search(
		p.db.Search.WithContext(context.Background()),
		p.db.Search.WithIndex(config.ElasticProductIndex),
		p.db.Search.WithBody(&buf),
		p.db.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, errors.New("error while get documents on elastic")
	}
	if response.IsError() {
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		p.log.Error("errror while get  products for supplier orders ", logger.Any("res", string(data)))
		return nil, errors.New("error while  products for supplier orders" + string(data))
	}

	err = json.NewDecoder(response.Body).Decode(&r)
	if err != nil {
		return nil, errors.Wrap(err, "error while json.decode elastic res.Body")
	}

	for _, source := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		product := catalog_service.ProductES{}

		jsonString, _ := json.Marshal(source.(map[string]interface{})["_source"])

		err = json.Unmarshal(jsonString, &product)
		if err != nil {
			return nil, errors.Wrap(err, "error while json.Unmarshal jsonString &product")
		}

		res[product.Id] = &catalog_service.ProductES{
			Id:                product.Id,
			CompanyId:         product.CompanyId,
			Sku:               product.Sku,
			Name:              product.Name,
			MeasurementUnit:   product.MeasurementUnit,
			Supplier:          product.Supplier,
			Vat:               product.Vat,
			ParentId:          product.ParentId,
			Barcodes:          product.Barcodes,
			ProductTypeId:     product.ProductTypeId,
			Image:             product.Image,
			MxikCode:          product.MxikCode,
			IsMarking:         product.IsMarking,
			MeasurementValues: product.MeasurementValues,
			Description:       product.Description,
			CreatedAt:         product.CreatedAt,
			ShopPrices:        product.ShopPrices,
			Categories:        product.Categories,
			CreatedBy:         product.CreatedBy,
		}
	}

	return res, nil
}

func (p *productRepo) UpsertShopMeasurmentValue(req *catalog_service.UpsertShopMeasurmentValueRequest) error {
	if !exists(p.db, config.ElasticProductIndex) {
		res, err := p.db.Indices.Create(config.ElasticProductIndex)
		if err != nil {
			return errors.Wrap(err, "error while create index")
		}
		if res.IsError() {
			data, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			p.log.Error("errror while get  products for supplier orders ", logger.Any("res", string(data)))
			return errors.New("error while  products for supplier orders" + string(data))
		}
	}

	var (
		productIds       = make([]string, 0)
		produtcAmountMap = make(map[string]*catalog_service.ShopMeasurementValue)
	)

	for _, val := range req.ProductsValues {
		productIds = append(productIds, val.ProductId)

		produtcAmountMap[val.ProductId] = &catalog_service.ShopMeasurementValue{
			ShopId:      req.ShopId,
			IsAvailable: true,
			Amount:      val.Amount,
		}
	}

	query := H{
		"query": H{
			"terms": H{
				"id.keyword": productIds,
			},
		},
		"script": H{
			"source": "ctx._source.measurement_values[params.products[ctx._source.id].shop_id] = params.products[ctx._source.id]",
			"lang":   "painless",
			"params": H{
				"products": produtcAmountMap,
			},
		},
	}

	body, err := json.Marshal(query)
	if err != nil {
		return err
	}

	request := esapi.UpdateByQueryRequest{
		Index: []string{config.ElasticProductIndex},
		Body:  strings.NewReader(string(body)),
	}

	res, err := request.Do(context.Background(), p.db)
	if err != nil {
		return err
	}

	if res.IsError() {
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		p.log.Error("errror while get all products ", logger.Any("res", string(data)))
		return errors.New("error while get  products on elastic " + string(data))
	}

	return nil
}

func (p *productRepo) UpsertShopPrice(req *catalog_service.UpsertShopPriceRequest) error {
	if !exists(p.db, config.ElasticProductIndex) {
		res, err := p.db.Indices.Create(config.ElasticProductIndex)
		if err != nil {
			return errors.Wrap(err, "error while create index")
		}
		if res.IsError() {
			data, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			p.log.Error("errror while get  products for supplier orders ", logger.Any("res", string(data)))
			return errors.New("error while  products for supplier orders" + string(data))
		}
	}

	var (
		productIds       = make([]string, 0)
		prodcutsPriceMap = make(map[string]*catalog_service.ShopPrice)
	)

	for _, val := range req.ProductsValues {
		productIds = append(productIds, val.ProductId)

		prodcutsPriceMap[val.ProductId] = val.Price
	}

	query := H{
		"query": H{
			"terms": H{
				"id.keyword": productIds,
			},
		},
		"script": H{
			"source": "ctx._source.shop_prices[params.products[ctx._source.id].shop_id] = params.products[ctx._source.id]",
			"lang":   "painless",
			"params": H{
				"products": prodcutsPriceMap,
			},
		},
	}

	body, err := json.Marshal(query)
	if err != nil {
		return err
	}

	request := esapi.UpdateByQueryRequest{
		Index: []string{config.ElasticProductIndex},
		Body:  strings.NewReader(string(body)),
	}

	res, err := request.Do(context.Background(), p.db)
	if err != nil {
		return err
	}

	if res.IsError() {
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		p.log.Error("errror while update products price", logger.Any("res", string(data)))
		return errors.New("error while update products price " + string(data))
	}

	return nil
}

func (p *productRepo) UpdateOnOrder(order *common.OrderCopyRequest) error {
	if !exists(p.db, config.ElasticProductIndex) {
		res, err := p.db.Indices.Create(config.ElasticProductIndex)
		if err != nil {
			return errors.Wrap(err, "error while create index")
		}

		if err := checkResponseCodeToSuccess(res.StatusCode); err != nil {
			return err
		}
	}

	productsMap, err := p.getOrderProducts(order)
	if err != nil {
		return errors.Wrap(err, "error while getting order prodcuts")
	}

	for _, orderItem := range order.Items {

		if productsMap[orderItem.ProductId] == nil || productsMap[orderItem.ProductId].MeasurementValues[order.ShopId] == nil {
			p.log.Error("errror while update product. Create order copy. product not found", logger.Any("shopId", order.ShopId))
			continue
		}
		productsMap[orderItem.ProductId].MeasurementValues[order.ShopId].Amount -= orderItem.Value

		var body bytes.Buffer

		err := config.JSONPBMarshaler.Marshal(&body, &catalog_service.UpdateProductES{Doc: productsMap[orderItem.ProductId]})
		if err != nil {
			return errors.Wrap(err, "error while marshaling")
		}

		res, err := p.db.Update(config.ElasticProductIndex, orderItem.ProductId, bytes.NewReader(body.Bytes()))
		if err != nil {
			return errors.Wrap(err, "error while update document on elastic")
		}

		if res.IsError() {
			data, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			p.log.Error("errror while update product. Create order copy ", logger.Any("res", string(data)), logger.Any("statusCode", res.StatusCode))
		}
	}

	return nil
}

func (p *productRepo) InsertMany(products []*common.CreateProductCopyRequest) error {

	var (
		buf bytes.Buffer
	)

	if len(products) == 0 {
		return errors.Wrap(errors.New("no products were provided"), "no product")
	}

	if !exists(p.db, config.ElasticProductIndex) {
		res, err := p.db.Indices.Create(config.ElasticProductIndex)
		if err != nil {
			return errors.Wrap(err, "error while create index")
		}

		if res.IsError() {
			data, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			p.log.Error("errror while insert many products", logger.Any("res", string(data)))
			return errors.New("error while insert many" + string(data))
		}
	}

	for _, product := range products {

		var (
			body                  bytes.Buffer
			shopMeasurementValues = make(map[string]*catalog_service.ShopMeasurementValue)
			shopPrices            = make(map[string]*catalog_service.ShopPrice)
		)

		for _, measurementValue := range product.ShopMeasurementValues {
			shopMeasurementValues[measurementValue.ShopId] = &catalog_service.ShopMeasurementValue{
				ShopId:      measurementValue.ShopId,
				Amount:      float32(measurementValue.InStock),
				SmallLeft:   0,
				HasTrigger:  false,
				IsAvailable: measurementValue.IsAvailable,
			}

			shopPrices[measurementValue.ShopId] = &catalog_service.ShopPrice{
				ShopId:         measurementValue.ShopId,
				SupplyPrice:    float32(measurementValue.SupplyPrice),
				RetailPrice:    float32(measurementValue.RetailPrice),
				WholeSalePrice: float32(measurementValue.WholeSalePrice),
				MinPrice:       float32(measurementValue.MinPrice),
				MaxPrice:       float32(measurementValue.MaxPrice),
			}

		}

		err := config.JSONPBMarshaler.Marshal(&body, &catalog_service.UpsertProductES{
			Doc: &catalog_service.ProductES{
				Id:                product.Id,
				Sku:               product.Sku,
				Name:              product.Name,
				Image:             product.Image,
				IsMarking:         product.IsMarking,
				MxikCode:          product.MxikCode,
				ParentId:          product.ParentId,
				CompanyId:         product.Request.CompanyId,
				Description:       product.Description,
				ProductTypeId:     product.ProductTypeId,
				Barcodes:          product.Barcode,
				ShopPrices:        shopPrices,
				CreatedAt:         time.Now().Format(config.DateTimeFormat),
				UpdatedAt:         float64(time.Now().UnixMilli()),
				MeasurementValues: shopMeasurementValues,
				Supplier: &catalog_service.ShortSupplier{
					Id: product.SupplierId,
				},
				Vat: &catalog_service.ShortVat{
					Id: product.VatId,
				},
				MeasurementUnit: &catalog_service.ShortMeasurementUnit{
					Id: product.MeasurementUnitId,
				},
			},
			DocAsUpsert: true,
		})
		if err != nil {
			return errors.Wrap(err, "error while marshaling, jsonpb")
		}

		meta := []byte(fmt.Sprintf(`{ "update": { "_index": "%s", "_id" : "%s", "retry_on_conflict": 3 } }%s`, config.ElasticProductIndex, product.Id, "\n"))

		body.Grow(len("\n"))
		body.Write(bytes.NewBufferString("\n").Bytes())

		buf.Grow(len(meta) + len(body.Bytes()))
		buf.Write(meta)
		buf.Write(body.Bytes())
	}

	res, err := p.db.Bulk(
		bytes.NewReader(buf.Bytes()),
		p.db.Bulk.WithIndex(config.ElasticProductIndex),
		p.db.Bulk.WithRefresh("wait_for"),
		p.db.Bulk.WithPretty(),
	)
	if err != nil {
		return errors.Wrap(err, "Failed to bulk insert products")
	}
	defer res.Body.Close()

	if res.IsError() {
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		p.log.Error("errror while create many ", logger.Any("res", string(data)))
		return errors.New("error while create product on elastic")
	}

	return nil
}

func (p *productRepo) GetAllForExcel(req *catalog_service.GetAllProductsRequest) (*models.GetAllForExcelResponse, error) {

	var (
		res = models.GetAllForExcelResponse{
			Data:  make([]map[string]interface{}, 0),
			Total: 0,
		}
		r    map[string]interface{}
		buf  bytes.Buffer
		size = int(req.Limit)
		from = int((req.Page - 1) * req.Limit)
	)

	searchReq, err := makeGettAllSearchRequest(req, false)
	if err != nil {
		return nil, errors.Wrap(err, "error while getting searching products")
	}

	if err := json.NewEncoder(&buf).Encode(searchReq); err != nil {
		return nil, errors.Wrap(err, "error while encode")
	}

	response, err := p.db.Search(
		p.db.Search.WithContext(context.Background()),
		p.db.Search.WithIndex(config.ElasticProductIndex),
		p.db.Search.WithBody(&buf),
		p.db.Search.WithTrackTotalHits(true),
		p.db.Search.WithFrom(from),
		p.db.Search.WithSize(size),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error while get documents on elastic")
	}

	if response.IsError() {
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, errors.Wrap(err, "error while reading data")
		}

		p.log.Error("errror while get all products ", logger.Any("res", string(data)))
		return nil, errors.New("error while get  products on elastic " + string(data))
	}

	err = json.NewDecoder(response.Body).Decode(&r)
	if err != nil {
		return nil, errors.Wrap(err, "error while json.decode elastic res.Body")
	}

	for _, source := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {

		var data = make(map[string]interface{})

		product := catalog_service.ProductES{}

		jsonString, _ := json.Marshal(source.(map[string]interface{})["_source"])

		err = json.Unmarshal(jsonString, &product)
		if err != nil {
			return nil, errors.Wrap(err, "error while json.Unmarshal jsonString &product")
		}

		if product.Image != "" {
			product.Image = fmt.Sprintf("https://%s/%s/%s", p.cfg.MinioEndpoint, config.FileBucketName, product.Image)
		}

		data["product_id"] = product.Id
		data["name"] = product.Name
		data["sku"] = product.Sku
		data["mxik_code"] = product.MxikCode
		data["barcode"] = strings.Join(product.Barcodes, ", ")
		data["category"] = ""

		for _, category := range product.Categories {
			data["category"] = fmt.Sprintf("%s%s, ", data["category"], category.Name)
		}

		for _, shop := range product.ShopPrices {
			data[fmt.Sprintf("supply_price(%s)", shop.ShopName)] = shop.SupplyPrice
			data[fmt.Sprintf("retail_price(%s)", shop.ShopName)] = shop.RetailPrice

		}

		for _, measurementValue := range product.MeasurementValues {
			data[fmt.Sprintf("amount(%s)", measurementValue.ShopName)] = measurementValue.Amount
			data[fmt.Sprintf("low_stock(%s)", measurementValue.ShopName)] = measurementValue.SmallLeft
		}

		res.Data = append(res.Data, data)
	}

	res.Total = int64(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))

	return &res, nil
}

func (p *productRepo) GetAllForCSV(req *catalog_service.GetAllProductsRequest) (*models.GetAllForCsvResponse, error) {

	var (
		res = models.GetAllForCsvResponse{
			Data:  make([]map[string]interface{}, 0),
			Total: 0,
		}
		r    map[string]interface{}
		buf  bytes.Buffer
		size = int(req.Limit)
		from = int((req.Page - 1) * req.Limit)
	)

	searchReq, err := makeGettAllSearchRequest(req, false)
	if err != nil {
		return nil, errors.Wrap(err, "error while getting searching products")
	}

	if err := json.NewEncoder(&buf).Encode(searchReq); err != nil {
		return nil, errors.Wrap(err, "error while encode")
	}

	response, err := p.db.Search(
		p.db.Search.WithContext(context.Background()),
		p.db.Search.WithIndex(config.ElasticProductIndex),
		p.db.Search.WithBody(&buf),
		p.db.Search.WithTrackTotalHits(true),
		p.db.Search.WithFrom(from),
		p.db.Search.WithSize(size),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error while get documents on elastic")
	}

	if response.IsError() {
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, errors.Wrap(err, "error while reading data")
		}

		p.log.Error("errror while get all products ", logger.Any("res", string(data)))
		return nil, errors.New("error while get  products on elastic " + string(data))
	}

	err = json.NewDecoder(response.Body).Decode(&r)
	if err != nil {
		return nil, errors.Wrap(err, "error while json.decode elastic res.Body")
	}

	for _, source := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {

		var data = make(map[string]interface{})

		product := catalog_service.ProductES{}

		jsonString, _ := json.Marshal(source.(map[string]interface{})["_source"])

		err = json.Unmarshal(jsonString, &product)
		if err != nil {
			return nil, errors.Wrap(err, "error while json.Unmarshal jsonString &product")
		}

		if product.Image != "" {
			product.Image = fmt.Sprintf("https://%s/%s/%s", p.cfg.MinioEndpoint, config.FileBucketName, product.Image)
		}

		data["product_id"] = product.Id
		data["name"] = product.Name
		data["sku"] = product.Sku
		data["mxik_code"] = product.MxikCode
		data["barcode"] = strings.Join(product.Barcodes, ", ")
		data["category"] = ""

		for _, category := range product.Categories {
			data["category"] = fmt.Sprintf("%s%s, ", data["category"], category.Name)
		}

		for _, shop := range product.ShopPrices {
			data[fmt.Sprintf("supply_price(%s)", shop.ShopName)] = shop.SupplyPrice
			data[fmt.Sprintf("retail_price(%s)", shop.ShopName)] = shop.RetailPrice

		}

		for _, measurementValue := range product.MeasurementValues {
			data[fmt.Sprintf("amount(%s)", measurementValue.ShopName)] = measurementValue.Amount
			data[fmt.Sprintf("low_stock(%s)", measurementValue.ShopName)] = measurementValue.SmallLeft
		}

		res.Data = append(res.Data, data)
	}

	res.Total = int64(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))

	return &res, nil
}

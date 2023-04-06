package elastic

import (
	"bytes"
	"context"
	"fmt"
	"genproto/catalog_service"
	"io"

	"github.com/clarketm/json"

	"github.com/Invan2/invan_catalog_service/config"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/pkg/errors"
)

func makeGetLabelProductsSearchRequest(req *catalog_service.GetProductLabelsRequest) H {

	must := make([]H, 0)

	must = append(must, H{
		"term": H{
			"company_id.keyword": H{
				"value": req.GetRequest().GetCompanyId(),
			},
		},
	})

	must = append(must, H{
		"terms": H{
			"id.keyword": req.GetProductIds(),
		},
	})

	query := H{
		"bool": H{
			"must": must,
		},
	}

	return H{
		"query": query,
	}
}

func (p *productRepo) GetForLabel(req *catalog_service.GetProductLabelsRequest) (*catalog_service.GetAllProductsResponse, error) {

	var (
		res = catalog_service.GetAllProductsResponse{
			Data:       make([]*catalog_service.ProductES, 0),
			Statistics: &catalog_service.Statistics{},
			Total:      0,
		}
		r    map[string]interface{}
		buf  bytes.Buffer
		size = int(10000)
	)

	searchReq := makeGetLabelProductsSearchRequest(req)
	if err := json.NewEncoder(&buf).Encode(searchReq); err != nil {
		return nil, err
	}

	response, err := p.db.Search(
		p.db.Search.WithContext(context.Background()),
		p.db.Search.WithIndex(config.ElasticProductIndex),
		p.db.Search.WithBody(&buf),
		p.db.Search.WithTrackTotalHits(true),
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
			Description:       product.Description,
			CreatedAt:         product.CreatedAt,
			ShopPrices:        product.ShopPrices,
			Categories:        product.Categories,
		})
	}

	res.Total = int64(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))

	return &res, nil
}

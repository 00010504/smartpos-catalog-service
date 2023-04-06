package elastic

import (
	"context"
	"fmt"
	"genproto/catalog_service"
	"io"
	"strconv"
	"strings"

	"github.com/Invan2/invan_catalog_service/config"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/clarketm/json"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/pkg/errors"
)

func (p *productRepo) BulkUpdateProduct(req *catalog_service.ProductBulkOperationRequest, productMap map[string]*catalog_service.ProductES) error {

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

			p.log.Error("errror while get  products for bulk update", logger.Any("res", string(data)))
			return errors.New("error while  products for bulk update" + string(data))
		}
	}

	smallLeft, err := strconv.ParseFloat(req.Value, 64)
	if err != nil {
		smallLeft = 0
	}

	query := H{
		"query": H{
			"terms": H{
				"id.keyword": req.ProductIds,
			},
		},
		"script": H{
			"source": fmt.Sprintf(`
				String key = '%s';
				double small_left = %.3f;
				
				if(key == 'name') {
					ctx._source.name = params.products[ctx._source.id][key];
					return;
				}
				
				if (key == 'measurement_value'){
					ctx._source.measurement_unit = params.products[ctx._source.id].measurement_unit;
					return;
				}

				if (key == 'category'){
					ctx._source.categories = params.products[ctx._source.id].categories;
					return;
				}

				if (key == 'low_stock'){
					for (shop in params.shop_ids) {
						 ctx._source.measurement_values[shop].small_left = small_left;
					}
					return;
				}

			`, req.ProductField, smallLeft), // in this case req.Value is small_left
			"lang": "painless",
			"params": H{
				"products": productMap,
				"shop_ids": req.ShopIds,
			},
		},
	}

	body, err := json.Marshal(query)
	if err != nil {
		return err
	}

	refresh := true
	request := esapi.UpdateByQueryRequest{
		Index:   []string{config.ElasticProductIndex},
		Body:    strings.NewReader(string(body)),
		Refresh: &refresh,
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

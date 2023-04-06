package elastic

import (
	"errors"
	"genproto/common"
	"strings"
)

type filterFunction func(filter *common.FilterField) (H, error)

var (
	ErrFilterNotFound = errors.New("filter not found")

	filterFunctionMap = map[string]filterFunction{
		"category":         category,
		"measurement_unit": measurementUnit,
		"product_ids":      productIds,
	}
)

func category(filter *common.FilterField) (H, error) {
	return H{
		"bool": H{
			"should": []H{
				{
					"terms": H{
						"categories.parent_id": strings.Split(filter.Value, ","),
					},
				},
				{
					"terms": H{
						"categories.id": strings.Split(filter.Value, ","),
					},
				},
			},
		},
	}, nil
}

func measurementUnit(filter *common.FilterField) (H, error) {
	return H{
		"terms": H{
			"measurement_unit.id.keyword": strings.Split(filter.Value, ","),
		},
	}, nil
}

func productIds(filter *common.FilterField) (H, error) {
	return H{
		"terms": H{
			"id.keyword": strings.Split(filter.Value, ","),
		},
	}, nil
}

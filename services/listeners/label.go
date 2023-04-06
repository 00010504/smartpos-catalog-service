package listeners

import (
	"context"

	"genproto/catalog_service"
	"genproto/common"

	"github.com/pkg/errors"
)

func (c *catalogService) CreateLabel(ctx context.Context, req *catalog_service.CreateLabelRequest) (*common.ResponseID, error) {

	tr, err := c.strg.WithTransaction()
	if err != nil {
		return nil, errors.Wrap(err, "error while begin starting")
	}

	defer func() {
		if err == nil {
			tr.Commit()
		} else {
			tr.Rollback()
		}
	}()

	res, err := tr.Label().Create(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *catalogService) GetLabelById(ctx context.Context, req *common.RequestID) (*catalog_service.GetLabelResponse, error) {
	return c.strg.Label().GetById(req)
}

func (c *catalogService) UpdateLabelById(ctx context.Context, req *catalog_service.UpdateLabelRequest) (*common.ResponseID, error) {

	tr, err := c.strg.WithTransaction()
	if err != nil {
		return nil, errors.Wrap(err, "error while begin starting")
	}

	defer func() {
		if err == nil {
			tr.Commit()
		} else {
			tr.Rollback()
		}
	}()

	res, err := tr.Label().UpdateById(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *catalogService) GetAllLabels(ctx context.Context, req *common.SearchRequest) (*catalog_service.GetAllLabelsResponse, error) {
	return c.strg.Label().GetAll(req)
}

func (c *catalogService) GetProductFields(ctx context.Context, req *catalog_service.GetProductFieldsRequest) (*catalog_service.GetProductFieldsResponse, error) {

	var res = catalog_service.GetProductFieldsResponse{
		Fields: []*catalog_service.GetProductFieldResponse{},
	}

	if req.RequestType == "label" {
		res.Fields = append(res.Fields,
			&catalog_service.GetProductFieldResponse{Name: "retail_price"},
			&catalog_service.GetProductFieldResponse{Name: "name"},
			&catalog_service.GetProductFieldResponse{Name: "barcode"},
			&catalog_service.GetProductFieldResponse{Name: "sku"},
			&catalog_service.GetProductFieldResponse{Name: "mxik_code"},
			&catalog_service.GetProductFieldResponse{Name: "date"},
			&catalog_service.GetProductFieldResponse{Name: "currency"},
		)
	} else if req.RequestType == "export_excel" {
		res.Fields = append(res.Fields,
			&catalog_service.GetProductFieldResponse{Name: "product_id"},
			&catalog_service.GetProductFieldResponse{Name: "name"},
			&catalog_service.GetProductFieldResponse{Name: "sku"},
			&catalog_service.GetProductFieldResponse{Name: "mxik_code"},
			&catalog_service.GetProductFieldResponse{Name: "barcode"},
			&catalog_service.GetProductFieldResponse{Name: "category"},
			&catalog_service.GetProductFieldResponse{Name: "supply_price"},
			&catalog_service.GetProductFieldResponse{Name: "retail_price"},
			&catalog_service.GetProductFieldResponse{Name: "amount"},
			&catalog_service.GetProductFieldResponse{Name: "low_stock"},
		)
	}

	customFields, err := c.strg.Product().GetProductCustomFields(req.Request)
	if err != nil {
		return nil, err
	}

	for _, cuscustomField := range customFields {
		res.Fields = append(res.Fields, &catalog_service.GetProductFieldResponse{
			Name:          cuscustomField.Name,
			IsCustomField: true,
		})
	}

	return &res, nil
}

func (c *catalogService) DeleteLabelById(ctx context.Context, req *common.RequestID) (*common.ResponseID, error) {
	return c.strg.Label().DeleteLabelById(req)
}

func (c *catalogService) DeleteLabelsByIds(ctx context.Context, req *common.RequestIDs) (*common.Empty, error) {
	return c.strg.Label().DeleteLabelsByIds(req)
}
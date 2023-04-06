package listeners

import (
	"context"
	"genproto/catalog_service"
	"genproto/common"

	"github.com/pkg/errors"
)

func (c *catalogService) CreateVat(ctx context.Context, req *catalog_service.CreateVatRequest) (*common.ResponseID, error) {

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

	res, err := tr.Vat().Create(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *catalogService) GetVatById(ctx context.Context, req *common.RequestID) (*catalog_service.GetVatByIdResponse, error) {
	return c.strg.Vat().GetById(ctx, req)
}

func (c *catalogService) UpdateVatById(ctx context.Context, req *catalog_service.UpdateVatRequest) (*common.ResponseID, error) {

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

	res, err := tr.Vat().Update(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *catalogService) GetAllVats(ctx context.Context, req *common.SearchRequest) (*catalog_service.GetAllVatsResponse, error) {
	return c.strg.Vat().GetAll(ctx, req)
}

func (c *catalogService) DeleteVat(ctx context.Context, req *common.RequestID) (*common.ResponseID, error) {
	return c.strg.Vat().Delete(req)
}

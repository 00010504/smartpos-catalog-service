package listeners

import (
	"context"
	"genproto/common"

	"genproto/catalog_service"
)

func (c *catalogService) CreateCategory(ctx context.Context, req *catalog_service.CreateCategoryRequest) (*common.ResponseID, error) {

	id, err := c.strg.Category().Create(req)
	if err != nil {
		return nil, err
	}

	return &common.ResponseID{Id: id}, nil
}
func (c *catalogService) GetCategoryByID(ctx context.Context, req *common.RequestID) (*catalog_service.GetCategoryByIDResponse, error) {

	category, err := c.strg.Category().GetByID(req)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (c *catalogService) UpdateCategory(ctx context.Context, req *catalog_service.UpdateCategoryRequest) (*common.ResponseID, error) {

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

	_, err = tr.Category().Update(req)
	if err != nil {
		return nil, err
	}

	return &common.ResponseID{Id: req.Id}, nil
}

func (c *catalogService) GetAllCategories(ctx context.Context, req *catalog_service.GetAllCategoriesRequest) (*catalog_service.GetAllCategoriesResponse, error) {

	res, err := c.strg.Category().GetAll(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *catalogService) DeleteCategoryById(ctx context.Context, req *common.RequestID) (*common.ResponseID, error) {

	_, err := c.strg.Category().Delete(req)
	if err != nil {
		return nil, err
	}

	return &common.ResponseID{Id: req.Id}, nil
}

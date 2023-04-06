package models

type Category struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	ParentId  string `json:"parent_id"`
	CreatedAt string `json:"created_at"`
	CreatedBy string `json:"created_by"`
}

type CreateCategoryRequest struct {
	Id       string `json:"id" swaggerignore:"true"`
	Name     string `json:"name"`
	ParentId string `json:"parent_id"`
}

type UpdateCategoryRequest struct {
	Id       string `json:"id" swaggerignore:"true"`
	Name     string `json:"name"`
	ParentId string `json:"parent_id"`
}

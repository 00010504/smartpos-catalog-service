package models

type GetShopNameRespone struct {
	Id   string
	Name string
}

type GetShopsReq struct {
	ShopIds   []string
	CompanyId string
}

type ProductExcelFields struct {
	Name        string
	ProductId   string
	Barcode     string
	Category    string
	SupplyPrice string
	RetailPrice string
	LowStock    string
	MxikCode    string
	SKU         string
	Margin      string
}

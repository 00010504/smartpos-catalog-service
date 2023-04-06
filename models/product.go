package models

import (
	"database/sql"
)

type ShopMeasurementValue struct {
	ShopID              string  `json:"shop_id"`
	IsAvailable         bool    `json:"is_available"`
	TotalActive         float64 `json:"total_active"`
	TotalInActive       float64 `json:"total_inactive"`
	Total               float64 `json:"total"`
	HasTrigger          bool    `json:"has_trigger"`
	Amount              float64 `json:"amount"`
	SmallLeft           float64 `json:"small_left"`
	TotalImported       float64 `json:"total_imported"`
	TotalSold           float64 `json:"total_sold"`
	TotalTrasfered      float64 `json:"total_transfered"`
	TotalTrasferArrived float64 `json:"total_transfer_arrived"`
	TotalSupplierOrder  float64 `json:"total_supplier_order"`
	TotalPostponeOrder  float64 `json:"total_postpone_order"`
}

type ShopPrice struct {
	ShopID           string  `json:"shop_id"`
	SupplyPrice      float64 `json:"supply_price"`
	RetailPrice      float64 `json:"retail_price"`
	WholeSalePrice   float64 `json:"whole_sale_price"`
	MinPrice         float64 `json:"min_price"`
	MaxPrice         float64 `json:"max_price"`
	TotalSupplyPrice float32 `json:"total_supply_price"`
	TotalRetailPrice float32 `json:"total_retail_price"`
}

type ProductNullBrand struct {
	Id   sql.NullString
	Name sql.NullString
}

type ProductNullSupplier struct {
	Id   sql.NullString
	Name sql.NullString
}

type VatNullSupplier struct {
	Id         sql.NullString
	Name       sql.NullString
	Percentage sql.NullString
}
type ProductNullMeasurementUnit struct {
	Id                   sql.NullString       `json:"id"`
	IsDeletable          sql.NullBool         `json:"is_deletable"`
	ShortName            sql.NullString       `json:"short_name"`
	LongName             sql.NullString       `json:"long_name"`
	ShortNameTranslation map[string]string    `json:"short_name_translation"`
	LongNameTranslation  map[string]string    `json:"long_name_translation"`
	Precision            ProductNullPrecision `json:"precision"`
}

type ProductNullPrecision struct {
	Id    sql.NullString
	Value sql.NullString
}

type GetProductCustomFieldResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type ProductShopPrice struct {
	ShopId         string
	ShopName       string
	RetailPrice    float64
	WholeSalePrice float64
	MinPrice       float64
	MaxPrice       float64
}

type GetProductForLabel struct {
	Id            string            //`json:"id"`
	Sku           string            //`json:"sku"`
	Name          string            //`json:"name"`
	Image         string            //`json:"image"`
	IsMarking     bool              //`json:"is_marking"`
	MxikCode      string            //`json:"mxik_code"`
	ParentId      string            //`json:"parent_id"`
	CompanyId     string            //`json:"company_id"`
	CreatedAt     string            //`json:"created_at"`
	Description   string            //`json:"description"`
	ProductTypeId string            //`json:"product_type_id"`
	Barcodes      []string          //`json:"barcodes"`
	ShopPrice     *ProductShopPrice //`json:"value"`
	// MeasurementUnit   *ShortMeasurementUnit            //`json:"measurement_unit"`
	// MeasurementValues map[string]*ShopMeasurementValue //`json:"value"`
}

type GetAllForExcelResponse struct {
	Data  []map[string]interface{}
	Total int64
}

type GetAllForCsvResponse struct {
	Data  []map[string]interface{}
	Total int64
}

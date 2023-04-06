package listeners

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"genproto/catalog_service"
	"genproto/common"
	"log"
	"os"

	"github.com/Invan2/invan_catalog_service/models"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	excelize "github.com/xuri/excelize/v2"
)

func (c *catalogService) makeProductExeclStyle(f *excelize.File, sheetName string) error {

	style, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{WrapText: true},
	})
	if err != nil {
		return errors.Wrap(err, "error while styling")
	}

	// to fill text in columns
	err = f.SetColStyle(sheetName, "A:AA", style)
	if err != nil {
		return errors.Wrap(err, "error while styling column")
	}

	err = f.SetColWidth(sheetName, "A", "AA", 30)
	if err != nil {
		return errors.Wrap(err, "error while styling column width")
	}

	return nil
}

type WriteExcelRowRequest struct {
	File              *excelize.File
	SheetName         string
	ProductsFilterReq *catalog_service.GetAllProductsRequest
	ExeclHeaders      []string
}

type WriteCSVRowRequest struct {
	File              *csv.Writer
	ProductsFilterReq *catalog_service.GetAllProductsRequest
	CSVHeader         []string
}

func (c *catalogService) writeExcelRows(req WriteExcelRowRequest) error {
	productsMap, err := c.elastic.Product().GetAllForExcel(req.ProductsFilterReq)
	if err != nil {
		return errors.Wrap(err, "error while getting products for excel")
	}

	for i, item := range productsMap.Data {
		var row = make([]interface{}, 0)

		for _, key := range req.ExeclHeaders {
			row = append(row, item[key])
		}

		startCell, err := excelize.JoinCellName("A", i+2)
		if err != nil {
			return errors.Wrap(err, "error while startCell")
		}
		if err := req.File.SetSheetRow(req.SheetName, startCell, &row); err != nil {
			return errors.Wrap(err, "error while setSheetRow")
		}
	}
	return nil
}

func (c *catalogService) writeCSVRows(req WriteCSVRowRequest) error {
	productsMap, err := c.elastic.Product().GetAllForCSV(req.ProductsFilterReq)
	if err != nil {
		return errors.Wrap(err, "error while getting products for excel")
	}

	for _, item := range productsMap.Data {
		var row = make([]string, 0)

		for _, key := range req.CSVHeader {
			val := item[key]
			if val == nil {
				row = append(row, "0")
			} else {
				row = append(row, fmt.Sprintf("%v", val))
			}
		}

		if err := req.File.Write(row); err != nil {
			return errors.Wrap(err, "error while Writing")
		}

	}
	return nil
}

func (c *catalogService) CreateProductExelTemplate(ctx context.Context, req *catalog_service.GetProductExcelDownloadRequest) (*common.ResponseID, error) {
	var (
		excelFileName = "Sheet1"
		sheetName     = "Product"
		res           common.ResponseID
		bucketName    string = "file"

		excelMainHeaders = []string{
			"product_id",
			"name",
			"sku",
			"mxik_code",
			"barcode",
			"category",
		}
		excelShopHeaders = []string{
			"supply_price",
			"retail_price",
			"amount",
			"low_stock",
		}
		excelHeader = []string{}
	)

	shops, err := c.strg.Shop().GetAll(&models.GetShopsReq{
		CompanyId: req.Request.CompanyId,
		ShopIds:   req.ShopIds,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error while getting shops")
	}

	for _, excelMainHeader := range excelMainHeaders {
		_, ok := req.ProductFields[excelMainHeader]
		if ok {
			excelHeader = append(excelHeader, excelMainHeader)
		}
	}

	for _, shop := range shops {
		for _, excelShopHeader := range excelShopHeaders {
			_, ok := req.ProductFields[excelShopHeader]
			if ok {
				excelHeader = append(excelHeader, fmt.Sprintf("%s(%s)", excelShopHeader, shop.Name))
			}
		}
	}

	f := excelize.NewFile()
	f.SetSheetName(excelFileName, sheetName)

	if err := c.makeProductExeclStyle(f, sheetName); err != nil {
		return nil, err
	}

	startCell, err := excelize.JoinCellName("A", 1)
	if err != nil {
		return nil, err
	}
	if err := f.SetSheetRow(sheetName, startCell, &excelHeader); err != nil {
		return nil, err
	}

	if req.ExportType == "data" {
		err = c.writeExcelRows(WriteExcelRowRequest{
			File:         f,
			SheetName:    sheetName,
			ExeclHeaders: excelHeader,
			ProductsFilterReq: &catalog_service.GetAllProductsRequest{
				Limit:      10000,
				Page:       1,
				Search:     "",
				ShopIds:    req.ShopIds,
				Request:    req.Request,
				Filters:    req.Filters,
				ProductIds: req.ProductIds,
				Statistics: false,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	if req.ExportType == "all" {
		err = c.writeExcelRows(WriteExcelRowRequest{
			File:         f,
			SheetName:    sheetName,
			ExeclHeaders: excelHeader,
			ProductsFilterReq: &catalog_service.GetAllProductsRequest{
				Limit:      10000,
				Page:       1,
				Search:     "",
				ShopIds:    req.ShopIds,
				Request:    req.Request,
				Filters:    req.Filters,
				Statistics: false,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	if err := f.SaveAs("Products.xlsx"); err != nil {
		return nil, err
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	fileName := uuid.NewString()

	_, err = c.minio.PutObject(context.Background(), bucketName, fileName, buf, -1, minio.PutObjectOptions{ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"})
	if err != nil {
		return nil, errors.Wrap(err, "error while upload file to minio")
	}

	res.Id = fmt.Sprintf("https://%s/%s/%s", c.cfg.MinioEndpoint, bucketName, fileName)

	return &res, nil
}

func (c *catalogService) CreateProductCsvTemplate(ctx context.Context, req *catalog_service.GetProductCsvDownloadRequest) (*common.ResponseID, error) {
	var (
		res        common.ResponseID
		bucketName string = "file"

		csvMainHeaders = []string{
			"product_id",
			"name",
			"sku",
			"mxik_code",
			"barcode",
			"category",
		}
		csvShopHeaders = []string{
			"retail_price",
			"supply_price",
			"low_stock",
			"amount",
		}
		csvHeader = []string{}
	)

	shops, err := c.strg.Shop().GetAll(&models.GetShopsReq{
		CompanyId: req.Request.CompanyId,
		ShopIds:   req.ShopIds,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error while getting shops")
	}

	for _, csvMainHeader := range csvMainHeaders {
		_, ok := req.ProductFields[csvMainHeader]
		if ok {
			csvHeader = append(csvHeader, csvMainHeader)
		}
	}

	for _, shop := range shops {
		for _, csvShopHeader := range csvShopHeaders {
			_, ok := req.ProductFields[csvShopHeader]
			if ok {
				csvHeader = append(csvHeader, fmt.Sprintf("%s(%s)", csvShopHeader, shop.Name))
			}
		}
	}

	f, err := os.Create("products.csv")
	if err != nil {
		log.Fatalln("failed to open file", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)

	if err := w.Write(csvHeader); err != nil {
		return nil, errors.Wrap(err, "error while Writing")
	}

	if req.ExportType == "data" {
		err = c.writeCSVRows(WriteCSVRowRequest{
			File:      w,
			CSVHeader: csvHeader,
			ProductsFilterReq: &catalog_service.GetAllProductsRequest{
				Limit:      10000,
				Page:       1,
				Search:     "",
				ShopIds:    req.ShopIds,
				Request:    req.Request,
				Filters:    req.Filters,
				ProductIds: req.ProductIds,
				Statistics: false,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	if req.ExportType == "all" {
		err = c.writeCSVRows(WriteCSVRowRequest{
			File:      w,
			CSVHeader: csvHeader,
			ProductsFilterReq: &catalog_service.GetAllProductsRequest{
				Limit:      10000,
				Page:       1,
				Search:     "",
				ShopIds:    req.ShopIds,
				Filters:    req.Filters,
				Request:    req.Request,
				Statistics: false,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		return nil, errors.Wrap(err, "error while Writing. w.Error()")
	}

	fileName := uuid.NewString()

	fileData, err := os.ReadFile("./products.csv")

	if err != nil {
		return nil, err
	}

	_, err = c.minio.PutObject(context.Background(), bucketName, fileName, bytes.NewReader(fileData), -1, minio.PutObjectOptions{ContentType: "text/csv"})
	if err != nil {
		return nil, errors.Wrap(err, "error while upload file to minio")
	}

	res.Id = fmt.Sprintf("https://%s/%s/%s", c.cfg.MinioEndpoint, bucketName, fileName)

	return &res, nil
}

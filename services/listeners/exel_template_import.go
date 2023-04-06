package listeners

import (
	"context"
	"fmt"
	"genproto/common"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	excelize "github.com/xuri/excelize/v2"
)

func (c *catalogService) makeExcelTemplateStyle(f *excelize.File, sheetName string) error {

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

	err = f.SetColWidth(sheetName, "A", "AA", 15)
	if err != nil {
		return errors.Wrap(err, "error while styling column width")
	}

	return nil
}

func (c *catalogService) CreateExelTemplate(ctx context.Context, req *common.Request) (*common.ResponseID, error) {

	var (
		res        common.ResponseID
		bucketName string = "file"
	)

	f := excelize.NewFile()
	sheetName := "Product"
	f.SetSheetName("Sheet1", sheetName)

	if err := c.makeProductExeclStyle(f, sheetName); err != nil {
		return nil, err
	}

	data := [][]interface{}{
		{"VARIATION_ID", "NAME", "SKU", "BARCODE", "QUANTITY", "SUPPLY_PRICE (UZS)", "RETAIL_PRICE (UZS)", "CATEGORY_NAME", "BRAND_NAME", "MEASUREMENT_UNIT", "SUPPLIER", "MIN_PRICE (UZS)", "MAX_PRICE (UZS)", "WHOLESALE_PRICE(UZS)"},
	}

	for i, row := range data {
		startCell, err := excelize.JoinCellName("A", i+1)
		if err != nil {
			return nil, err
		}

		if err := f.SetSheetRow(sheetName, startCell, &row); err != nil {
			return nil, err
		}
	}

	style1, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: string("FF0000")},
	})

	style2, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal:      "center",
			Indent:          1,
			JustifyLastLine: false,
			ReadingOrder:    0,
			RelativeIndent:  1,
			ShrinkToFit:     true,
			Vertical:        "",
		},
	})
	if err != nil {
		return nil, err
	}

	if err := f.SetCellStyle(sheetName, "A1", "A1", style2); err != nil {
		return nil, err
	}

	if err := f.SetCellStyle(sheetName, "F1", "G1", style1); err != nil {
		return nil, err
	}

	if err := f.SetCellStyle(sheetName, "B1", "C1", style1); err != nil {
		return nil, err
	}

	if err := f.SetCellStyle(sheetName, "D1", "E1", style1); err != nil {
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

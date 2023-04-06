package listeners

import (
	"context"
	"fmt"
	"genproto/catalog_service"
	"genproto/common"
	"os"
	"strconv"
	"time"

	"github.com/Invan2/invan_catalog_service/config"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
)

func (c *catalogService) BulkGenerateProductLabels(ctx context.Context, req *catalog_service.GetProductLabelsRequest) (*common.ResponseID, error) {

	var (
		res         common.ResponseID
		productsMap = make([]map[string]interface{}, 0)
	)

	label, err := c.strg.Label().GetById(&common.RequestID{Id: req.LabelId, Request: req.Request})
	if err != nil {
		return nil, err
	}

	products, err := c.elastic.Product().GetForLabel(req)
	if err != nil {
		return nil, err
	}

	for _, product := range products.Data {

		r := map[string]interface{}{
			"id":           product.Id,
			"name":         product.Name,
			"barcode":      product.Barcodes,
			"mxik_code":    product.MxikCode,
			"date":         time.Now().Format(config.DateFormat),
			"retail_price": strconv.FormatFloat(float64(product.ShopPrices[req.ShopId].RetailPrice), 'E', -1, 64),
		}

		if len(product.Barcodes) > 0 {
			r["barcode"] = product.Barcodes[0]
		}

		productsMap = append(productsMap, r)
	}

	res.Id, err = c.pdf.MakeProductsLabel(productsMap, label)
	if err != nil {
		return nil, err
	}

	res.Id, err = c.uploadLabelToMinio(res.Id)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *catalogService) uploadLabelToMinio(str string) (string, error) {
	var (
		bucketName  = "file"
		contentType = "text/html"
	)

	buffer, err := os.Open("./" + str)
	if err != nil {
		return "", errors.Wrap(err, "error while os.Open html file")
	}

	defer buffer.Close()

	buffer.Sync()

	fStat, err := buffer.Stat()
	if err != nil {
		return "", errors.Wrap(err, "error while get file statistics")
	}

	_, err = c.minio.PutObject(context.Background(), bucketName, str, buffer, fStat.Size(), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", errors.Wrap(err, "error while upload file to minio")
	}

	err = os.Remove("./" + str)
	if err != nil {
		c.log.Error("Error while remove file")
	}

	return fmt.Sprintf("https://%s/%s/%s", c.cfg.MinioEndpoint, bucketName, str), nil
}

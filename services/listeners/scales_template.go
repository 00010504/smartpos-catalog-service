package listeners

import (
	"context"
	"fmt"
	"genproto/catalog_service"
	"genproto/common"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Invan2/invan_catalog_service/config"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
)

func (c *catalogService) CreateScalesTemplates(ctx context.Context, req *catalog_service.CreateScalesTemplateRequest) (res *common.ResponseID, err error) {

	c.log.Info("CreateScalesTemplates", logger.Any("request", req))

	res, err = c.strg.ScalesTemplate().CreateScalesTemplates(req)
	if err != nil {
		return nil, err
	}

	return
}
func (c *catalogService) GetScalesTemplateByID(ctx context.Context, req *catalog_service.GetScalesTemplateByIDRequest) (res *catalog_service.ScalesTemplate, err error) {

	c.log.Info("GetScalesTemplateByID", logger.Any("request", req))

	res, err = c.strg.ScalesTemplate().GetScalesTemplateByID(req)
	if err != nil {
		return nil, err
	}

	// Get products
	products, err := c.elastic.Product().GetAll(&catalog_service.GetAllProductsRequest{
		Limit:          10000,
		Page:           1,
		Request:        req.GetRequest(),
		ShopIds:        []string{req.GetShopId()},
		MeasurementIds: res.GetMeasurementUnitId(),
	})
	if err != nil {
		c.log.Error("c.elastic.Product().GetAll for GetScalesTemplateByID", logger.Any("request", err))
		return nil, err
	}
	fmt.Println("len of products --- ", len(products.GetData()))

	var (
		line string
		text = ``
	)
	for _, v := range products.GetData() {
		line = strings.ReplaceAll(res.GetValues(), "{sku}", v.GetSku())
		line = strings.ReplaceAll(line, "{name}", v.GetName())
		if res.GetName() == "Mettler toledo Spct 1" {
			line = strings.ReplaceAll(line, "{price}", strconv.Itoa(int(v.GetShopPrices()[req.GetShopId()].GetRetailPrice())/100))
		} else {
			line = strings.ReplaceAll(line, "{price}", strconv.Itoa(int(v.GetShopPrices()[req.GetShopId()].GetRetailPrice())))
		}
		text += line + `
`
	}

	res.Url, err = c.UploadTemplateToMinio(text)
	if err != nil {
		return nil, err
	}

	return
}
func (c *catalogService) GetAllScalesTemplates(ctx context.Context, req *catalog_service.GetAllScalesTemplatesRequest) (res *catalog_service.GetAllScalesTemplatesResponse, err error) {

	c.log.Info("GetAllScalesTemplates", logger.Any("request", req))

	res, err = c.strg.ScalesTemplate().GetAllScalesTemplates(req)
	if err != nil {
		return nil, err
	}

	return
}

func (c *catalogService) UploadTemplateToMinio(str string) (string, error) {
	var (
		bucketName  = "file"
		contentType = "text/plain"
	)

	fileName := "scales_template_" + time.Now().Format(config.DateTimeFormatWithoutSpaces) + ".txt"
	file, err := os.Create("./" + fileName)
	if err != nil {
		return "", errors.Wrap(err, "error while create file")
	}

	defer file.Close()

	_, err = file.WriteString(str)
	if err != nil {
		return "", errors.Wrap(err, "error while write str to file")
	}

	buffer, err := os.Open("./" + fileName)
	if err != nil {
		return "", errors.Wrap(err, "error while os.Open txt file")
	}

	defer buffer.Close()

	buffer.Sync()

	fStat, err := buffer.Stat()
	if err != nil {
		return "", errors.Wrap(err, "error while get file statistics")
	}

	_, err = c.minio.PutObject(context.Background(), bucketName, fileName, buffer, fStat.Size(), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", errors.Wrap(err, "error while upload file to minio")
	}

	err = os.Remove("./" + fileName)
	if err != nil {
		c.log.Error("Error while remove file")
	}

	return fmt.Sprintf("https://%s/%s/%s", c.cfg.MinioEndpoint, bucketName, fileName), nil
}

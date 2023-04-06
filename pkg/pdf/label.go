package pdfmaker

import (
	"fmt"
	"genproto/catalog_service"
	"os"
	"time"

	"github.com/Invan2/invan_catalog_service/config"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var qrcodeScript string

type getCard map[string]func(*catalog_service.LabelContent, map[string]interface{}) (string, error)

func (p *PdFmaker) getCardDivByType() getCard {

	var card = make(getCard)

	card["text"] = func(content *catalog_service.LabelContent, product map[string]interface{}) (string, error) {

		if content.FieldName == "date" {
			return fmt.Sprintf(`
			<div class="content"
				style="
          			bottom: %dpx;
         			left: %dpx;
         			 font-weight: %d;
					font-size: %dpx;
				">
				%s
			</div>
		`,
				content.Bottom,
				content.Position.X,
				content.Format.FontWeight,
				content.Format.FontSize,
				time.Now().Format(config.DateTimeFormat),
			), nil
		}

		value, ok := product[content.FieldName]
		if !ok {

			p.log.Info("product", logger.Any("data", product), logger.Any("content", content), logger.Any("key", content.FieldName))
			return "", errors.New("product field invalid 1")
		}

		return fmt.Sprintf(`
			<div class="content"
				style="
					top: %dpx;
					left: %dpx;
					width: %dpx;
          			font-weight: %d;
					font-size: %dpx;
				">
				%s
			</div>
		`,
			content.Position.Y,
			content.Position.X,
			content.Width,
			content.Format.FontWeight,
			content.Format.FontSize,
			value,
		), nil
	}

	card["image"] = func(content *catalog_service.LabelContent, product map[string]interface{}) (string, error) {

		var src string

		if content.ProductImage != "" {
			value, ok := product[content.FieldName]
			if !ok {
				return "", errors.New("Product field invalid 2")
			}

			src = fmt.Sprintf("https://%s/%s/%s", p.cfg.MinioEndpoint, "file", value)
		} else {
			src = fmt.Sprintf("https://%s/%s/%s", p.cfg.MinioEndpoint, "file", content.FieldName)
		}

		return fmt.Sprintf(`
			<div
				class="content"
				style="
					width: %dpx;
					height: %dpx;
					top: %dpx;
					left: %dpx;
          			bottom: %dpx;
			">
				<img src="%s" width="100%%" height="100%%" />
			</div>
			`,
			content.Width,
			content.Height,
			content.Position.Y,
			content.Position.X,
			content.Bottom,
			src,
		), nil
	}

	card["barcode"] = func(content *catalog_service.LabelContent, product map[string]interface{}) (string, error) {

		value, ok := product[content.FieldName]
		if !ok {
			return "", errors.New("Product field invalid 3")
		}

		fmt.Printf("%+v\n", content.Id)

		qrcodeScript += fmt.Sprintf(
			`
				new QRCode(document.getElementById('%s'), {
				text: '%s',
				width: %d,
				height:  %d,
				colorDark: '#000',
				colorLight: '#fff',
				correctLevel: QRCode.CorrectLevel.H
				});
			`,
			content.Id,
			value,
			content.Width,
			content.Height,
		)

		return fmt.Sprintf(`
			<div
				class="qrcode"
				id="%s"
				style="
					width: %dpx;
					height: %dpx;
					bottom: %dpx;
					right: %dpx;
				"></div>
			`,
			content.Id,
			content.Width,
			content.Height,
			content.Bottom,
			content.Right,
		), nil
	}

	return card
}

func (p *PdFmaker) MakeProductsLabel(products []map[string]interface{}, label *catalog_service.GetLabelResponse) (string, error) {

	html := `
  <!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta http-equiv="X-UA-Compatible" content="IE=edge" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>A4</title>
    <link
      href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap"
      rel="stylesheet"
    />
    <style>
      * {
        margin: 0;
        padding: 0;
        box-sizing: border-box;
        font-family: "Inter", sans-serif;
      }
      .wrapper {
        width: 100%;
        height: 100%;
        display: flex;
        align-items: center;
        flex-wrap: wrap;
        font-weight: sans-serif;
      }
      .label {
        border: 1px solid #000000;
        position: relative;
      }
      .label * {
        position: absolute;
      }
      .label * {
        overflow: hidden;
        text-overflow: ellipsis;
        display: -webkit-box;
        -webkit-box-orient: vertical;
        -webkit-line-clamp: 2; /* limit to 2 lines */
        letter-spacing: 0.06em;
      }
    </style>
	<script src="https://cdn.jsdelivr.net/gh/davidshimjs/qrcodejs/qrcode.min.js"></script>
  </head>
  <body>
	`

	// container,
	container := `<div class="wrapper" style="width: 210mm">`

	for _, product := range products {

		card := fmt.Sprintf(
			`
      <div class="label" style="width: %dmm; height: %dmm">
			`,
			label.Parameters.Width,
			label.Parameters.Height,
		)

		chDiv := make(chan string)
		chError := make(chan error)

		for _, content := range label.Content {
			fmt.Printf("%+v\n", content)
			go func(content *catalog_service.LabelContent) {

				div, err := p.getCardDivByType()[content.Type](content, product)
				if err != nil {
					chError <- err
					chDiv <- ""
					return
				}

				chError <- nil
				chDiv <- div
			}(content)
		}

		for i := 0; i < len(label.Content); i++ {
			err := <-chError
			if err != nil {
				return "", err
			}

			card += <-chDiv
		}

		container += card + " </div>"
	}  

	html += fmt.Sprintf("%s%s%s", container+"</div> <script>", qrcodeScript, "</script></body></html>")

	outputPath := uuid.NewString() + ".html"

	f, err := os.Create(outputPath)
	if err != nil {
		return "", errors.Wrap(err, "error while create html file")
	}
	defer f.Close()

	_, err = f.WriteString(html)
	if err != nil {
		return "", errors.Wrap(err, "error while write html")
	}

	return outputPath, nil
}

func (p *PdFmaker) MakeProductsPriceTag(products []map[string]interface{}, label *catalog_service.GetLabelResponse) (string, error) {

	html := `
		<!DOCTYPE html>
		<html lang="en">

		<head>
			<meta charset="UTF-8">
			<meta http-equiv="X-UA-Compatible" content="IE=edge">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Label</title>
			<style>
				.content {
					position: absolute;
					word-break: break-all;
				}

				.container {
					display: flex;
					flex-wrap: wrap;
					gap: 20px;
				}

				.qrcode {
					position: absolute;
					justify-content: center;
				}
			</style>
			<script src="https://cdn.jsdelivr.net/gh/davidshimjs/qrcodejs/qrcode.min.js"></script>
		</head>

		<body>
	`

	// container,
	container := fmt.Sprintf(`<div class="%s">`, "container")

	for _, product := range products {

		card := fmt.Sprintf(
			`<div
				style="
					background-color: #fff;
					box-shadow: 0px 0px 20px rgba(0, 0, 0, 0.05);
					border-radius: 10px;
					position: relative;
					color: #232323;
					width: %dpx;
					height: %dpx;
					position: relative;
					">
			`,
			label.Parameters.Width,
			label.Parameters.Height,
		)

		chDiv := make(chan string)
		chError := make(chan error)

		for _, content := range label.Content {

			go func(content *catalog_service.LabelContent) {

				div, err := p.getCardDivByType()[content.Type](content, product)
				if err != nil {
					chError <- err
					chDiv <- ""
					return
				}

				chError <- nil
				chDiv <- div
			}(content)
			// card += div
		}

		for i := 0; i < len(label.Content); i++ {
			err := <-chError
			if err != nil {
				return "", err
			}

			card += <-chDiv
		}

		container += card + " </div>"
	}

	html += fmt.Sprintf("%s%s%s", container+"</div> <script>", qrcodeScript, "</script></body></html>")

	outputPath := uuid.NewString() + ".html"

	f, err := os.Create(outputPath)
	if err != nil {
		return "", errors.Wrap(err, "error while create html file")
	}
	defer f.Close()

	_, err = f.WriteString(html)
	if err != nil {
		return "", errors.Wrap(err, "error while write html")
	}

	return outputPath, nil
}

func (p *PdFmaker) MakeProductsPriceReceipt(products []map[string]string, label *catalog_service.GetLabelResponse) (string, error) {

	var res string

	return res, nil
}

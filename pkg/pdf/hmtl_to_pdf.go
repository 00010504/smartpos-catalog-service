package pdfmaker

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
)

func (p *PdFmaker) HTMLtoPDF(htmlPath string, pdfName string) (string, error) {

	fileName := uuid.NewString()
	bucketName := "file"
	contentType := "application/pdf"

	_, err := exec.LookPath("wkhtmltopdf")
	if err != nil {
		return "", errors.Wrap(err, "error while wkhtmltopdf LookPath")
	}

	cmd := exec.Command("wkhtmltopdf", htmlPath, pdfName+".pdf")
	if err := cmd.Run(); err != nil {
		return "", errors.Wrap(err, "error while wkhtmltopdf Command")
	}

	file, err := os.Open("./" + pdfName + ".pdf")
	if err != nil {
		return "", errors.Wrap(err, "error while os.Open pdf file")
	}

	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return "", errors.Wrap(err, "error while getting pdf file.Stat()")
	}

	_, err = p.minioClient.PutObject(
		context.Background(),
		bucketName,
		fileName,
		file,
		fileStat.Size(),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", errors.Wrap(err, "error while upload html file to minio")
	}

	err = os.Remove(htmlPath)
	if err != nil {
		return "", errors.Wrap(err, "error while remove html. htmlPath")
	}

	err = os.Remove("./" + pdfName + ".pdf")
	if err != nil {
		return "", errors.Wrap(err, "error while remove pdf. pdfName")
	}

	return fmt.Sprintf("https://%s/%s/%s", p.cfg.MinioEndpoint, bucketName, fileName), nil
}

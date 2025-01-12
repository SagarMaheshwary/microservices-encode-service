package aws

import (
	"bytes"
	"io"
	"net/http"
	"os"

	awslib "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/config"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/logger"
)

func NewSession() (*session.Session, error) {
	c := config.Conf.AWS

	s, err := session.NewSession(&awslib.Config{
		Region:      awslib.String(c.Region),
		Credentials: credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, ""),
	})

	if err != nil {
		logger.Error("Unable to create aws session: %v", err)

		return nil, err
	}

	return s, nil
}

func UploadObjectToS3(filePath string, uploadPath string) error {
	c := config.Conf.AWS

	s, err := NewSession()

	if err != nil {
		return err
	}

	svc := s3.New(s)

	f, err := os.ReadFile(filePath)

	if err != nil {
		logger.Error("Unable to read file for upload %v", err)

		return err
	}

	fileStat, err := os.Stat(filePath)

	if err != nil {
		logger.Error("Unable to read file stats %v", err)

		return err
	}

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:               awslib.String(c.S3Bucket),
		Key:                  awslib.String(uploadPath),
		Body:                 bytes.NewReader(f),
		ContentLength:        awslib.Int64(fileStat.Size()),
		ContentType:          awslib.String(http.DetectContentType(f)),
		ContentDisposition:   awslib.String("attachment"),
		ServerSideEncryption: awslib.String("AES256"),
	})

	if err != nil {
		logger.Error("File upload failed %v", err)

		return err
	}

	return nil
}

func GetS3Object(key string) (*s3.GetObjectOutput, error) {
	c := config.Conf.AWS
	s, err := NewSession()
	svc := s3.New(s)

	if err != nil {
		return nil, err
	}

	res, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: awslib.String(c.S3Bucket),
		Key:    awslib.String(key),
	})

	if err != nil {
		logger.Error("S3 get object error %v", err)
	}

	return res, err
}

func DownloadS3Object(filename string, downloadPath string) error {
	res, err := GetS3Object(filename)

	if err != nil {
		return err
	}

	b, err := io.ReadAll(res.Body)

	if err != nil {
		logger.Error("Unable to read file bytes %v", err)

		return err
	}

	f, err := os.Create(downloadPath)

	if err != nil {
		logger.Error("Unable to create file %v", err)

		return err
	}

	_, err = f.Write(b)

	if err != nil {
		logger.Error("Unable to write to file %v", err)

		return err
	}

	return nil
}

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
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/log"
)

func NewSession() (*session.Session, error) {
	c := config.Conf.AWS

	s, err := session.NewSession(&awslib.Config{
		Region:      awslib.String(c.Region),
		Credentials: credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, ""),
	})

	if err != nil {
		log.Error("Unable to create aws session: %v", err)

		return nil, err
	}

	return s, nil
}

func UploadObjectToS3(filePath string, uploadPath string) error {
	c := config.Conf.AWS

	sess, err := NewSession()

	if err != nil {
		return err
	}

	svc := s3.New(sess)

	file, err := os.ReadFile(filePath)

	if err != nil {
		log.Error("Unable to read file for upload %v", err)

		return err
	}

	fileStat, err := os.Stat(filePath)

	if err != nil {
		log.Error("Unable to read file stats %v", err)

		return err
	}

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:               awslib.String(c.S3Bucket),
		Key:                  awslib.String(uploadPath),
		Body:                 bytes.NewReader(file),
		ContentLength:        awslib.Int64(fileStat.Size()),
		ContentType:          awslib.String(http.DetectContentType(file)),
		ContentDisposition:   awslib.String("attachment"),
		ServerSideEncryption: awslib.String("AES256"),
	})

	if err != nil {
		log.Error("File upload failed %v", err)

		return err
	}

	return nil
}

func GetS3Object(key string) (*s3.GetObjectOutput, error) {
	c := config.Conf.AWS

	sess, err := NewSession()
	svc := s3.New(sess)

	if err != nil {
		return nil, err
	}

	file, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: awslib.String(c.S3Bucket),
		Key:    awslib.String(key),
	})

	if err != nil {
		log.Error("S3 get object error %v", err)
	}

	return file, err
}

func DownloadS3Object(filename string, downloadPath string) error {
	res, err := GetS3Object(filename)

	if err != nil {
		return err
	}

	bytes, err := io.ReadAll(res.Body)

	if err != nil {
		log.Error("Unable to read file bytes %v", err)

		return err
	}

	f, err := os.Create(downloadPath)

	if err != nil {
		log.Error("Unable to create file %v", err)

		return err
	}

	_, err = f.Write(bytes)

	if err != nil {
		log.Error("Unable to write to file %v", err)

		return err
	}

	return nil
}

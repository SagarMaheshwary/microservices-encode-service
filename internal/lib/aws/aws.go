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

func NewS3Session() (*s3.S3, error) {
	c := config.GetS3()

	sess, err := session.NewSession(&awslib.Config{
		Region:      awslib.String(c.Region),
		Credentials: credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, ""),
	})

	if err != nil {
		log.Error("Unable to create s3 session: %v", err)

		return nil, err
	}

	svc := s3.New(sess)

	return svc, nil
}

func GetFileFromS3(key string) (*s3.GetObjectOutput, error) {
	c := config.GetS3()

	svc, err := NewS3Session()

	if err != nil {
		return nil, err
	}

	file, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: awslib.String(c.Bucket),
		Key:    awslib.String(key),
	})

	if err != nil {
		log.Error("S3 get object error %v", err)
	}

	return file, err
}

func PutFileToS3(path string, uploadPath string) error {
	c := config.GetS3()

	svc, err := NewS3Session()

	if err != nil {
		return err
	}

	file, err := os.ReadFile(path)

	if err != nil {
		log.Error("Unable to read file for upload %v", err)

		return err
	}

	fileStat, err := os.Stat(path)

	if err != nil {
		log.Error("Unable to read file stats %v", err)

		return err
	}

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:               awslib.String(c.Bucket),
		Key:                  awslib.String(uploadPath),
		ACL:                  awslib.String("private"),
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

func DownloadFileFromS3(filename string, downloadPath string) error {
	res, err := GetFileFromS3(filename)

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

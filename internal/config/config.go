package config

import (
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gofor-little/env"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/helper"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/log"
)

var Conf *Config

type Config struct {
	AWS  *AWS
	AMQP *AMQP
}

type AWS struct {
	Region               string
	S3Bucket             string
	AccessKey            string
	SecretKey            string
	S3PresignedURLExpiry int
	CloudFrontURL        string
}

type AMQP struct {
	Host           string
	Port           int
	Username       string
	Password       string
	PublishTimeout time.Duration
}

func Init() {
	envPath := path.Join(helper.RootDir(), "..", ".env")

	if err := env.Load(envPath); err != nil {
		log.Fatal("Failed to load %q: %v", envPath, err)
	}

	log.Info("Loaded %q", envPath)

	Conf = &Config{
		AWS: &AWS{
			Region:               getEnv("AWS_REGION", ""),
			AccessKey:            getEnv("AWS_ACCESS_KEY", ""),
			SecretKey:            getEnv("AWS_SECRET_KEY", ""),
			S3Bucket:             getEnv("AWS_S3_BUCKET", ""),
			S3PresignedURLExpiry: getEnvInt("AWS_S3_PRESIGNED_URL_EXPIRY", 15),
			CloudFrontURL:        getEnv("AWS_CLOUDFRONT_URL", ""),
		},
		AMQP: &AMQP{
			Host:           getEnv("AMQP_HOST", "localhost"),
			Port:           getEnvInt("AMQP_PORT", 5672),
			Username:       getEnv("AMQP_USERNAME", "guest"),
			Password:       getEnv("AMQP_PASSWORD", "guest"),
			PublishTimeout: getEnvDuration("AMQP_PUBLISH_TIMEOUT_SECONDS", 5),
		},
	}
}

func getEnv(key string, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return val
	}

	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return time.Duration(val) * time.Second
	}

	return defaultVal * time.Second
}

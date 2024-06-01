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

var conf *Config

type Config struct {
	S3   *s3
	AMQP *amqp
}

type s3 struct {
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
}

type amqp struct {
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

	amqpPort, err := strconv.Atoi(Getenv("AMQP_PORT", "5672"))

	if err != nil {
		log.Error("Invalid AMQP_PORT value %v", err)
	}

	amqpPublishTimeout, err := strconv.Atoi(Getenv("AMQP_PUBLISH_TIMEOUT_SECONDS", "5"))

	if err != nil {
		log.Error("Invalid AMQP_PORT value %v", err)
	}

	conf = &Config{
		S3: &s3{
			Bucket:    Getenv("AWS_S3_BUCKET", ""),
			Region:    Getenv("AWS_S3_REGION", ""),
			AccessKey: Getenv("AWS_S3_ACCESS_KEY", ""),
			SecretKey: Getenv("AWS_S3_SECRET_KEY", ""),
		},
		AMQP: &amqp{
			Host:           Getenv("AMQP_HOST", "localhost"),
			Port:           amqpPort,
			Username:       Getenv("AMQP_USERNAME", "guest"),
			Password:       Getenv("AMQP_PASSWORD", "guest"),
			PublishTimeout: time.Duration(amqpPublishTimeout) * time.Second,
		},
	}
}

func GetS3() *s3 {
	return conf.S3
}

func Getamqp() *amqp {
	return conf.AMQP
}

func Getenv(key string, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return defaultVal
}

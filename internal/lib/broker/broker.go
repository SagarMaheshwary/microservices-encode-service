package broker

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	amqplib "github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/config"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/constant"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/consumer"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/logger"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/publisher"
)

var (
	Conn          *amqplib.Connection
	reconnectLock sync.Mutex
)

func MaintainConnection(ctx context.Context) error {
	if err := connect(); err != nil {
		logger.Error("Initial AMQP connection attempt failed: %v", err)
	}

	attempts := config.Conf.AMQP.ConnectionRetryAttempts
	intervalSeconds := config.Conf.AMQP.ConnectionRetryIntervalSeconds

	t := time.NewTicker(intervalSeconds)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			if err := tryReconnect(attempts, intervalSeconds); err != nil {
				logger.Error(err.Error())
				return err
			}
		}
	}
}

func connect() error {
	c := config.Conf.AMQP
	address := fmt.Sprintf("%s://%s:%s@%s:%d", constant.ProtocolAMQP, c.Username, c.Password, c.Host, c.Port)
	var err error

	Conn, err = amqplib.Dial(address)
	if err != nil {
		logger.Error("AMQP connection error %v", err)
		return err
	}

	logger.Info("AMQP connected on %q", address)

	publisherChan, err := NewChannel()
	if err != nil {
		logger.Error("Unable to create publisher channel %v", err)
		return err
	}
	publisher.Init(publisherChan)

	consumerChan, err := NewChannel()
	if err != nil {
		logger.Error("Unable to create consumer channel %v", err)
		return err
	}
	go consumer.Init(consumerChan).Consume()

	return nil
}

func tryReconnect(attempts int, intervalSeconds time.Duration) error {
	reconnectLock.Lock()
	defer reconnectLock.Unlock()

	if HealthCheck() {
		return nil
	}

	for i := range attempts {
		logger.Info("AMQP connection attempt: %d, interval: %v", i+1, intervalSeconds*(1<<i))

		if err := connect(); err == nil {
			return nil
		}

		if i+1 < attempts {
			//retry with exponential backoff
			exponent := math.Pow(2, float64(i))
			delay := time.Duration(float64(intervalSeconds) * exponent)
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("could not reconnect after %d retries", attempts)
}

func NewChannel() (*amqplib.Channel, error) {
	c, err := Conn.Channel()
	if err != nil {
		logger.Error("AMQP channel error %v", err)
		return nil, err
	}

	return c, nil
}

func HealthCheck() bool {
	if Conn == nil || Conn.IsClosed() {
		logger.Info("AMQP health check failed!")
		return false
	}

	return true
}

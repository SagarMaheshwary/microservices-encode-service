package consumer

import (
	"encoding/json"
	"time"

	amqplib "github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/constant"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/handler"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/broker"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/logger"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/prometheus"
)

var C *Consumer

type Consumer struct {
	channel *amqplib.Channel
}

func (c *Consumer) Consume() error {
	q, err := c.declareQueue(constant.QueueEncodeService)

	if err != nil {
		logger.Fatal("AMQP queue listen failed %v", err)
	}

	messages, err := c.channel.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		logger.Fatal("AMQP queue listen failed %v", err)
	}

	logger.Info("AMQP listening on queue %q", constant.QueueEncodeService)

	var forever chan struct{}

	go func() {
		for message := range messages {

			m := broker.MessageType{}
			json.Unmarshal(message.Body, &m)

			logger.Info("AMQP Message received %q: %v", m.Key, m.Data)

			prometheus.TotalMessagesCounter.WithLabelValues(m.Key).Inc()
			start := time.Now()

			switch m.Key {
			case constant.MessageTypeEncodeUploadedVideo:
				type MessageType struct {
					Key  string                       `json:"key"`
					Data handler.VideoUploadedMessage `json:"data"`
				}

				d := new(MessageType)

				json.Unmarshal(message.Body, &d)

				err := handler.ProcessVideoUploadedMessage(&d.Data)

				if err == nil {
					message.Ack(false)
					prometheus.MessageProcessingDuration.WithLabelValues(m.Key).Observe(time.Since(start).Seconds())
				} else {
					logger.Error("Failed to process message %s: %s", m.Key, err)
					prometheus.MessageProcessingErrorsCounter.WithLabelValues(m.Key, err.Error()).Inc()
				}
			}
		}
	}()

	<-forever

	return nil
}

func (c *Consumer) declareQueue(queue string) (*amqplib.Queue, error) {
	q, err := c.channel.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		logger.Error("AMQP declare queue error %v", err)

		return nil, err
	}

	return &q, err
}

func Init(channel *amqplib.Channel) *Consumer {
	C = &Consumer{channel: channel}

	return C
}

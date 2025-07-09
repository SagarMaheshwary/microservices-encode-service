package consumer

import (
	"context"
	"encoding/json"
	"time"

	amqplib "github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/constant"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/handler"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/logger"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

var C *Consumer

type Consumer struct {
	channel *amqplib.Channel
}

type MessageType struct {
	Key  string `json:"key"`
	Data any    `json:"data"`
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

	go func() {
		for message := range messages {
			ctx := contextWithOtelHeaders(message.Headers)

			tracer := otel.Tracer(constant.ServiceName)
			ctx, span := tracer.Start(ctx, constant.TraceTypeRabbitMQConsume)

			m := MessageType{}
			if err := json.Unmarshal(message.Body, &m); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to unmarshal base message")
				logger.Error("Failed to unmarshal message body: %v", err)

				continue
			}

			span.SetAttributes(attribute.String("message_key", m.Key))

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

				if err := json.Unmarshal(message.Body, d); err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, "failed to unmarshal message")
					logger.Error("Failed to unmarshal message %s: %s", m.Key, err)
					continue
				}

				err := handler.ProcessVideoUploadedMessage(ctx, &d.Data)
				if err == nil {
					message.Ack(false)
					prometheus.MessageProcessingDuration.WithLabelValues(m.Key).Observe(time.Since(start).Seconds())
					span.SetStatus(codes.Ok, "message processed successfully")
				} else {
					span.RecordError(err)
					span.SetStatus(codes.Error, "message processing failed")
					logger.Error("Failed to process message %s: %s", m.Key, err)
					prometheus.MessageProcessingErrorsCounter.WithLabelValues(m.Key, err.Error()).Inc()
				}
			default:
				span.AddEvent("unknown message type")
				logger.Warn("Unknown message key: %s", m.Key)
			}

			span.End()
		}
	}()

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

func contextWithOtelHeaders(headers amqplib.Table) context.Context {
	carrier := make(propagation.MapCarrier)
	for k, v := range headers {
		if str, ok := v.(string); ok {
			carrier[k] = str
		}
	}

	return otel.GetTextMapPropagator().Extract(context.Background(), carrier)
}

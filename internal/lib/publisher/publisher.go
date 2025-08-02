package publisher

import (
	"context"
	"encoding/json"

	amqplib "github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/config"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/constant"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

var P *Publisher

type Publisher struct {
	channel *amqplib.Channel
}

type MessageType struct {
	Key  string `json:"key"`
	Data any    `json:"data"`
}

func (p *Publisher) Publish(ctx context.Context, queue string, message *MessageType) error {
	tracer := otel.Tracer(constant.ServiceName)
	ctx, span := tracer.Start(ctx, constant.TraceTypeRabbitMQPublish)
	span.SetAttributes(attribute.String("message_key", message.Key))
	defer span.End()

	c := config.Conf.AMQP

	q, err := p.declareQueue(queue)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to declare queue")
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, c.PublishTimeoutSeconds)
	defer cancel()

	messageData, err := json.Marshal(&message)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal message")
		logger.Error("Unable to parse message %v", message)
		return err
	}

	// Inject trace context into headers
	headers := amqplib.Table{}
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	for k, v := range carrier {
		headers[k] = v
	}

	err = p.channel.PublishWithContext(
		ctx,
		"",
		q.Name,
		false,
		false,
		amqplib.Publishing{
			ContentType: constant.ContentTypeJSON,
			Body:        messageData,
			Headers:     headers,
		},
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to publish message")
		logger.Error("AMQP Unable to publish message %v", err)
		return err
	}

	span.SetStatus(codes.Ok, "message published")
	logger.Info("Message %q Sent", message.Key)

	return nil
}

func (p *Publisher) declareQueue(queue string) (*amqplib.Queue, error) {
	q, err := p.channel.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		logger.Error("Declare queue error %v", err)

		return nil, err
	}

	return &q, err
}

func Init(channel *amqplib.Channel) {
	P = &Publisher{channel: channel}
}

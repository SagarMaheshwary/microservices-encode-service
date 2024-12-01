package consumer

import (
	"encoding/json"

	amqplib "github.com/rabbitmq/amqp091-go"
	cons "github.com/sagarmaheshwary/microservices-encode-service/internal/constant"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/handler"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/broker"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/log"
)

var C *Consumer

type Consumer struct {
	channel *amqplib.Channel
}

func (c *Consumer) Consume() error {
	q, err := c.declareQueue(cons.QueueEncodeService)

	if err != nil {
		log.Fatal("Broker queue listen failed %v", err)
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
		log.Fatal("Broker queue listen failed %v", err)
	}

	log.Info("Broker listening on queue %q", cons.QueueEncodeService)

	var forever chan struct{}

	go func() {
		for message := range messages {
			m := broker.MessageType{}
			json.Unmarshal(message.Body, &m)

			log.Info("AMQP Message received %q: %v", m.Key, m.Data)

			switch m.Key {
			case cons.MessageTypeEncodeUploadedVideo:
				type MessageType struct {
					Key  string                       `json:"key"`
					Data handler.VideoUploadedMessage `json:"data"`
				}

				d := new(MessageType)

				json.Unmarshal(message.Body, &d)

				err := handler.ProcessVideoUploadedMessage(&d.Data)

				if err == nil {
					message.Ack(false)
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
		log.Error("Declare queue error %v", err)

		return nil, err
	}

	return &q, err
}

func Init(channel *amqplib.Channel) *Consumer {
	C = &Consumer{channel: channel}

	return C
}

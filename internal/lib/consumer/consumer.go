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
			s := broker.MessageType{}
			json.Unmarshal(message.Body, &s)

			log.Info("Message received %q: %v", s.Key, s.Data)

			switch s.Key {
			case cons.MessageTypeEncodeUploadedVideo:
				if s.Data != nil {
					data, ok := s.Data.(map[string]interface{})

					if !ok {
						log.Error("Invalid message data %v", message.Body)
						break
					}

					err := handler.ProcessVideoUploaded(&handler.VideoUploadedPayload{
						UploadId:    data["upload_id"].(string),
						Title:       data["title"].(string),
						Description: data["description"].(string),
						PublishedAt: data["published_at"].(string),
						UserId:      int(data["user_id"].(float64)),
					})

					if err == nil {
						message.Ack(false)
					}
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

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

func (c *Consumer) Consume() {
	q, err := c.channel.QueueDeclare(
		cons.QueueEncodeService, // name
		true,                    // durable
		false,                   // delete when unused
		false,                   // exclusive
		false,                   // no-wait
		nil,                     // arguments
	)

	if err != nil {
		log.Error("AMQP queue error %v", err)
	}

	messages, err := c.channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		log.Fatal("AMQP queue listen failed %v", err)
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
					})

					if err == nil {
						message.Ack(false)
					}
				}
			}
		}
	}()

	<-forever
}

func Init(channel *amqplib.Channel) *Consumer {
	C = &Consumer{channel: channel}

	return C
}

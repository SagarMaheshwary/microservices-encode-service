package amqp

import (
	"encoding/json"
	"fmt"

	amqplib "github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/config"
	cons "github.com/sagarmaheshwary/microservices-encode-service/internal/constant"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/log"
	qh "github.com/sagarmaheshwary/microservices-encode-service/internal/queue_handler"
)

var Conn *amqplib.Connection
var Channel *amqplib.Channel

type MessageType struct {
	Key  string `json:"key"`
	Data any    `json:"data"`
}

func Connect() {
	c := config.Getamqp()

	address := fmt.Sprintf("amqp://%s:%s@%s:%d", c.Username, c.Password, c.Host, c.Port)

	var err error

	Conn, err = amqplib.Dial(address)

	if err != nil {
		log.Error("AMQP connection error %v", err)
	}

	Channel, err = Conn.Channel()

	if err != nil {
		log.Error("AMQP channel error %v", err)
	}

	log.Info("AMQP connected on %q", address)
}

func ListenForMessages() {
	q, err := Channel.QueueDeclare(
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

	messages, err := Channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		log.Fatal("AMQP queue listen failed %v", err)
	}

	var forever chan struct{}

	go func() {
		for message := range messages {
			s := MessageType{}
			json.Unmarshal(message.Body, &s)
			log.Info("Message received %#v", s.Data)

			switch s.Key {
			case cons.MessageTypeEncodeUploadedVideo:
				if s.Data != nil {
					data, ok := s.Data.(map[string]interface{})

					if !ok {
						log.Error("Invalid message data %v", message.Body)
						continue
					}

					qh.HandleProcessUploadedVideo(&qh.ProcessUploadedVideoPayload{
						UploadId:    data["upload_id"].(string),
						Title:       data["title"].(string),
						Description: data["description"].(string),
						PublishedAt: data["published_at"].(string),
					})
				}
			}
		}
	}()

	<-forever
}

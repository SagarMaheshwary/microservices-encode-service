package broker

import (
	"fmt"

	amqplib "github.com/rabbitmq/amqp091-go"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/config"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/log"
)

var Conn *amqplib.Connection

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

	log.Info("AMQP connected on %q", address)
}

func NewChannel() (*amqplib.Channel, error) {
	c, err := Conn.Channel()

	if err != nil {
		log.Error("AMQP channel error %v", err)

		return nil, err
	}

	return c, nil
}

package main

import (
	"github.com/sagarmaheshwary/microservices-encode-service/internal/config"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/amqp"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/log"
)

func main() {
	log.Init()
	config.Init()

	amqp.Connect()
	defer amqp.Channel.Close()
	defer amqp.Conn.Close()

	amqp.ListenForMessages()
}

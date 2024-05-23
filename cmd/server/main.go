package main

import (
	"github.com/sagarmaheshwary/microservices-encode-service/internal/config"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/broker"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/consumer"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/log"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/publisher"
)

func main() {
	log.Init()
	config.Init()

	broker.Connect()
	defer broker.Conn.Close()

	publishChan, err := broker.NewChannel()

	if err != nil {
		log.Fatal("Unable to create listen channel %v", err)
	}

	publisher.Init(publishChan)

	listenChan, err := broker.NewChannel()

	if err != nil {
		log.Fatal("Unable to create listen channel %v", err)
	}

	c := consumer.Init(listenChan)

	c.Consume()
}

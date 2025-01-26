package main

import (
	"github.com/sagarmaheshwary/microservices-encode-service/internal/config"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/grpc/server"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/broker"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/consumer"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/logger"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/prometheus"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/publisher"
)

func main() {
	logger.Init()
	config.Init()

	broker.Connect()
	defer broker.Conn.Close()

	publishChan, err := broker.NewChannel()

	if err != nil {
		logger.Fatal("Unable to create publish channel %v", err)
	}

	publisher.Init(publishChan)

	listenChan, err := broker.NewChannel()

	if err != nil {
		logger.Fatal("Unable to create listen channel %v", err)
	}

	c := consumer.Init(listenChan)

	go func() {
		c.Consume()
	}()

	go func() {
		prometheus.Connect()
	}()

	server.Connect()
}

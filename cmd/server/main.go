package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/sagarmaheshwary/microservices-encode-service/internal/config"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/grpc/server"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/broker"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/jaeger"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/logger"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/prometheus"
	"google.golang.org/grpc"
)

func main() {
	logger.Init()
	config.Init()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	shutdownJaeger := jaeger.Init(ctx)

	go func() {
		if err := broker.MaintainConnection(ctx); err != nil {
			//Since the main purpose of this service is to receive rabbitmq messages and
			//encode videos so we will exit the application if it fails to connect to rabbitmq.
			stop()
		}
	}()

	promServer := prometheus.NewServer()
	go func() {
		if err := prometheus.Serve(promServer); err != nil && err != http.ErrServerClosed {
			stop()
		}
	}()

	grpcServer := server.NewServer()
	go func() {
		if err := server.Serve(grpcServer); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			stop()
		}
	}()

	<-ctx.Done()

	logger.Info("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := shutdownJaeger(shutdownCtx); err != nil {
		logger.Warn("failed to shutdown jaeger tracer: %v", err)
	}

	shutdownCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := promServer.Shutdown(shutdownCtx); err != nil {
		logger.Warn("Prometheus server shutdown error: %v", err)
	}

	grpcServer.GracefulStop()

	logger.Info("Shutdown complete")
}

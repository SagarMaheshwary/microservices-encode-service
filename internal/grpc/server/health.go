package server

import (
	"context"

	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/broker"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/logger"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type healthServer struct {
	healthpb.HealthServer
}

func (h *healthServer) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	status := getServicesHealthStatus()

	logger.Info("Overall health status: %q", status)

	return &healthpb.HealthCheckResponse{
		Status: status,
	}, nil
}

func getServicesHealthStatus() healthpb.HealthCheckResponse_ServingStatus {
	if !broker.HealthCheck() {
		return healthpb.HealthCheckResponse_NOT_SERVING
	}

	return healthpb.HealthCheckResponse_SERVING
}

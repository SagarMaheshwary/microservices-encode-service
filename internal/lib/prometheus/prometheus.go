package prometheus

import (
	"net/http"

	prometheuslib "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/config"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/logger"
)

var (
	TotalMessagesCounter = prometheuslib.NewCounterVec(
		prometheuslib.CounterOpts{
			Name: "messages_total",
			Help: "Total number of messages received from RabbitMQ",
		},
		[]string{"message_type"},
	)

	MessageProcessingDuration = prometheuslib.NewHistogramVec(
		prometheuslib.HistogramOpts{
			Name:    "message_processing_duration_seconds",
			Help:    "Time taken to process each message.",
			Buckets: prometheuslib.DefBuckets,
		},
		[]string{"message_type"},
	)

	MessageProcessingErrorsCounter = prometheuslib.NewCounterVec(
		prometheuslib.CounterOpts{
			Name: "message_processing_errors_total",
			Help: "Total number of message processing failures.",
		},
		[]string{"message_type", "reason"},
	)

	ServiceHealth = prometheuslib.NewGauge(prometheuslib.GaugeOpts{
		Name: "service_health_status",
		Help: "Health status of the service: 1=Healthy, 0=Unhealthy",
	})
)

func NewServer() *http.Server {
	prometheuslib.MustRegister(
		TotalMessagesCounter,
		MessageProcessingDuration,
		MessageProcessingErrorsCounter,
		ServiceHealth,
	)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    config.Conf.Prometheus.URL,
		Handler: mux,
	}

	return server
}

func Serve(server *http.Server) error {
	logger.Info("Starting Prometheus metrics server %s", server.Addr)

	if err := server.ListenAndServe(); err != nil {
		logger.Error("Prometheus http server error! %v", err)
		return err
	}

	return nil
}

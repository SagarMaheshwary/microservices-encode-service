# MICROSERVICES - ENCODE SERVICE

Encode service for the [Microservices](https://github.com/SagarMaheshwary/microservices) project.

### OVERVIEW

- Golang
- ZeroLog
- RabbitMQ – Enables asynchronous communication with the [upload service](https://github.com/SagarMaheshwary/microservices-upload-service) and [video catalog service](https://github.com/SagarMaheshwary/microservices-video-catalog-service).
- Prometheus Client – Exports default and custom metrics for Prometheus server monitoring
- AWS S3 – Stores processed video chunks and DASH manifests
- FFmpeg – Handles video encoding and processing
- Jaeger – Distributed request tracing

### SETUP

Follow the instructions in the [README](https://github.com/SagarMaheshwary/microservices?tab=readme-ov-file#setup) of the main microservices repository to run this service along with others using Docker Compose or Kubernetes (KIND).
### APIs (gRPC)

| SERVICE                                                        | RPC   | BODY | METADATA | DESCRIPTION          |
| -------------------------------------------------------------- | ----- | ---- | -------- | -------------------- |
| [Health](https://google.golang.org/grpc/health/grpc_health_v1) | Check | -    | -        | Service health check |

### APIs (REST)

| API      | METHOD | BODY | Headers | Description                 |
| -------- | ------ | ---- | ------- | --------------------------- |
| /metrics | GET    | -    | -       | Prometheus metrics endpoint |

### RABBITMQ MESSAGES

#### Received Messages (Consumed from the Queue)

| MESSAGE NAME        | RECEIVED FROM                                                                     | DESCRIPTION                                                        |
| ------------------- | --------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| EncodeUploadedVideo | [Upload Service](https://github.com/SagarMaheshwary/microservices-upload-service) | Processes uploaded raw video to generate chunks and DASH manifests |

#### Sent Messages (Published to the Queue)

| MESSAGE NAME           | SENT TO                                                                                         | DESCRIPTION                                                                              |
| ---------------------- | ----------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| VideoEncodingCompleted | [Video Catalog Service](https://github.com/SagarMaheshwary/microservices-video-catalog-service) | Notifies video catalog service that video encoding is complete and metadata is available |

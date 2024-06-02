# MICROSERVICES - ENCODE SERVICE

This service is a part of the Microservices project built for encoding uploaded videos into multiple formats, create MPEG-DASH format chunks, and upload it to S3 for streaming.

NOTE: This service doesn't have a gRPC server as it's only listening for rabbitmq messages for background processing.

### TECHNOLOGIES

- Golang (1.22.2)
- RabbitMQ (3.12)
- Amazon S3
- FFMPEG

### SETUP

cd into the project directory and copy **.env.example** to **.env** and update the required variables.

Create executable and start the server:

```bash
go build cmd/server/main.go && ./main
```

Or install "[air](https://github.com/cosmtrek/air)" and run it to autoreload when making file changes:

```bash
air -c .air-toml
```

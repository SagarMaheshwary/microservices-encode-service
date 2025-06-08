FROM golang:1.23 AS builder

WORKDIR /app

ARG MODE=production

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# create executable in case of production mode
RUN if [ "$MODE" = "production" ]; then CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o . ./cmd/server/main.go; fi

FROM alpine:3.21 AS production

WORKDIR /app

RUN apk update && apk add --no-cache ffmpeg

COPY --from=builder /app/main /app/main
COPY --from=builder /app/videos /app/videos

CMD [ "./main" ]

FROM builder AS development

WORKDIR /app

COPY --from=builder /app /app

RUN apt-get update && apt-get install ffmpeg -y

RUN go install github.com/air-verse/air@v1.52.3

CMD ["air", "-c", ".air.toml"]

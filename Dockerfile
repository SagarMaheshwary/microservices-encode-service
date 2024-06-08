FROM golang:1.22

WORKDIR /app

RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN apt-get update

RUN apt-get install ffmpeg -y

CMD ["air", "-c", ".air.toml"]

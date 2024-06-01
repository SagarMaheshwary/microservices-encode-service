FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN go get github.com/cosmtrek/air

RUN go install github.com/cosmtrek/air

RUN apt-get update

RUN apt-get install ffmpeg -y

COPY . .

CMD [ "/go/bin/air" ]

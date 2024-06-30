FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY ./ ./

RUN GOOS=linux GOARCH=amd64 go build -o /my_app ./cmd

CMD ["/my_app"]
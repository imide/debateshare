FROM golang:1.26-alpine

WORKDIR /app

RUN go install -v github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

CMD ["air", "-c", ".air.toml", "--", "run"]
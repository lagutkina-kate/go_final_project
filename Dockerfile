FROM golang:1.26.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o go-final-project

FROM alpine:3.21

ENV TODO_PORT=:7540 TODO_DBFILE="scheduler.db" TODO_PASSWORD=12345

WORKDIR /app

COPY --from=builder /app/go-final-project .
COPY web ./web
COPY scheduler.db .

EXPOSE 7540

CMD ["./go-final-project"]

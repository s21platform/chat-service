FROM golang:1.24 as builder

WORKDIR /usr/src/service
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . .

RUN go build -o build/main cmd/service/main.go
RUN go build -o build/worker_kafka_user cmd/workers/kafka/user/main.go
RUN go build -o build/worker_kafka_avatar cmd/workers/kafka/avatar/main.go

FROM alpine

WORKDIR /app

COPY --from=builder /usr/src/service/build/main /app
COPY --from=builder /usr/src/service/build/worker_kafka_user .
COPY --from=builder /usr/src/service/build/worker_kafka_avatar .

RUN apk add --no-cache gcompat
RUN chmod +x main worker_kafka_user worker_kafka_avatar

CMD ./main & ./worker_kafka_user & ./worker_kafka_avatar
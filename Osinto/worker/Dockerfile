FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY worker/worker.go ./worker/

RUN CGO_ENABLED=0 GOOS=linux \
    go build \
    -o worker \
    ./worker


FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/worker .

ENTRYPOINT ["./worker"]

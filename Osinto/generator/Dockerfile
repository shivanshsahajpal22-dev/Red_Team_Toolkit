FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY generator.go .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o generator \
    generator.go


FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/generator .

ENTRYPOINT ["./generator"]

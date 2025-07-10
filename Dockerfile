FROM registry.vsfi.ru/library/golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go get github.com/nats-io/nats.go@v1.31.0
RUN go get github.com/sirupsen/logrus@v1.9.3
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o factory ./cmd/main.go

FROM registry.vsfi.ru/library/alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/factory .
COPY --from=builder /app/web ./web

EXPOSE 8080
CMD ["./factory"] 
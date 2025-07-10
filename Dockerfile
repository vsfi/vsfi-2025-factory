FROM registry.vsfi.ru/library/golang:1.23-bookworm AS builder
ENV GOPROXY=https://go-proxy-user:fn298f0g21fwr@nexus.vsfi.ru/repository/go-mod-shisha-server
RUN printf "deb [trusted=yes] https://nexus.vsfi.ru/repository/debian-12/ bookworm main non-free-firmware\ndeb [trusted=yes] https://nexus.vsfi.ru/repository/debian-12/ bookworm-updates main non-free-firmware\ndeb [trusted=yes] https://nexus.vsfi.ru/repository/debian-12-security/ bookworm-security main\ndeb [trusted=yes] https://nexus.vsfi.ru/repository/apt-docker/ bookworm stable\n" > /etc/apt/sources.list
RUN printf "machine nexus.vsfi.ru\nlogin debian\npassword debian\n" > /etc/apt/auth.conf
RUN  apt update && apt install -y git ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY go.mod go.sum ./
RUN GOPROXY=https://go-proxy-user:fn298f0g21fwr@nexus.vsfi.ru/repository/go-mod-shisha-server go mod download -x && \
    go get github.com/nats-io/nats.go@v1.31.0 && \
    go get github.com/sirupsen/logrus@v1.9.3 && \
    go mod tidy
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o factory ./cmd/main.go

FROM registry.vsfi.ru/library/alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/factory .
COPY --from=builder /app/web ./web
EXPOSE 8080
CMD ["./factory"]

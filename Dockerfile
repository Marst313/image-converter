FROM golang:1.25 AS builder

WORKDIR /app

RUN apt-get update && \
    apt-get install -y \
    gcc \
    libc6-dev \
    libwebp-dev

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o app ./cmd/server

FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y \
    ca-certificates \
    libwebp7 && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /root

COPY --from=builder /app/app .

EXPOSE 3001

CMD ["./app"]
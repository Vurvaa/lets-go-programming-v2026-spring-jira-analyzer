FROM golang:1.22.7-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o backend ./cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/backend .
COPY --from=builder /app/configs ./configs

CMD ["./backend"]
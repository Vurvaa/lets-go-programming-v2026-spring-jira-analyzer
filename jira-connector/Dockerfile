FROM golang:1.22.7-alpine AS builder

WORKDIR /app

COPY go.mod ./
#RUN go mod download

COPY . .

RUN go build -o jira-connector ./cmd/connector/main.go

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/jira-connector .

CMD ["./jira-connector"]
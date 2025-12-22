FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o internal-tools

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/internal-tools .
EXPOSE 8080
CMD ["./internal-tools"]

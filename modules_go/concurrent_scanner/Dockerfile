# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o nocturne main.go

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/nocturne .
COPY --from=builder /app/config.yaml .

# Default port for 'serve' command
EXPOSE 8080

# Entry point defaults to help
ENTRYPOINT ["./nocturne"]
CMD ["help"]
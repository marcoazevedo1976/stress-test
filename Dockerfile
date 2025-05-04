FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o stress-test .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/stress-test .

ENTRYPOINT ["/app/stress-test"]
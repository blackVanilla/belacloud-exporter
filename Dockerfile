FROM golang:1.23.6 AS builder

WORKDIR /app
COPY . .
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build -o exporter belabox-exporter.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder --chmod=755 /app/exporter .

CMD ["/app/exporter"]
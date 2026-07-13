FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway ./services/gateway/cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/gateway /gateway
COPY --from=builder /app/configs /configs
COPY --from=builder /app/certs /certs
CMD ["/gateway"]
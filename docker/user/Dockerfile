FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o user ./services/user/cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/user /user
COPY --from=builder /app/configs /configs
COPY --from=builder /app/certs /certs
CMD ["/user"]
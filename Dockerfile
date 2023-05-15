FROM golang:1.20-alpine3.17 AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache git make build-base
RUN CGO_ENABLED=1 go build -trimpath -o botastic

FROM alpine:3.17
WORKDIR /app
COPY --from=builder /app/botastic .
RUN chmod +x /app/botastic
ENTRYPOINT ["/app/botastic"]

FROM alpine:3.17
WORKDIR /app
COPY botastic .
ENTRYPOINT ["/app/botastic"]

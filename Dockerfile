FROM alpine as build
RUN apk --no-cache add ca-certificates

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY botastic /usr/bin/
ENTRYPOINT ["/usr/bin/botastic"]
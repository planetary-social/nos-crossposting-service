FROM golang:1.20-alpine AS build

WORKDIR /usr/src/app
COPY . .
# get ssl certs to copy into scratch image, as it won't have them by default.
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o /usr/local/bin/notification-service ./cmd/notification-service


FROM scratch as app
WORKDIR /
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/local/bin/notification-service /notification-service
CMD ["/notification-service"]

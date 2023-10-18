FROM golang:1.21 AS build

WORKDIR /usr/src/app
COPY . .
# get ssl certs to copy into scratch image, as it won't have them by default.
RUN apt update && apt-get install git ca-certificates && update-ca-certificates
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -v -o /usr/local/bin/crossposting-service ./cmd/crossposting-service


FROM scratch as app
WORKDIR /
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/local/bin/crossposting-service /crossposting-service
CMD ["/crossposting-service"]

FROM golang:1.21-alpine3.19 AS builder
WORKDIR /go/src/github.com/mwmahlberg/so-postgres
COPY . .
RUN go build -o /go/bin/so-postgres

FROM alpine:3.19
COPY --from=builder /go/bin/so-postgres /usr/local/bin/so-postgres
CMD ["so-postgres"]
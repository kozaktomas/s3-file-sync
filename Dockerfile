FROM golang:1.16 AS builder
ENV CGO_ENABLED 0
ADD . /app
WORKDIR /app
RUN go build -i -v -o s3sync main.go

FROM alpine:3
RUN apk update && \
    apk add openssl && \
    rm -rf /var/cache/apk/* \
    && mkdir /app

WORKDIR /app

ADD Dockerfile /Dockerfile

COPY --from=builder /app/s3sync /app/s3sync

RUN chown nobody /app/s3sync \
    && chmod 500 /app/s3sync

USER nobody

ENTRYPOINT ["/app/s3sync"]

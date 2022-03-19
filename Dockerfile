FROM golang:1.17 as builder

WORKDIR /build
# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod .
RUN go mod download && go mod verify
COPY . .
ARG ENCRYPTION_KEYS
ARG DEFAULT_CONFIG_VALUE
ARG CA_PATH_VALUE
ARG PROMETHEUS_BASIC_AUTH
RUN make build_encrypted

FROM alpine:3.11.3 as advanced

RUN apk add --update curl && rm  -rf /tmp/* /var/cache/apk/*

WORKDIR /usr/src/app
COPY --from=builder /build/db1000n .

CMD ["./db1000n", "--enable-primitive=false"]

FROM alpine:3.11.3

RUN apk add --update curl && rm  -rf /tmp/* /var/cache/apk/*

WORKDIR /usr/src/app
COPY --from=builder /build/db1000n .

VOLUME /usr/src/app/config

ENTRYPOINT ["./db1000n"]

FROM golang:1.17 as builder

WORKDIR /build
# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod .
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o main .

FROM alpine:3.11.3 as openvpn

RUN apk add --update supervisor openvpn curl && rm  -rf /tmp/* /var/cache/apk/*

ADD supervisord/supervisord.conf /etc/
ADD supervisord/supervisord-openvpn.conf \
    supervisord/supervisord-db1000n.conf /etc/supervisor/conf.d/
ADD run/openvpn-up.sh \
    run/run-openvpn.sh \
    run/db1000n.sh /usr/local/bin/

RUN chmod +x /usr/local/bin/openvpn-up.sh \
    /usr/local/bin/run-openvpn.sh \
    /usr/local/bin/db1000n.sh

WORKDIR /usr/src/app
COPY --from=builder /build/main .

ENTRYPOINT ["supervisord", "--configuration", "/etc/supervisord.conf"]

FROM alpine:3.11.3

WORKDIR /usr/src/app
COPY --from=builder /build/main .

CMD ["./main"]

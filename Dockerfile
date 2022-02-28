FROM golang:1.17
ARG APP_VERSION
ARG BUILD_TIME
WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod ./
RUN go mod download && go mod verify

COPY . .
RUN go build -ldflags="-X 'main.Version=$APP_VERSION' -X 'main.Time=$BUILD_TIME'" -v -o /usr/local/bin/main ./main.go

CMD ["main"]
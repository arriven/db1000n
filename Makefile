build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ${LDFLAGS} -o main ./main.go

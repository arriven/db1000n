build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ${LDFLAGS} -o main ./main.go
docker_image:
	docker build -f Dockerfile --tag bomber:latest .
terra_init:
	terraform init
terra_plan:
	terraform plan -out=infra.out
terra_apply:
	terraform apply "infra.out"

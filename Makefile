REPOSITORY_BASE_PATH := github.com/Arriven/db1000n
build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o main ./main.go

ifneq ($(ENCRYPTION_KEYS),)
LDFLAGS += -X '$(REPOSITORY_BASE_PATH)/src/utils.EncryptionKeys=$(ENCRYPTION_KEYS)'
endif
ifneq ($(DEFAULT_CONFIG_VALUE),)
LDFLAGS += -X '$(REPOSITORY_BASE_PATH)/src/runner/config.DefaultConfig=$(DEFAULT_CONFIG_VALUE)'
endif
ifneq ($(CA_PATH_VALUE),)
LDFLAGS += -X '$(REPOSITORY_BASE_PATH)/src/metrics.PushGatewayCA=$(CA_PATH_VALUE)'
endif

build_encrypted: build

encrypt_config:
	@if [ "$(DEFAULT_CONFIG)" = "" ]; then \
		echo "Not specified DEFAULT_CONFIG"; \
	else \
		file=`tempfile` && \
		age --encrypt -p --output $${file} $(DEFAULT_CONFIG) && \
		echo "Saved in file: $${file}"; \
		config=`cat $${file} | base64 | tr -d '\n'` && \
		echo "Save value as env variable: \nexport DEFAULT_CONFIG_VALUE='$${config}'"; \
	fi

encrypt_ca:
	@if [ "$(CA_PATH)" = "" ]; then \
		echo "Not specified CA_PATH"; \
	else \
		file=`tempfile` && \
		age --encrypt -p --output $${file} $(CA_PATH) && \
		echo "Saved in file: $${file}"; \
		config=`cat $${file} | base64 | tr -d '\n'` && \
		echo "Save value as env variable: \nexport CA_PATH_VALUE='$${config}'"; \
	fi

docker_build:
	@docker build -t "ghcr.io/arriven/db1000n" -f Dockerfile .
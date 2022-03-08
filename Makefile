build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o main ./main.go

ifneq ($(ENCRYPTION_KEYS),)
LDFLAGS += -X 'github.com/Arriven/db1000n/src/utils.EncryptionKeys=$(ENCRYPTION_KEYS)'
endif
ifneq ($(DEFAULT_CONFIG_VALUE),)
LDFLAGS += -X 'github.com/Arriven/db1000n/src/runner/config.DefaultConfig=$(DEFAULT_CONFIG_VALUE)'
endif

build_encrypted: build

encrypt_config:
	@if [ "$(DEFAULT_CONFIG)" = "" ]; then \
		echo "Not specified DEFAULT_CONFIG"; \
	else \
		file=`tempfile` && \
		echo "Saved in file: $${file}"; \
		age --encrypt -p --output $${file} $(DEFAULT_CONFIG) && \
		config=`cat $${file} | base64 | tr -d '\n'` && \
		echo "Save value as env variable: \nexport DEFAULT_CONFIG_VALUE='$${config}'"; \
	fi

docker_build:
	@docker build -t "ghcr.io/arriven/db1000n" -f Dockerfile .
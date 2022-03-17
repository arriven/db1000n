APP_NAME := db1000n

ifeq ($(GOOS),windows)
	APP_NAME := $(addsuffix .exe,$(APP_NAME))
endif

REPOSITORY_BASE_PATH := github.com/Arriven/db1000n
LATEST_TAG := $(shell git describe --tags --abbrev=0)

# Remove debug information (ELF) to strip the binary size
LDFLAGS += -s -w

ifneq ($(LATEST_TAG),)
LDFLAGS += -X '$(REPOSITORY_BASE_PATH)/src/utils/ota.Version=$(LATEST_TAG)'
endif
ifneq ($(ENCRYPTION_KEYS),)
LDFLAGS += -X '$(REPOSITORY_BASE_PATH)/src/utils.EncryptionKeys=$(ENCRYPTION_KEYS)'
BUILD_TAGS += encrypted
endif
ifneq ($(DEFAULT_CONFIG_VALUE),)
LDFLAGS += -X '$(REPOSITORY_BASE_PATH)/src/runner/config.DefaultConfig=$(DEFAULT_CONFIG_VALUE)'
endif
ifneq ($(CA_PATH_VALUE),)
LDFLAGS += -X '$(REPOSITORY_BASE_PATH)/src/utils/metrics.PushGatewayCA=$(CA_PATH_VALUE)'
endif
ifneq ($(PROMETHEUS_BASIC_AUTH),)
LDFLAGS += -X '$(REPOSITORY_BASE_PATH)/src/utils/metrics.BasicAuth=$(PROMETHEUS_BASIC_AUTH)'
endif

build:
	CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -tags="${BUILD_TAGS}" -o $(APP_NAME) -a ./main.go

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
DOCKER_REGISTRY := ghcr.io/lucasl0st
DOCKER_PLATFORMS := linux/amd64,linux/arm64
GO_PLATFORMS := $(shell echo "$(DOCKER_PLATFORMS)" | tr ',' ' ')
APPLICATION := ipv4-for-ipv6_only-http-proxy
TAG := $(shell git describe --always)

.PHONY: build
build:
	$(foreach platform,$(GO_PLATFORMS), \
		os=$$(echo $(platform) | cut -d'/' -f1); \
		arch=$$(echo $(platform) | cut -d'/' -f2); \
		out="build/bin/$(APPLICATION)_$${os}-$${arch}"; \
		GOOS=$${os} GOARCH=$${arch} CGO_ENABLED=0 go build -o $${out} ./cmd/proxy ; \
		echo "built $${out}"; \
	)\


ADDITIONAL_DOCKER_OPTS :=
ifdef push
	ADDITIONAL_DOCKER_OPTS += --push
endif

ifdef tag-latest
	ADDITIONAL_DOCKER_OPTS += -t $(DOCKER_REGISTRY)/$(APPLICATION):latest
endif

.PHONY: docker
docker: build
	docker buildx build . -t $(DOCKER_REGISTRY)/$(APPLICATION):$(TAG) --platform=$(DOCKER_PLATFORMS) -f Dockerfile $(ADDITIONAL_DOCKER_OPTS)

.PHONY: fmt
fmt:
	go tool mvdan.cc/gofumpt -w .

.PHONY: lint
lint:
	go tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --timeout 5m

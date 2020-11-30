GOARCH = amd64

UNAME = $(shell uname -s)

ifndef OS
	ifeq ($(UNAME), Linux)
		OS = linux
	else ifeq ($(UNAME), Darwin)
		OS = darwin
	endif
endif

.DEFAULT_GOAL := all

all: fmt build lint start 

build:
	GOOS=$(OS) GOARCH="$(GOARCH)" go build -o vault/plugins/vault-plugin-secrets-openstack cmd/vault-plugin-secrets-openstack/main.go

start: build
	vault server -dev -dev-root-token-id=root -dev-plugin-dir=./vault/plugins

enable:
	vault secrets enable -path=openstack vault-plugin-secrets-openstack

clean:
	rm -f ./vault/plugins/vault-plugin-secrets-openstack

fmt:
	go fmt $$(go list ./...)

# Run golangci-lint code
lint:
	golangci-lint run
.PHONY: build clean fmt start enable

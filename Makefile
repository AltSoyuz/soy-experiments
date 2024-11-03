PKG_PREFIX := golang-template-htmx-alpine
BUILDINFO_TAG ?= $(shell echo $$(git describe --long --all | tr '/' '-')$$( \
	      git diff-index --quiet HEAD -- || echo '-dirty-'$$(git diff-index -u HEAD | openssl sha1 | cut -d' ' -f2 | cut -c 1-8)))
LATEST_TAG ?= latest
PKG_TAG ?= $(shell git tag -l --points-at HEAD)
ifeq ($(PKG_TAG),)
PKG_TAG := $(BUILDINFO_TAG)
endif

GO_BUILDINFO = -X '$(PKG_PREFIX)/lib/buildinfo.Version=todo$(DATEINFO_TAG)-$(BUILDINFO_TAG)'

include apps/*/Makefile

vet:
	go vet ./...

fmt:
	go fmt ./...

test:
	go test ./...

test-race:
	go test -race ./...

test-full:
	go test -coverprofile=coverage.txt -covermode=atomic ./...

app-local:
	CGO_ENABLED=1 go build $(RACE) -ldflags "$(GO_BUILDINFO)" -o bin/$(APP_NAME)$(RACE) $(PKG_PREFIX)/apps/$(APP_NAME)

update:
	go get -u ./...

golangci-lint: install-golangci-lint
	golangci-lint run

check-all: fmt vet golangci-lint govulncheck 

clean-checkers: remove-golangci-lint remove-govulncheck sqlc

install-govulncheck:
	which govulncheck || go install golang.org/x/vuln/cmd/govulncheck@latest

govulncheck: install-govulncheck
	govulncheck ./...

install-golangci-lint:
	which golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.60.3

remove-golangci-lint:
	rm -rf `which golangci-lint`

install-sqlc:
	which sqlc || go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

sqlc: install-sqlc
	sqlc vet
	sqlc generate

remove-sqlc:
	rm -rf `which sqlc`
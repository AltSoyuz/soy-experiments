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

update:
	go get -u ./...

golangci-lint: install-golangci-lint
	golangci-lint run

check-all: fmt vet golangci-lint govulncheck

clean-checkers: remove-golangci-lint remove-govulncheck

install-govulncheck:
	which govulncheck || go install golang.org/x/vuln/cmd/govulncheck@latest

govulncheck: install-govulncheck
	govulncheck ./...

install-golangci-lint:
	which golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.60.3

remove-golangci-lint:
	rm -rf `which golangci-lint`
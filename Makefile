PKG_PREFIX := github.com/AltSoyuz/soy-experiments
BUILDINFO_TAG ?= $(shell echo $$(git describe --long --all | tr '/' '-')$$( \
	      git diff-index --quiet HEAD -- || echo '-dirty-'$$(git diff-index -u HEAD | openssl sha1 | cut -d' ' -f2 | cut -c 1-8)))
LATEST_TAG ?= latest
PKG_TAG ?= $(shell git tag -l --points-at HEAD)
ifeq ($(PKG_TAG),)
PKG_TAG := $(BUILDINFO_TAG)
endif
MAKE_CONCURRENCY ?= $(shell getconf _NPROCESSORS_ONLN)
MAKE_PARALLEL := $(MAKE) -j $(MAKE_CONCURRENCY)

GO_BUILDINFO = -X '$(PKG_PREFIX)/lib/buildinfo.Version=todo$(DATEINFO_TAG)-$(BUILDINFO_TAG)'
TAR_OWNERSHIP ?= --owner=1000 --group=1000

TAR_OWNERSHIP ?= --owner=1000 --group=1000

GITHUB_RELEASE_SPEC_FILE="/tmp/maun-github-release"
GITHUB_DEBUG_FILE="/tmp/maun-github-debug"

include apps/*/Makefile
include deployment/Makefile

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
	CGO_ENABLED=1 go build $(RACE) -ldflags "$(GO_BUILDINFO)" -o bin/$(APP_NAME)$(RACE) $(PKG_PREFIX)/apps/$(APP_NAME)/cmd

vendor-update:
	go get -u ./apps/...
	go get -u ./lib/...
	go mod tidy -compat=1.23
	go mod vendor


golangci-lint: install-golangci-lint
	golangci-lint run

check-all: fmt vet golangci-lint govulncheck 

clean-checkers: \
	remove-golangci-lint \
	remove-govulncheck 

install-govulncheck:
	which govulncheck || go install golang.org/x/vuln/cmd/govulncheck@latest

govulncheck: install-govulncheck
	govulncheck ./...

install-golang-migrate:
	which migrate || go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

install-golangci-lint:
	which golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.60.3

install-trufflehog:
	which trufflehog || curl -sSfL https://raw.githubusercontent.com/trufflesecurity/trufflehog/main/scripts/install.sh | sh -s -- -b /usr/local/bin

migrate: install-golang-migrate
	migrate -path apps/$(APP_NAME)/store/migrations \
			-database "sqlite3://$(APP_NAME).db" $(MIGRATE_CMD)

migrate-up: 
	APP_NAME=$(APP_NAME) MIGRATE_CMD=up $(MAKE) migrate

migrate-down:
	APP_NAME=$(APP_NAME) MIGRATE_CMD="down 1" $(MAKE) migrate 

delete-test-db:
	rm -rf $(APP_NAME).test.db

remove-golangci-lint:
	rm -rf `which golangci-lint`

install-sqlc:
	which sqlc || go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

sqlc: install-sqlc
	sqlc vet
	sqlc generate

remove-sqlc:
	rm -rf `which sqlc`

release:
	$(MAKE_PARRALLEL) release-todo

release-todo: \
	release-todo-linux-amd64 \
	release-todo-linux-arm64 

release-todo-linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) release-todo-goos-goarch

release-todo-linux-arm64:
	GOOS=linux GOARCH=arm64 $(MAKE) release-todo-goos-goarch

release-todo: todo-$(GOOS)-$(GOARCH)-prod
	cd bin && \
		tar $(TAR_OWNERSHIP) --transform="flags=r;s|-$(GOOS)-$(GOARCH)||" -czf todo-$(GOOS)-$(GOARCH)-$(PKG_TAG).tar.gz \
			todo-$(GOOS)-$(GOARCH)-prod \
		&& sha256sum todo-$(GOOS)-$(GOARCH)-$(PKG_TAG).tar.gz \
			todo-$(GOOS)-$(GOARCH)-prod \
			| sed s/-$(GOOS)-$(GOARCH)-prod/-prod/ > todo-$(GOOS)-$(GOARCH)-$(PKG_TAG)_checksums.txt
	cd bin && rm -rf todo-$(GOOS)-$(GOARCH)-prod

github-token-check:
ifndef GITHUB_TOKEN
	$(error missing GITHUB_TOKEN env var. It must contain github token for soy experiment project obtained from https://github.com/settings/tokens)
endif

github-tag-check:
ifndef TAG
	$(error missing TAG env var. It must contain github release tag to create)
endif

github-create-release: github-token-check github-tag-check
	@result=$$(curl -o $(GITHUB_RELEASE_SPEC_FILE) -s -w "%{http_code}" \
		-X POST \
		-H "Accept: application/vnd.github+json" \
		-H "Authorization: token $(GITHUB_TOKEN)" \
		https://api.github.com/AltSoyuz/soy-experiments/releases \
		-d '{"tag_name":"$(TAG)","name":"$(TAG)","body":"TODO: put here the changelog for $(TAG) release from docs/CHANGELOG.md","draft":true,"prerelease":false,"generate_release_notes":false}'); \
		if [ $${result} = 201 ]; then \
			release_id=$$(cat $(GITHUB_RELEASE_SPEC_FILE) | grep '"id"' -m 1 | sed -E 's/.* ([[:digit:]]+)\,/\1/'); \
			printf "Created release $(TAG) with id=$${release_id}\n"; \
		else \
			printf "Failed to create release $(TAG)\n"; \
			cat $(GITHUB_RELEASE_SPEC_FILE); \
			exit 1; \
		fi

github-upload-assets:
	@release_id=$$(cat $(GITHUB_RELEASE_SPEC_FILE) | grep '"id"' -m 1 | sed -E 's/.* ([[:digit:]]+)\,/\1/'); \
	$(foreach file, $(wildcard bin/*.zip), FILE=$(file) RELEASE_ID=$${release_id} CONTENT_TYPE="application/zip" $(MAKE) github-upload-asset || exit 1;) \
	$(foreach file, $(wildcard bin/*.tar.gz), FILE=$(file) RELEASE_ID=$${release_id} CONTENT_TYPE="application/x-gzip" $(MAKE) github-upload-asset || exit 1;) \
	$(foreach file, $(wildcard bin/*_checksums.txt), FILE=$(file) RELEASE_ID=$${release_id} CONTENT_TYPE="text/plain" $(MAKE) github-upload-asset || exit 1;) 

github-upload-asset: github-token-check
ifndef FILE
	$(error missing FILE env var. It must contain path to file to upload to github release)
endif
	@printf "Uploading $(FILE)\n"
	@result=$$(curl -o $(GITHUB_DEBUG_FILE) -w "%{http_code}" \
		-X POST \
		-H "Accept: application/vnd.github+json" \
		-H "Authorization: token $(GITHUB_TOKEN)" \
		-H "Content-Type: $(CONTENT_TYPE)" \
		--data-binary "@$(FILE)" \
		https://uploads.github.com/repos/AltSoyuz/soy-experiments/releases/$(RELEASE_ID)/assets?name=$(notdir $(FILE))); \
		if [ $${result} = 201 ]; then \
			printf "Upload OK: $${result}\n"; \
		elif [ $${result} = 422 ]; then \
			printf "Asset already uploaded, you need to delete it from UI if you want to re-upload it\n"; \
		else \
			printf "Upload failed: $${result}\n"; \
			cat $(GITHUB_DEBUG_FILE); \
			exit 1; \
		fi

github-delete-release: github-token-check
	@release_id=$$(cat $(GITHUB_RELEASE_SPEC_FILE) | grep '"id"' -m 1 | sed -E 's/.* ([[:digit:]]+)\,/\1/'); \
	result=$$(curl -o $(GITHUB_DEBUG_FILE) -s -w "%{http_code}" \
		-X DELETE \
		-H "Accept: application/vnd.github+json" \
		-H "Authorization: token $(GITHUB_TOKEN)" \
		https://api.github.com/repos/AltSoyuz/soy-experiments/releases/$${release_id}); \
		if [ $${result} = 204 ]; then \
			printf "Deleted release with id=$${release_id}\n"; \
		else \
			printf "Failed to delete release with id=$${release_id}\n"; \
			cat $(GITHUB_DEBUG_FILE); \
			exit 1; \
		fi
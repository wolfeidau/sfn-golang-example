APPNAME := aws-sfn-golang
STAGE ?= dev
BRANCH ?= master

GOLANGCI_VERSION = 1.32.0

BIN_DIR ?= $(shell pwd)/bin

GIT_HASH := $(shell git rev-parse --short HEAD)

default: clean build archive deploy
.PHONY: default

ci: clean lint test
.PHONY: ci

LDFLAGS := -ldflags="-s -w -X main.version=${GIT_HASH}"

$(BIN_DIR)/golangci-lint: $(BIN_DIR)/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} $(BIN_DIR)/golangci-lint
$(BIN_DIR)/golangci-lint-${GOLANGCI_VERSION}:
	@curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | BINARY=golangci-lint bash -s -- v${GOLANGCI_VERSION}
	@mv $(BIN_DIR)/golangci-lint $@

clean:
	@echo "--- clean all the things"
	@rm -rf ./dist
.PHONY: clean

lint: $(BIN_DIR)/golangci-lint
	@echo "--- lint all the things"
	@$(BIN_DIR)/golangci-lint run
.PHONY: lint

lint-fix: $(BIN_DIR)/golangci-lint
	@echo "--- lint all the things"
	@$(BIN_DIR)/golangci-lint run --fix
.PHONY: lint-fix

test:
	@echo "--- test all the things"
	@go test -coverprofile=coverage.txt ./...
	@go tool cover -func=coverage.txt
.PHONY: test

build:
	@echo "--- build all the things"
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -trimpath -o dist ./cmd/...
.PHONY: build

archive:
	@echo "--- build an archive"
	@cd dist && zip -X -9 -r ./handler.zip *-lambda
.PHONY: archive

deploy:
	@echo "--- deploy stack $(APPNAME)-$(STAGE)-$(BRANCH)"
	@sam deploy \
		--no-fail-on-empty-changeset \
		--template-file sam/app/sfn.yaml \
		--capabilities CAPABILITY_IAM \
		--s3-bucket $(SAM_BUCKET) \
		--s3-prefix sam/$(GIT_HASH) \
		--tags "environment=$(STAGE)" "branch=$(BRANCH)" "service=$(APPNAME)" \
		--stack-name $(APPNAME)-$(STAGE)-$(BRANCH) \
		--parameter-overrides AppName=$(APPNAME) Stage=$(STAGE) Branch=$(BRANCH) Commit=$(GIT_HASH)
.PHONY: deploy
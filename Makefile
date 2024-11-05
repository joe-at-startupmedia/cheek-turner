GO ?= go
GOFMT ?= gofmt "-s"
GOFILES := $(shell find . -name "*.go")
PACKAGES ?= $(shell $(GO) list ./...)
TEST_REGEX := $(or $(TEST_REGEX),"Test")
DEFAULT_TEST_PACKAGES := $(shell  $(GO) list ./... | awk '!/(cmd|mocks)/' | tr "\n" ",")
TEST_PACKAGES := $(or $(TEST_PACKAGES),$(DEFAULT_TEST_PACKAGES))

all: build

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: clean
clean: ## remove files created during build pipeline
	$(call print-target)
	rm -f coverage.*
	rm -f '"$(shell go env GOCACHE)/../golangci-lint"'
	go clean -i -cache -testcache -fuzzcache -x

.PHONY: fmt
fmt: ## format files
	$(call print-target)
	$(GOFMT) -w $(GOFILES)

.PHONY: lint
lint: ## lint files
	$(call print-target)
	golangci-lint run --fix

.PHONY: misspell
misspell: ## check for misspellings
	$(call print-target)
	misspell -error $(GOFILES)

.PHONY: tools
tools: ## go install tools
	$(call print-target)
	cd tools && go install $(shell cd tools && $(GO) list -e -f '{{ join .Imports " " }}' -tags=tools)

.PHONY: mod
mod: ## go mod tidy
	$(call print-target)
	go mod tidy
	cd tools && go mod tidy

.PHONY: build
build: mod fmt tools vuln misspell
	cd tools && $(GO) mod tidy
	$(ENV_VARS) $(GO) build -buildvcs=false $(BUILD_FLAGS) -o bin/cheek-turner cmd/cheek-turner/main.go

.PHONY: test
test: build ## run the tests
	$(call print-target)
	$(GO) test $(BUILD_FLAGS) -v -run $(TEST_REGEX) -p 1 ./...

.PHONY: test_cover
test_cover: build ## run the tests and generate a coverage report
	$(call print-target)
	$(GO) test $(BUILD_FLAGS) -v -run $(TEST_REGEX) -p 1 -coverprofile=coverage.out -coverpkg=$(TEST_PACKAGES) ./...

.PHONY: codecov
codecov: ## process the coverage report and upload it
	$(call print-target)
	codecov upload-process -t $(CODECOV_TOKEN)

.PHONY: test_codecov
test_codecov: test_cover codecov ## run the tests and process/upload the coverage reports
	$(call print-target)

.PHONY: vuln
vuln: ## govulncheck
	$(call print-target)
	govulncheck ./...

.PHONY: install
install: ## install the binary in the systems executable path
	$(call print-target)
	cp -R bin/* /usr/local/bin/

.PHONY: mockery
mockery: ## generates the mocks
	$(call print-target)
	mockery --output mocks --name ElectionInterface --dir pkg --filename election.go --structname Election

.PHONY: npm_build
npm_build: ## npm build client
	$(call print-target)
	cd web && npm run build

define print-target
    @printf "Executing target: \033[36m$@\033[0m\n"
endef

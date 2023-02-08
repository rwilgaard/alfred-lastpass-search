PROJECT_NAME := "alfred-lastpass-search"
PKG := "github.com/rwilgaard/$(PROJECT_NAME)"
GO111MODULE=on

.EXPORT_ALL_VARIABLES:
.PHONY: all dep lint vet build clean

all: build

dep: ## Get the dependencies
	@go mod download

lint: ## Lint Golang files
	@golangci-lint run --timeout 3m

vet: ## Run go vet
	@go vet ./src

build: dep ## Build the binary file
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o workflow/$(PROJECT_NAME)-amd64 ./src
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o workflow/$(PROJECT_NAME)-arm64 ./src

universal-binary:
	@lipo -create -output workflow/$(PROJECT_NAME) workflow/$(PROJECT_NAME)-amd64 workflow/$(PROJECT_NAME)-arm64

clean: ## Remove previous build
	@rm -f workflow/$(PROJECT_NAME)-amd64 workflow/$(PROJECT_NAME)-arm64

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

package-alfred: build
	@cd ./workflow \
	&& zip -r ../$(PROJECT_NAME).alfredworkflow ./* \
	&& cd .. && rm -rf workflow && git checkout workflow

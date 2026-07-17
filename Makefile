PROJECT_NAME := "alfred-lastpass-search"
PKG          := "github.com/rwilgaard/$(PROJECT_NAME)"
GO111MODULE  = on
VERSION      := $(shell plutil -extract version raw -o - workflow/info.plist)

.EXPORT_ALL_VARIABLES:
.PHONY: all dep fmt lint vet build clean universal-binary package-alfred zip-alfred release help

all: build

dep: ## Get the dependencies
	@go mod download

fmt: ## Format Go files with gofumpt
	@gofumpt -l -w ./src

lint: ## Lint Golang files
	@golangci-lint run --timeout 3m

vet: ## Run go vet
	@go vet ./src

build: dep ## Build arch-specific binaries
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o workflow/$(PROJECT_NAME)-amd64 ./src
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o workflow/$(PROJECT_NAME)-arm64 ./src

universal-binary: ## Combine arch binaries into universal binary
	@lipo -create -output workflow/$(PROJECT_NAME) workflow/$(PROJECT_NAME)-amd64 workflow/$(PROJECT_NAME)-arm64
	@rm -f workflow/$(PROJECT_NAME)-amd64 workflow/$(PROJECT_NAME)-arm64

clean: ## Remove build artifacts
	@rm -f workflow/$(PROJECT_NAME) workflow/$(PROJECT_NAME)-amd64 workflow/$(PROJECT_NAME)-arm64

package-alfred: build universal-binary ## Build and package into .alfredworkflow
	@cd ./workflow && zip -r ../$(PROJECT_NAME).alfredworkflow ./*
	@rm -f workflow/$(PROJECT_NAME)
	@echo "Created $(PROJECT_NAME).alfredworkflow"

zip-alfred: ## Package the workflow directory into .alfredworkflow without rebuilding
	@cd ./workflow && zip -r ../$(PROJECT_NAME).alfredworkflow ./*
	@echo "Created $(PROJECT_NAME).alfredworkflow"

release: ## Prepare and tag a new release (usage: make release V=x.y.z)
	@test -n "$(V)" || { echo "usage: make release V=x.y.z (current: $(VERSION))"; exit 1; }
	@echo "$(V)" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$$' || { echo "invalid version: $(V)"; exit 1; }
	@! git rev-parse -q --verify "refs/tags/v$(V)" >/dev/null || { echo "tag v$(V) already exists"; exit 1; }
	@git diff-index --quiet HEAD || { echo "uncommitted changes — commit or stash first"; exit 1; }
	@plutil -replace version -string "$(V)" workflow/info.plist
	@make package-alfred
	@git add workflow/info.plist
	@git commit -m "chore: release v$(V)"
	@git tag "v$(V)"
	@echo "tagged v$(V) — push with: git push origin main v$(V)"

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

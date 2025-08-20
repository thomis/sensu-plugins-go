BUILDOPT := -ldflags '-s -w'
SOURCES  := $(shell find cmd -type f -name "*.go" -not -name "*_test.go")
# Fix: Properly find non-oracle Go files, excluding test files
SOURCES_NO_ORACLE := $(shell find cmd -type f -name "*.go" -not -path "*oracle*" -not -name "*_test.go")
BINARIES := $(wildcard bin/*)
GREEN := \033[32m
RESET := \033[0m

.DEFAULT_GOAL := build

.PHONY: help build outdated update test format lint vul tools_install cover build_linux_amd64 build_linux_arm64 build_darwin_arm64 build_all clean

# Default target - build for local platform
default: build

help:
	@echo "Available targets (default: build):"
	@echo "  build             - Format, test, lint and build for local platform (default)"
	@echo "  help              - Show this help message"
	@echo "  test              - Run tests with coverage report"
	@echo "  format            - Format Go code"
	@echo "  lint              - Run static analysis"
	@echo "  vul               - Check for vulnerabilities"
	@echo "  cover             - Open HTML coverage report"
	@echo "  outdated          - List outdated dependencies"
	@echo "  update            - Update dependencies"
	@echo "  build_linux_amd64 - Build for Linux AMD64"
	@echo "  build_linux_arm64 - Build for Linux ARM64"
	@echo "  build_darwin_arm64- Build for macOS ARM64"
	@echo "  build_all         - Build for all platforms"
	@echo "  clean             - Clean bin and releases directories"
	@echo "  tools_install     - Install staticcheck and govulncheck"

build: format test lint vul
	@echo "\nBuilding for local platform..."
	@echo "-------------------------------"
	@mkdir -p bin
	@$(foreach FILE, $(SOURCES), echo $(FILE); go mod tidy && go build -o bin/`basename $(FILE) .go` $(FILE);)

outdated:
	@echo "$(GREEN)List outdated direct-dependencies$(RESET)"
	@go list -u -m -f '{{if and .Update (not .Indirect)}}{{.}}{{end}}' all

update:
	@echo "$(GREEN)Update direct-dependencies$(RESET)"
	@go get -u ./...

test:
	@echo "\nRunning tests with coverage..."
	@echo "------------------------------"
	@go test -coverprofile=coverage.out ./... 2>&1 | tee test.log
	@echo "\nGenerating HTML coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "\nOverall test coverage:"
	@go tool cover -func=coverage.out | grep total | awk '{print $$3}'
	@echo "\nDetailed coverage saved to coverage.html"

format:
	@echo "\nAbout to format..."
	@echo "---------------------"
	go fmt ./...

lint:
		@echo "\nAbout to lint..."
		@echo "----------------"
		staticcheck ./...

tools_install:
	@echo "\nInstalling staticcheck and govulncheck..."
	@echo "-----------------------------------------"
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	asdf reshim golang

vul:
	@echo "\nAbout to check for vulnerabilities..."
	@echo "--------------------------------------"
	govulncheck ./...

cover:
	@echo "\nOpening HTML coverage report..."
	@echo "--------------------------------"
	@if [ -f coverage.html ]; then \
		open coverage.html 2>/dev/null || xdg-open coverage.html 2>/dev/null || echo "Please open coverage.html in your browser"; \
	else \
		echo "No coverage report found. Run 'make test' first to generate coverage."; \
	fi

build_linux_amd64:
	@echo "\nbuilding for linux.amd64..."
	@echo "---------------------------"
	@mkdir -p bin releases
	@$(foreach FILE, $(SOURCES), echo $(FILE); \
		GOOS=linux GOARCH=amd64 go build $(BUILDOPT) -o bin/`basename $(FILE) .go` $(FILE);)
	tar cvf - bin/* | gzip > releases/sensu-checks-go.linux.amd64.tar.gz
	(cd releases && sha512sum sensu-checks-go.linux.amd64.tar.gz > sensu-checks-go.linux.amd64.tar.gz.sha512)

build_linux_arm64:
	@echo "\nbuilding for linux.arm64..."
	@echo "---------------------------"
	@mkdir -p bin releases
	@$(foreach FILE, $(SOURCES_NO_ORACLE), echo $(FILE); \
		GOOS=linux GOARCH=arm64 go build $(BUILDOPT) -o bin/`basename $(FILE) .go` $(FILE);)
	tar cvf - bin/* | gzip > releases/sensu-checks-go.linux.arm64.tar.gz
	(cd releases && sha512sum sensu-checks-go.linux.arm64.tar.gz > sensu-checks-go.linux.arm64.tar.gz.sha512)

build_darwin_arm64:
	@echo "\nbuilding for darwin.arm64..."
	@echo "---------------------------"
	@mkdir -p bin releases
	@$(foreach FILE, $(SOURCES), echo $(FILE); \
		GOOS=darwin GOARCH=arm64 go build $(BUILDOPT) -o bin/`basename $(FILE) .go` $(FILE);)
	tar cvf - bin/* | gzip > releases/sensu-checks-go.darwin.arm64.tar.gz
	(cd releases && sha512sum sensu-checks-go.darwin.arm64.tar.gz > sensu-checks-go.darwin.arm64.tar.gz.sha512)

build_all: clean format test lint vul build_linux_amd64 build_linux_arm64 build_darwin_arm64

clean:
	@echo "\nCleaning bin, releases, and coverage files..."
	@echo "----------------------------------------------"
	rm -f bin/*
	rm -r -f releases
	rm -f coverage.out coverage.html test.log c.out
	mkdir -p releases

BUILDOPT := -ldflags '-s -w'
SOURCES  := $(shell find cmd -type f -name "*.go" -not -name "*_test.go")
# Fix: Properly find non-oracle Go files, excluding test files
SOURCES_NO_ORACLE := $(shell find cmd -type f -name "*.go" -not -path "*oracle*" -not -name "*_test.go")
BINARIES := $(wildcard bin/*)
GREEN := \033[32m
RESET := \033[0m

.PHONY: build
build: clean_bin format lint vul
	@echo "\nBuilding local..."
	@echo "-----------------"
	@$(foreach FILE, $(SOURCES), echo $(FILE); go mod tidy && go build -o bin/`basename $(FILE) .go` $(FILE);)

.PHONY: outdated
outdated:
	@echo "$(GREEN)List outdated direct-dependencies$(RESET)"
	@go list -u -m -f '{{if and .Update (not .Indirect)}}{{.}}{{end}}' all

.PHONY: update
update:
	@echo "$(GREEN)Update direct-dependencies$(RESET)"
	@go get -u ./...

.PHONY: test
test:
	@echo "\nAbout to test..."
	@echo "----------------"
	go test -coverprofile=c.out ./...

.PHONY: format
format:
	@echo "\nAbout to format..."
	@echo "---------------------"
	go fmt ./...

.PHONY: lint
lint:
		@echo "\nAbout to lint..."
		@echo "----------------"
		staticcheck ./...

.PHONY: lint_install
lint_install:
	@echo "\nInstalling staticcheck..."
	@echo "-------------------------"
	go install honnef.co/go/tools/cmd/staticcheck@latest && asdf reshim golang

.PHONY: vul
vul:
	@echo "\nAbout to check for vulnerabilities..."
	@echo "--------------------------------------"
	govulncheck ./...

.PHONY: vul_install
vul_install:
	@echo "\nInstalling govulncheck..."
	@echo "-------------------------"
	go install golang.org/x/vuln/cmd/govulncheck@latest && asdf reshim golang

.PHONY: cover
cover:
	@echo "\nAbout to generate test coverage..."
	@echo "------------------------------------"
	go tool cover -html="c.out"

.PHONY: build_linux_amd64
build_linux_amd64: clean_bin
	@echo "\nbuilding for linux.amd64..."
	@echo "---------------------------"
	@$(foreach FILE, $(SOURCES), echo $(FILE); \
		GOOS=linux GOARCH=amd64 go build $(BUILDOPT) -o bin/`basename $(FILE) .go` $(FILE);)
	tar cvf - bin/* | gzip > releases/sensu-checks-go.linux.amd64.tar.gz
	(cd releases && sha512sum sensu-checks-go.linux.amd64.tar.gz > sensu-checks-go.linux.amd64.tar.gz.sha512)

.PHONY: build_linux_arm64
build_linux_arm64: clean_bin
	@echo "\nbuilding for linux.arm64..."
	@echo "---------------------------"
	@$(foreach FILE, $(SOURCES_NO_ORACLE), echo $(FILE); \
		GOOS=linux GOARCH=arm64 go build $(BUILDOPT) -o bin/`basename $(FILE) .go` $(FILE);)
	tar cvf - bin/* | gzip > releases/sensu-checks-go.linux.arm64.tar.gz
	(cd releases && sha512sum sensu-checks-go.linux.arm64.tar.gz > sensu-checks-go.linux.arm64.tar.gz.sha512)

.PHONY: build_darwin_amd64
build_darwin_amd64: clean_bin
	@echo "\nbuilding for darwin.amd64..."
	@echo "---------------------------"
	@$(foreach FILE, $(SOURCES), echo $(FILE); \
		GOOS=darwin GOARCH=amd64 go build $(BUILDOPT) -o bin/`basename $(FILE) .go` $(FILE);)
	tar cvf - bin/* | gzip > releases/sensu-checks-go.darwin.amd64.tar.gz
	(cd releases && sha512sum sensu-checks-go.darwin.amd64.tar.gz > sensu-checks-go.darwin.amd64.tar.gz.sha512)

.PHONY: build_darwin_arm64
build_darwin_arm64: clean_bin
	@echo "\nbuilding for darwin.arm64..."
	@echo "---------------------------"
	@$(foreach FILE, $(SOURCES_NO_ORACLE), echo $(FILE); \
		GOOS=darwin GOARCH=arm64 go build $(BUILDOPT) -o bin/`basename $(FILE) .go` $(FILE);)
	tar cvf - bin/* | gzip > releases/sensu-checks-go.darwin.arm64.tar.gz
	(cd releases && sha512sum sensu-checks-go.darwin.arm64.tar.gz > sensu-checks-go.darwin.arm64.tar.gz.sha512)

.PHONY: build_all
build_all: clean_release format test lint vul build_linux_amd64 build_linux_arm64 build_darwin_amd64 build_darwin_arm64

.PHONY: clean_bin
clean_bin:
	@echo "\nCleaning bin..."
	@echo "---------------"
	rm -f bin/*

.PHONY: clean_release
clean_release:
	@echo "\nCleaning releases..."
	@echo "--------------------"
	rm -r -f releases
	mkdir -p releases

BUILDOPT := -ldflags '-s -w'
SOURCES  := $(wildcard cmd/*/*.go)
# there is currently no instant client for darwin arm64, therefore we exclude oracle based files
SOURCES_NO_ORACLE := $(wildcard $(shell find cmd -type f -name "*.go" -not -path "*oracle*"))
BINARIES := $(wildcard bin/*)

build: clean_bin
	@echo "\nBuilding local..."
	@echo "-----------------"
	@$(foreach FILE, $(SOURCES), echo $(FILE); go build -o bin/`basename $(FILE) .go` $(FILE);)

test:
	@echo "\nAbout to test..."
	@echo "----------------"
	go test -coverprofile=c.out ./...

format:
	@echo "\nAbout to format..."
	@echo "---------------------"
	go fmt ./...

lint:
		@echo "\nAbout to lint..."
		@echo "----------------"
		@if ! command -v staticcheck &> /dev/null; then \
			echo "staticcheck not found, installing..."; \
			go install honnef.co/go/tools/cmd/staticcheck@latest && asdf reshim golang; \
		fi
		staticcheck ./...

vul:
	@echo "\nAbout to check for vulnerabilities..."
	@echo "--------------------------------------"
	@if ! command -v govulncheck &> /dev/null; then \
		echo "govulncheck not found, installing..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest && asdf reshim golang; \
	fi
	govulncheck ./...

cover:
	@echo "\nAbout to generate test coverage..."
	@echo "------------------------------------"
	go tool cover -html="c.out"

build_linux_amd64: clean_bin
	@echo "\nbuilding for linux.amd64..."
	@echo "---------------------------"
	@$(foreach FILE, $(SOURCES), echo $(FILE); \
		GOOS=linux GOARCH=amd64 go build $(BUILDOPT) -o bin/`basename $(FILE) .go` $(FILE);)
	tar cvf - bin/* | gzip > releases/sensu-checks-go.linux.amd64.tar.gz
	(cd releases && sha512sum sensu-checks-go.linux.amd64.tar.gz > sensu-checks-go.linux.amd64.tar.gz.sha512)

build_linux_arm64: clean_bin
	@echo "\nbuilding for linux.arm64..."
	@echo "---------------------------"
	@$(foreach FILE, $(SOURCES_NO_ORACLE), echo $(FILE); \
		GOOS=linux GOARCH=arm64 go build $(BUILDOPT) -o bin/`basename $(FILE) .go` $(FILE);)
	tar cvf - bin/* | gzip > releases/sensu-checks-go.linux.arm64.tar.gz
	(cd releases && sha512sum sensu-checks-go.linux.arm64.tar.gz > sensu-checks-go.linux.arm64.tar.gz.sha512)

build_darwin_amd64: clean_bin
	@echo "\nbuilding for darwin.amd64..."
	@echo "---------------------------"
	@$(foreach FILE, $(SOURCES), echo $(FILE); \
		GOOS=darwin GOARCH=amd64 go build $(BUILDOPT) -o bin/`basename $(FILE) .go` $(FILE);)
	tar cvf - bin/* | gzip > releases/sensu-checks-go.darwin.amd64.tar.gz
	(cd releases && sha512sum sensu-checks-go.darwin.amd64.tar.gz > sensu-checks-go.darwin.amd64.tar.gz.sha512)

build_darwin_arm64: clean_bin
	@echo "\nbuilding for darwin.arm64..."
	@echo "---------------------------"
	@$(foreach FILE, $(SOURCES_NO_ORACLE), echo $(FILE); \
		GOOS=darwin GOARCH=arm64 go build $(BUILDOPT) -o bin/`basename $(FILE) .go` $(FILE);)
	tar cvf - bin/* | gzip > releases/sensu-checks-go.darwin.arm64.tar.gz
	(cd releases && sha512sum sensu-checks-go.darwin.arm64.tar.gz > sensu-checks-go.darwin.arm64.tar.gz.sha512)

build_all: clean_release format test lint vul build_linux_amd64 build_linux_arm64 build_darwin_amd64 build_darwin_arm64

clean_bin:
	@echo "\nCleaning bin..."
	@echo "---------------"
	rm -f bin/*

clean_release:
	@echo "\nCleaning releases..."
	@echo "--------------------"
	rm -r -f releases
	mkdir -p releases

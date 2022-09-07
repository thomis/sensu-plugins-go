BUILDOPT := -ldflags '-s -w'
SOURCES  := $(wildcard */*.go)
# there is currently no instant client for darwin arm64
SOURCES_NO_ORACLE := $(filter-out $(wildcard */*oracle*), $(SOURCES))
BINARIES := $(wildcard bin/*)

build: clean_bin
	@echo "\nBuilding local..."
	@echo "-----------------"
	@$(foreach FILE, $(SOURCES), echo $(FILE); go build -o bin/`basename $(FILE) .go` $(FILE);)

test:
	@echo "\nAbout to test..."
	@echo "----------------"
	@$(foreach FILE, $(SOURCES), echo $(FILE); go test $(FILE);)

format:
	@echo "\nAbout to format..."
	@echo "---------------------"
	@$(foreach FILE, $(SOURCES), echo $(FILE); go fmt $(FILE) -e;)

lint:
	@echo "\nAbout to lint..."
	@echo "----------------"
	@$(foreach FILE, $(SOURCES), echo $(FILE); staticcheck $(FILE);)

vul:
	@echo "\nAbout to check for vulnerabilities..."
	@echo "--------------------------------------"
	@$(foreach FILE, $(BINARIES), echo $(FILE); govulncheck $(FILE);)

build_linux_amd64: clean_bin
	@echo "\nbuilding for linux.amd64..."
	@echo "---------------------------"
	@$(foreach FILE, $(SOURCES), echo $(FILE); \
		GOOS=linux GOARCH=amd64 go build $(BUILDOPT) -o bin/`basename $(FILE) .go` $(FILE);)
	tar cvf - bin/* | gzip > releases/sensu-checks-go.linux.amd64.tar.gz
	(cd releases && sha512sum sensu-checks-go.linux.amd64.tar.gz > sensu-checks-go.linux.amd64.tar.gz.sha512)

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

build_all: clean_release format test lint build_linux_amd64 build_darwin_amd64 build_darwin_arm64

clean_bin:
	@echo "\nCleaning bin..."
	@echo "---------------"
	rm -f bin/*

clean_release:
	@echo "\nCleaning releases..."
	@echo "--------------------"
	rm -r -f releases
	mkdir -p releases

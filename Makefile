BUILDOPT := -ldflags '-s -w'
SOURCES  := $(wildcard */*.go)

export GOOS=linux
export GOARCH=amd64

build:
	@$(foreach FILE, $(SOURCES), echo $(FILE); go build $(BUILDOPT) -o bin/`basename $(FILE) .go` $(FILE);)

clean:
	rm -f bin/*
	rm sensu-checks-go*

asset:
	tar cvf - bin/* | gzip > sensu-checks-go.linux.amd64.tar.gz
	sha512sum sensu-checks-go.linux.amd64.tar.gz > sensu-checks-go.linux.amd64.tar.gz.sha512

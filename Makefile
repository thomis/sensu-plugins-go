BUILDOPT := -ldflags '-s -w'
SOURCES  := $(wildcard */*.go)

build:
	@$(foreach FILE, $(SOURCES), echo $(FILE); go build $(BUILDOPT) -o bin/`basename $(FILE) .go` $(FILE);)

clean:
	rm -f bin/*
	rm sensu-checks-go*

asset:
	tar cvf - bin/* | gzip > releases/sensu-checks-go.linux.amd64.tar.gz
	sha512sum releases/sensu-checks-go.linux.amd64.tar.gz > releases/sensu-checks-go.linux.amd64.tar.gz.sha512

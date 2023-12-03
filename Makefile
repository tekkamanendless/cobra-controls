all: cobra-cli

ALL_GO_FILES=$(shell find ./ -iname '*.go' -type f)

.PHONY: cobra-cli
cobra-cli: build/cobra-cli build/cobra-cli.exe

build:
	mkdir -p build

.PHONY: clean
clean:
	rm -rf build

build/cobra-cli: build $(ALL_GO_FILES)
	CGO_ENABLED=0 GOOS=linux go build -o $@ ./cmd/cobra-cli/*.go

build/cobra-cli.exe: build $(ALL_GO_FILES)
	CGO_ENABLED=0 GOOS=windows go build -o $@ ./cmd/cobra-cli/*.go

all: cobra-cli

.PHONY: cobra-cli
cobra-cli: build/cobra-cli build/cobra-cli.exe

build:
	mkdir -p build

.PHONY: clean
clean:
	rm -rf build

build/cobra-cli: build
	CGO_ENABLED=0 GOOS=linux go build -o $@ ./cmd/cobra-cli/*.go

build/cobra-cli.exe: build
	CGO_ENABLED=0 GOOS=windows go build -o $@ ./cmd/cobra-cli/*.go

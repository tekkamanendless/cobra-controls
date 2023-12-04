all: cobra-cli

ALL_GO_FILES=$(shell find ./ -iname '*.go' -type f)

.PHONY: cobra-cli
cobra-cli: bin/cobra-cli bin/cobra-cli.exe

bin:
	mkdir -p bin

.PHONY: clean
clean:
	rm -rf bin

bin/cobra-cli: bin $(ALL_GO_FILES)
	CGO_ENABLED=0 GOOS=linux go build -o $@ ./cmd/cobra-cli/*.go

bin/cobra-cli.exe: bin $(ALL_GO_FILES)
	CGO_ENABLED=0 GOOS=windows go build -o $@ ./cmd/cobra-cli/*.go

.PHONY: test
test:
	go vet ./...
	go test ./...


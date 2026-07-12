.PHONY: build test lint clean install

BINARY   := vhoster
GO       ?= go
GOFLAGS  ?= -ldflags="-s -w"

build:
	$(GO) build $(GOFLAGS) -o $(BINARY) .

test:
	$(GO) test -count=1 -v ./...

vet:
	$(GO) vet ./...

lint: vet
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"

clean:
	rm -f $(BINARY) $(BINARY).exe

install: build
	sudo cp $(BINARY) /usr/local/bin/

.PHONY: build static tiny test lint vendor clean

VERSION := $(shell git describe --tags --always 2>/dev/null || echo "dev")
LDFLAGS := -X main.version=$(VERSION)

build:
	go build -ldflags="$(LDFLAGS)" -o chooz .

static:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w $(LDFLAGS)" -o chooz .

tiny: static
	upx --best chooz

test:
	go test ./...

lint:
	golangci-lint run

vendor:
	go mod vendor

clean:
	rm -f chooz

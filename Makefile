.PHONY: build

LDFLAGS = -X main.version=$(shell git describe --tags --always --dirty) -X main.commit=$(shell git rev-parse HEAD)

build:
	go build -ldflags "$(LDFLAGS)" -o ./build/gestic .

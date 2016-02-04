.PHONY: all build test static container push clean

all: build

build:
	@echo "Building tmpl-renderer..."
	ROOTPATH=$(shell pwd -P); \
	GO15VENDOREXPERIMENT=1 go build -o $$ROOTPATH/bin/tmpl-renderer

test:
	@echo "Running tests..."
	GO15VENDOREXPERIMENT=1 go test

static:
	ROOTPATH=$(shell pwd -P); \
	mkdir -p $$ROOTPATH/bin; \
	GO15VENDOREXPERIMENT=1 CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build \
		-a -tags netgo -installsuffix cgo -ldflags '-extld ld -extldflags -static' -a -x \
		-o $$ROOTPATH/bin/tmpl-renderer-linux-amd64 \
		. \
	; \
	ROOTPATH=$(shell pwd -P); \
	GO15VENDOREXPERIMENT=1 CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 \
	go build \
		-a -tags netgo -installsuffix cgo -ldflags '-extld ld -extldflags -static' -a -x \
		-o $$ROOTPATH/bin/tmpl-renderer-darwin-amd64 \
		. \
	; \
	GO15VENDOREXPERIMENT=1 CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
	go build \
		-a -tags netgo -installsuffix cgo -ldflags '-extld ld -extldflags -static' -a -x \
		-o $$ROOTPATH/bin/tmpl-renderer-windows-amd64 \
		.

clean:
	rm -f bin/tmpl-renderer*

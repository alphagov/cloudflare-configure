.PHONY: deps test build

BINARY := cloudflare-configure

all: deps test build

deps:
	go get github.com/tools/godep
	godep restore

test: deps
	godep go test

build: deps
	godep go build -o $(BINARY)

clean:
	rm -rf $(BINARY)

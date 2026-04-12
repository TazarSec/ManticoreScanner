.PHONY: build test vet lint clean install

BINARY := manticore
PKG := ./cmd/manticore

build:
	go build -o $(BINARY) $(PKG)

test:
	go test ./...

vet:
	go vet ./...

lint: vet

clean:
	rm -f $(BINARY)

install:
	go install $(PKG)

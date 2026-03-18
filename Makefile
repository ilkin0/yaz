BINARY=yaz
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/yaz

run: build
	sudo $(BUILD_DIR)/$(BINARY)

test:
	go test -v ./...

clean:
	rm -rf $(BUILD_DIR)

.PHONY: build run test clean

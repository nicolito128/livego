BINARY = bin/livego
BINARY_EXE = bin\livego.exe

OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)

build:
	@if [ "$(OS)" = "windows" ]; then \
		go build -o $(BINARY_EXE) .; \
	else \
		go build -o $(BINARY) .; \
	fi

dev:
	make build
	@if [ "$(OS)" = "windows" ]; then \
		$(BINARY_EXE) -path ./examples/; \
	else \
		./$(BINARY) -path ./examples/; \
	fi

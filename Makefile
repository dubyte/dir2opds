.DEFAULT_GOAL := build
.PHONY: fmt vet build

NAME := dir2opds
SRCS := $(wildcard *.go) $(wildcard */*.go)

fmt: $(SRCS)
	go fmt ./...

vet: $(SRCS) fmt
	go vet ./...

build: $(SRCS) vet
	go build .

build-all: darwin freebsd illumos linux netbsd openbsd windows

darwin: bin/${NAME}-darwin-arm64
freebsd: bin/${NAME}-freebsd-amd64
illumos: bin/${NAME}-illumos-amd64
linux: bin/${NAME}-linux-amd64 bin/${NAME}-linux-arm64 bin/${NAME}-linux-armv7
netbsd: bin/${NAME}-netbsd-amd64
openbsd: bin/${NAME}-openbsd-amd64
windows: bin/${NAME}-windows-amd64.exe

bin/${NAME}-darwin-arm64: $(SRCS) vet
	@mkdir -p bin/darwin-arm64/
	@echo "Building darwin-arm64..."
	env GOOS=darwin GOARCH=arm64 go build -o bin/darwin-arm64/${NAME}

bin/${NAME}-freebsd-amd64: $(SRCS) vet
	@mkdir -p bin/freebsd-amd64/
	@echo "Building freebsd-amd64..."
	env GOOS=freebsd GOARCH=amd64 go build -o bin/freebsd-amd64/${NAME}

bin/${NAME}-illumos-amd64: $(SRCS) vet
	@mkdir -p bin/illumos-amd64/
	@echo "Building illumos-amd64..."
	env GOOS=illumos GOARCH=amd64 go build -o bin/illumos-amd64/${NAME}

bin/${NAME}-linux-amd64: $(SRCS) vet
	@mkdir -p bin/linux-amd64/
	@echo "Building linux-amd64..."
	env GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/${NAME}

bin/${NAME}-linux-arm64: $(SRCS) vet
	@mkdir -p bin/linux-arm64/
	@echo "Building linux-arm64..."
	env GOOS=linux GOARCH=arm64 go build -o bin/linux-arm64/${NAME}

bin/${NAME}-linux-armv7: $(SRCS) vet
	@mkdir -p bin/linux-armv7/
	@echo "Building linux-armv7..."
	env GOOS=linux GOARCH=arm GOARM=7 go build -o bin/linux-armv7/${NAME}

bin/${NAME}-netbsd-amd64: $(SRCS) vet
	@mkdir -p bin/netbsd-amd64/
	@echo "Building netbsd-amd64..."
	env GOOS=netbsd GOARCH=amd64 go build -o bin/netbsd-amd64/${NAME}

bin/${NAME}-openbsd-amd64: $(SRCS) vet
	@mkdir -p bin/openbsd-amd64/
	@echo "Building openbsd-amd64..."
	env GOOS=openbsd GOARCH=amd64 go build -o bin/openbsd-amd64/${NAME}

bin/${NAME}-windows-amd64.exe: $(SRCS) vet
	@mkdir -p bin/windows-amd64/
	@echo "Building windows-amd64..."
	env GOOS=windows GOARCH=amd64 go build -o bin/windows-amd64/${NAME}.exe

clean:
	go clean
	rm -r bin/

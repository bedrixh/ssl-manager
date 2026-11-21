
.PHONY: compile build install uninstall clean run

RM       = rm -f
PREFIX   ?= /usr/local
GO_LDFLAGS := -ldflags "-X main.Version=$$(git describe --tags --exclude release) -X main.Commit=$$(git rev-parse --short HEAD) -X main.BuildTime=$$(date -u +%Y-%m-%dT%H:%M:%SZ)"
BIN_NAME ?= ssl-manager

compile:
	@echo "Compiling for every OS and Platform"
	GOOS=linux   GOARCH=amd64 go build $(GO_LDFLAGS) -o bin/$(BIN_NAME)-linux-amd64 .
	GOOS=linux   GOARCH=arm   go build $(GO_LDFLAGS) -o bin/$(BIN_NAME)-linux-arm   .
	GOOS=linux   GOARCH=arm64 go build $(GO_LDFLAGS) -o bin/$(BIN_NAME)-linux-arm64 .
	GOOS=freebsd GOARCH=386   go build $(GO_LDFLAGS) -o bin/$(BIN_NAME)-freebsd-386 .

build:
	@echo "Building for your architecture"
	go build $(GO_LDFLAGS) -o bin/$(BIN_NAME) .

install: build
	@echo Installing built package
	install -m 755 bin/${BIN_NAME} ${DESTDIR}${PREFIX}/bin/$(BIN_NAME)

uninstall: 
	@echo Removing binaries
	${RM} ${DESTDIR}${PREFIX}/bin/$(BIN_NAME)

clean:
	@echo Cleaning build targets
	${RM} bin/$(BIN_NAME) 

run: build
	@echo Executing binary
	bin/$(BIN_NAME)

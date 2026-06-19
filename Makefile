GO_LDFLAGS := -ldflags "-X main.Version=$$(git describe --tags) -X main.Commit=$$(git rev-parse --short HEAD) -X main.BuildTime=$$(date -u +%Y-%m-%dT%H:%M:%SZ)"

compile:
	@echo "Compiling for every OS and Platform"
	GOOS=linux   GOAECH=amd64 go build $(GO_LDFLAGS) -o bin/ssl-manager-linux-amd64 .
	GOOS=linux   GOARCH=arm   go build $(GO_LDFLAGS) -o bin/ssl-manager-linux-arm   .
	GOOS=linux   GOARCH=arm64 go build $(GO_LDFLAGS) -o bin/ssl-manager-linux-arm64 .
	GOOS=freebsd GOARCH=386   go build $(GO_LDFLAGS) -o bin/ssl-manager-freebsd-386 .
	GOOS=windows GOARCH=amd64 go build $(GO_LDFLAGS) -o bin/ssl-manager-windows-amd64.exe .

build:
	@echo "Building for your architecture"
	go build -o bin/ssl-manager .

install: build
	@echo Installing builded package
	cp bin/ssl-manager /usr/bin/ssl-manager

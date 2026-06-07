compile:
	echo "Compiling for every OS and Platform"
	GOOS=linux   GOAECH=amd64 go build -o bin/ssl-manager-linux-amd64 .
	GOOS=linux   GOARCH=arm   go build -o bin/ssl-manager-linux-arm   .
	GOOS=linux   GOARCH=arm64 go build -o bin/ssl-manager-linux-arm64 .
	GOOS=freebsd GOARCH=386   go build -o bin/ssl-manager-freebsd-386 .
	GOOS=windows GOARCH=amd64 go build -o bin/ssl-manager-windows-amd64.exe .

build:
	echo "Building for your computer"
	go build -o bin/ssl-manager .

install: build
	echo Installing builded package
	cp bin/ssl-manager /bin/ssl-manager

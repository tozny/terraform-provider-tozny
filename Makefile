BINARY=terraform-provider-tozny
VERSION=0.0.1

default: install

all: lint install release

lint:
	go vet ./...
	go mod tidy

build:
	go build -o ${BINARY}

install: build
	mv ${BINARY} ~/.terraform.d/plugins

install-mac: build
	# Build and move binary to implicit local file system location for a tozny third party provider
	# https://www.terraform.io/docs/configuration/provider-requirements.html#in-house-providers
	# https://www.terraform.io/upgrade-guides/0-13.html#in-house-providers
	# https://www.terraform.io/docs/commands/cli-config.html#implied-local-mirror-directories
	mkdir -p ~/Library/Application\ Support/io.terraform/plugins/terraform.tozny.com/tozny/tozny/${VERSION}/darwin_amd64/
	mv ${BINARY} ~/Library/Application\ Support/io.terraform/plugins/terraform.tozny.com/tozny/tozny/${VERSION}/darwin_amd64/${BINARY}_v${VERSION}

clean-mac:
	rm -rf ~/Library/Application\ Support/io.terraform/plugins/terraform.tozny.com/tozny/tozny/

release:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_v${VERSION}_darwin_amd64
	GOOS=freebsd GOARCH=386 go build -o ./bin/${BINARY}_v${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -o ./bin/${BINARY}_v${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm go build -o ./bin/${BINARY}_v${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 go build -o ./bin/${BINARY}_v${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY}_v${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build -o ./bin/${BINARY}_v${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386 go build -o ./bin/${BINARY}_v${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -o ./bin/${BINARY}_v${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -o ./bin/${BINARY}_v${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 go build -o ./bin/${BINARY}_v${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -o ./bin/${BINARY}_v${VERSION}_windows_amd64
	chmod +x ./bin/*

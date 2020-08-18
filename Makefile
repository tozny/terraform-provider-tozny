BINARY=terraform-provider-tozny
VERSION=0.0.1

default: build

all: lint install

lint:
	go vet ./...
	go mod tidy

build:
	go build -o ${BINARY}

install: build
	mv ${BINARY} ~/.terraform.d/plugins/${BINARY}_v${VERSION}

clean:
	rm ~/.terraform.d/plugins/${BINARY}_v${VERSION} || true

install-mac: build
	# Build and move binary to implicit local file system location for a tozny third party provider
	# https://www.terraform.io/docs/configuration/provider-requirements.html#in-house-providers
	# https://www.terraform.io/upgrade-guides/0-13.html#in-house-providers
	# https://www.terraform.io/docs/commands/cli-config.html#implied-local-mirror-directories
	mkdir -p ~/Library/Application\ Support/io.terraform/plugins/terraform.tozny.com/tozny/tozny/${VERSION}/darwin_amd64/
	mv ${BINARY} ~/Library/Application\ Support/io.terraform/plugins/terraform.tozny.com/tozny/tozny/${VERSION}/darwin_amd64/${BINARY}_v${VERSION}

clean-mac:
	rm -rf ~/Library/Application\ Support/io.terraform/plugins/terraform.tozny.com/tozny/tozny/ || true

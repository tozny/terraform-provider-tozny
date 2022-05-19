BINARY=terraform-provider-tozny
VERSION=0.21.0

default: build

all: lint install

lint:
	go vet ./...
	go mod tidy
	go fmt ./...

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
	mkdir -p ~/Library/Application\ Support/io.terraform/plugins/registry.terraform.io/tozny/tozny/${VERSION}/darwin_amd64/
	mv ${BINARY} ~/Library/Application\ Support/io.terraform/plugins/registry.terraform.io/tozny/tozny/${VERSION}/darwin_amd64/${BINARY}_v${VERSION}

clean-mac:
	rm -rf ~/Library/Application\ Support/io.terraform/plugins/registry.terraform.io/tozny/tozny/ || true

version:
	git tag v${VERSION}
	git push origin v${VERSION}

release:
	goreleaser release --rm-dist

test:
	./examples/realms/applications/roles/test.sh
	./examples/realms/groups/test.sh

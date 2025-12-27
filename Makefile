.PHONY: build test testacc generate docs install lint docker-up docker-down clean

HOSTNAME=registry.terraform.io
NAMESPACE=dklesev
NAME=pihole
BINARY=terraform-provider-${NAME}
VERSION=0.1.0

# Detect OS and architecture
OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)
OS_ARCH=${OS}_${ARCH}

default: build

build:
	go build -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	go test -v -cover -timeout 30s ./...

testacc:
	TF_ACC=1 go test -v -timeout 30m ./internal/provider/...

generate:
	go generate ./...

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate -provider-name pihole

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...
	gofumpt -l -w .

docker-up:
	docker-compose up -d
	@echo "Waiting for Pi-hole to be ready..."
	@until curl -s http://localhost:8080/api/info/version > /dev/null 2>&1; do \
		sleep 2; \
	done
	@echo "Pi-hole is ready!"

docker-down:
	docker-compose down -v

docker-logs:
	docker-compose logs -f pihole

clean:
	rm -f ${BINARY}
	rm -rf ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}

# Run a quick integration test against local Pi-hole
quicktest: docker-up
	cd examples/provider && terraform init && terraform plan

# Development workflow: rebuild, reinstall, and test
dev: install
	cd examples/provider && rm -rf .terraform* && terraform init && terraform plan

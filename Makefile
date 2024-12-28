.PHONY: test test-short test-coverage test-integration test-local test-vault test-aws test-azure test-gcp start-deps stop-deps

# Default target
all: test

# Run all tests
test:
	go test ./...

# Run tests without integration tests
test-short:
	go test -short ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Run only integration tests
test-integration:
	go test -run Integration ./...

# Run tests for specific providers
test-local:
	go test ./internal/providers/local/...

test-vault:
	go test ./internal/providers/vault/...

test-aws:
	go test ./internal/providers/aws/...

test-azure:
	go test ./internal/providers/azure/...

test-gcp:
	go test ./internal/providers/gcp/...

# Start all test dependencies
start-deps:
	# Start Vault in dev mode
	vault server -dev & \
	# Start LocalStack for AWS
	docker run -d --rm -p 4566:4566 --name keeper-localstack localstack/localstack & \
	# Start Azurite for Azure
	docker run -d --rm -p 10000:10000 -p 10001:10001 -p 10002:10002 --name keeper-azurite mcr.microsoft.com/azure-storage/azurite & \
	# Start GCP emulator
	docker run -d --rm -p 8085:8085 --name keeper-gcp gcr.io/google.com/cloudsdktool/cloud-sdk:latest gcloud beta emulators secretmanager start --host-port=0.0.0.0:8085

# Stop all test dependencies
stop-deps:
	-vault server -dev stop
	-docker stop keeper-localstack
	-docker stop keeper-azurite
	-docker stop keeper-gcp

# Clean up
clean:
	rm -f coverage.out
	$(MAKE) stop-deps

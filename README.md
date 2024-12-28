# Keeper (kpr) - Universal Secret Manager

Keeper is a powerful, unified command-line interface for managing secrets across multiple secret management systems. It provides a seamless experience for handling secrets in various cloud providers and local storage.

## Features

- 🔐 Unified interface for multiple secret backends
- 🚀 Support for major providers:
  - Local encrypted storage (AES-256-GCM)
  - HashiCorp Vault
  - AWS Secrets Manager
- 🔄 Secret versioning and rotation
- 📦 Bulk operations support
- 🔒 Local encrypted storage with master key
- 🎯 Interactive mode
- 🌈 Rich shell completion

## Installation

```bash
go install github.com/yourusername/keeper/cmd/kpr@latest
```

## Quick Start

```bash
# Initialize Keeper
kpr init

# Set a secret using local provider
kpr set myapp/db/password "mysecret"

# Get a secret
kpr get myapp/db/password

# List secrets
kpr list myapp/db

# Use with different providers
kpr --provider vault set myapp/api-key "secret-value"
kpr --provider aws get myapp/api-key
```

## Configuration

Keeper can be configured using either:
- Configuration file (`~/.keeper/config.yaml`)
- Environment variables
- Command-line flags

Example configuration:

```yaml
default_provider: "local"
default_data_dir: "~/.keeper"

providers:
  local:
    type: "local"
    parameters:
      path: "~/.keeper/secrets"
      key_file: "~/.keeper/master.key"

  vault:
    type: "vault"
    parameters:
      address: "http://localhost:8200"
      token: "hvs.example..."
      mount: "secret"

  aws:
    type: "aws"
    parameters:
      region: "us-west-2"
      prefix: "myapp/"
```

## Provider Configuration

### Local Provider
The local provider stores secrets in encrypted files on disk using AES-256-GCM encryption.

Required parameters:
- `path`: Directory to store secrets
- `key_file`: Path to the master key file

### HashiCorp Vault
The Vault provider integrates with HashiCorp Vault's KV v2 secret engine.

Required parameters:
- `address`: Vault server address
- `token`: Authentication token
- `mount`: Secret engine mount path (default: "secret")

### AWS Secrets Manager
The AWS provider integrates with AWS Secrets Manager.

Required parameters:
- `region`: AWS region
- `prefix`: Optional prefix for secret names

## Security

Keeper takes security seriously:
- All local secrets are encrypted using AES-256-GCM
- Strong authentication methods for each provider
- Audit logging for all operations
- Secure credential handling
- No secret logging to stdout/stderr

## Testing

### Running Tests

To run all tests:
```bash
go test ./...
```

To run tests for a specific provider:
```bash
go test ./internal/providers/local/...  # Local provider tests
go test ./internal/providers/vault/...  # Vault provider tests
go test ./internal/providers/aws/...    # AWS provider tests
go test ./internal/providers/azure/...  # Azure provider tests
go test ./internal/providers/gcp/...    # GCP provider tests
```

To skip integration tests:
```bash
go test -short ./...
```

### Setting Up Test Dependencies

#### Local Provider
No additional setup required. Tests use temporary directories.

#### HashiCorp Vault
1. Start a local Vault server in dev mode:
```bash
vault server -dev
```
2. Set environment variables:
```bash
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='dev-token'
```

#### AWS Secrets Manager
1. Start LocalStack:
```bash
docker run --rm -p 4566:4566 localstack/localstack
```
2. Set environment variables:
```bash
export AWS_ACCESS_KEY_ID='test'
export AWS_SECRET_ACCESS_KEY='test'
export AWS_DEFAULT_REGION='us-west-2'
```

#### Azure Key Vault
1. Start Azurite:
```bash
docker run --rm -p 10000:10000 -p 10001:10001 -p 10002:10002 mcr.microsoft.com/azure-storage/azurite
```
2. Set environment variables:
```bash
export AZURE_TENANT_ID='test'
export AZURE_CLIENT_ID='test'
export AZURE_CLIENT_SECRET='test'
```

#### Google Cloud Secret Manager
1. Start the GCP emulator:
```bash
docker run --rm -p 8085:8085 gcr.io/google.com/cloudsdktool/cloud-sdk:latest gcloud beta emulators secretmanager start --host-port=0.0.0.0:8085
```
2. Set environment variables:
```bash
export GOOGLE_CLOUD_PROJECT='test-project'
export GOOGLE_APPLICATION_CREDENTIALS='test'
```

### Test Coverage

To run tests with coverage:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # View coverage in browser
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

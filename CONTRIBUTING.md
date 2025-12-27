# Contributing to terraform-provider-pihole

Thank you for your interest in contributing to the Terraform Provider for Pi-hole!

## Development Environment

### Requirements

- [Go](https://golang.org/doc/install) >= 1.22
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0 or [OpenTofu](https://opentofu.org/) >= 1.6
- [Docker](https://docs.docker.com/get-docker/) (for running tests)
- [golangci-lint](https://golangci-lint.run/usage/install/) (for linting)

### Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/dklesev/terraform-provider-pihole.git
   cd terraform-provider-pihole
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the provider:
   ```bash
   make build
   ```

4. Install locally for testing:
   ```bash
   make install
   ```

## Testing

### Unit Tests

Unit tests don't require a Pi-hole instance and can be run quickly:

```bash
go test -v ./internal/client/...
```

### Acceptance Tests

Acceptance tests require a running Pi-hole instance. Use Docker:

```bash
# Start Pi-hole
make docker-up

# Run acceptance tests
make testacc

# Stop Pi-hole
make docker-down
```

**Note:** Acceptance tests create real resources. Always use a test Pi-hole instance.

### Test Requirements

- **All new resources** must have acceptance tests covering:
  - Basic create/read/destroy
  - All optional fields
  - Updates
  - Import

- **All bug fixes** should include a test that would fail without the fix

## Code Style

### Go

- Follow standard Go conventions
- Use `gofmt` and `gofumpt` for formatting
- Run `golangci-lint run` before committing
- Add godoc comments on exported types and functions

### Terraform Provider

- Follow [HashiCorp's provider design principles](https://developer.hashicorp.com/terraform/plugin/best-practices)
- Use meaningful attribute descriptions
- Include examples in resource/data source documentation

## Pull Request Process

1. **Fork** the repository and create a branch from `main`
2. **Make changes** following the code style guidelines
3. **Add/update tests** for your changes
4. **Update documentation** if needed
5. **Run all tests** locally:
   ```bash
   make test
   make lint
   make testacc
   ```
6. **Commit** using [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat:` new feature
   - `fix:` bug fix
   - `docs:` documentation changes
   - `test:` adding/updating tests
   - `chore:` maintenance tasks

7. **Submit PR** with a clear description of changes

## Reporting Issues

- Check existing issues before creating a new one
- Include Terraform/OpenTofu version, provider version, and Pi-hole version
- Provide minimal reproduction steps
- Include relevant configuration (sanitized)

## Security

Report security vulnerabilities privately via GitHub Security Advisories.

## Questions?

Open a [Discussion](https://github.com/dklesev/terraform-provider-pihole/discussions) for questions or ideas.

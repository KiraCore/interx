# Interx Integration Tests

This directory contains Go-based integration tests for the Interx API endpoints, designed to run in containers.

## Quick Start (Docker)

```bash
# Build and run all tests
make docker-test

# Run smoke tests only
make docker-smoke

# Run with custom Interx URL
INTERX_URL=http://your-server:11000 make docker-test
```

## Test Coverage

The test suite covers **60+ API endpoints** organized into the following categories:

| Category | Endpoints | Test File |
|----------|-----------|-----------|
| Account | 2 | `account_test.go` |
| Transactions | 9 | `transactions_test.go` |
| Validators | 5 | `validators_test.go` |
| Faucet | 2 | `faucet_test.go` |
| Proposals | 5 | `proposals_test.go` |
| Genesis | 4 | `genesis_test.go` |
| Data Reference | 2 | `data_reference_test.go` |
| Tokens | 2 | `tokens_test.go` |
| Node Discovery | 4 | `node_discovery_test.go` |
| Upgrade | 2 | `upgrade_test.go` |
| Identity Registrar | 7 | `identity_test.go` |
| Permissions/Roles | 3 | `permissions_roles_test.go` |
| Spending | 4 | `spending_test.go` |
| Staking | 3 | `staking_test.go` |
| Execution Fees | 2 | `execution_fee_test.go` |
| Status/Meta | 9 | `status_test.go` |

## Docker Commands

| Command | Description |
|---------|-------------|
| `make docker-build` | Build the test container |
| `make docker-test` | Run all tests in container |
| `make docker-smoke` | Run smoke tests (status endpoints) |
| `make docker-account` | Run account endpoint tests |
| `make docker-single TEST=TestName` | Run a specific test |
| `make docker-clean` | Remove containers and images |

### Using Docker Compose

```bash
# Run all integration tests
docker-compose up --build integration-tests

# Run smoke tests
docker-compose up --build smoke-tests

# Run account tests
docker-compose up --build account-tests
```

## Configuration

Tests can be configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `INTERX_URL` | `http://3.123.154.245:11000` | Interx server URL |
| `TEST_ADDRESS` | `kira143q8vxpvuykt9pq50e6hng9s38vmy844n8k9wx` | Test account address |
| `VALIDATOR_ADDRESS` | `kira1vvcj9avffvyav83gmptdlzrprgvsrjxzh7f9sz` | Validator address |
| `DELEGATOR_ADDRESS` | `kira177lwmjyjds3cy7trers83r4pjn3dhv8zrqk9dl` | Delegator address |

### Examples

```bash
# Test against chaosnet (default)
make docker-test

# Test against local instance
INTERX_URL=http://localhost:11000 make docker-test

# Test against custom server
INTERX_URL=http://your-server:11000 TEST_ADDRESS=kira1... make docker-test
```

## Local Development

For local development without Docker:

```bash
# Install dependencies
make deps

# Run tests locally
make test

# Run specific test
make test-single TEST=TestInterxStatus
```

## CI Integration

Integration tests run automatically via GitHub Actions on:
- Push to feature/*, hotfix/*, bugfix/*, release/*, major/* branches
- Pull requests to master
- Manual workflow dispatch (with custom URL)

See `.github/workflows/integration.yml` for the workflow configuration.

## Adding New Tests

1. Create a new test file following the naming convention `*_test.go`
2. Use the `GetConfig()` function to get the test configuration
3. Use `NewClient(cfg)` to create an HTTP client
4. Follow the existing test patterns for assertions

Example:
```go
func TestNewEndpoint(t *testing.T) {
    cfg := GetConfig()
    client := NewClient(cfg)

    t.Run("test case name", func(t *testing.T) {
        resp, err := client.Get("/api/new/endpoint", nil)
        require.NoError(t, err)
        assert.True(t, resp.IsSuccess())
    })
}
```

## Project Structure

```
tests/integration/
├── Dockerfile              # Container image for tests
├── docker-compose.yml      # Multi-service test configuration
├── Makefile                # Build and run commands
├── go.mod                  # Go module definition
├── config.go               # Environment configuration
├── client.go               # HTTP client utilities
├── account_test.go         # Account endpoint tests
├── transactions_test.go    # Transaction endpoint tests
├── validators_test.go      # Validator endpoint tests
├── faucet_test.go          # Faucet endpoint tests
├── proposals_test.go       # Governance proposal tests
├── genesis_test.go         # Genesis endpoint tests
├── data_reference_test.go  # Data reference tests
├── tokens_test.go          # Token endpoint tests
├── node_discovery_test.go  # Node discovery tests
├── upgrade_test.go         # Upgrade plan tests
├── identity_test.go        # Identity registrar tests
├── permissions_roles_test.go # Permissions/roles tests
├── spending_test.go        # Spending pool tests
├── staking_test.go         # Staking endpoint tests
├── execution_fee_test.go   # Execution fee tests
└── status_test.go          # Status/metadata tests
```

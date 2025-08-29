# Integration Tests for Harlequin Remote Signing Server

This document describes the comprehensive integration test suite for the remote signing server, which uses the goar library to test real-world Arweave data item signing workflows.

## Test Overview

The integration tests verify the complete end-to-end workflow of the remote signing service, including:

- Data submission and retrieval
- Arweave wallet integration via goar
- WebSocket real-time notifications
- Error handling and edge cases
- Concurrent request processing
- Performance benchmarking

## Test Files

- **`integration_test.go`** - Main integration test suite
- **`test-wallet.json`** - Test Arweave wallet for signing operations
- **`server/server_test.go`** - Unit tests for server components

## Test Suite Structure

### TestIntegrationSuite

The main test suite includes the following test cases:

#### 1. CompleteSigningWorkflow

Tests the full signing workflow:

- Creates an Arweave data item using goar
- Submits data to the remote signing server
- Retrieves unsigned data by UUID
- Simulates signing with the test wallet
- Submits signed data back to the server
- Verifies signature and wallet ownership

#### 2. JSONDataSubmission

Tests submitting data via JSON format:

- Submits data as JSON payload
- Verifies correct storage and retrieval
- Validates HTTP status codes (201 for creation, 200 for retrieval)

#### 3. BinaryDataSubmission

Tests submitting raw binary data:

- Generates random binary data (1KB)
- Submits via `application/octet-stream`
- Verifies binary data integrity

#### 4. WebSocketNotifications

Tests real-time notifications:

- Establishes WebSocket connection
- Subscribes to UUID-specific updates
- Submits signed data and verifies notification received
- Tests message format and timing

#### 5. ErrorHandling

Tests various error conditions:

- Invalid UUID format (400 Bad Request)
- Non-existent UUID (404 Not Found)
- Data too large (413 Request Entity Too Large)
- Invalid JSON format (400 Bad Request)

#### 6. MultipleRequests

Tests concurrent request handling:

- Submits 10 concurrent requests
- Verifies all requests succeed
- Tests data integrity under load

## Performance Benchmarks

The test suite includes benchmark tests:

### BenchmarkDataSubmission

- Tests data submission performance under load
- Uses parallel execution (`b.RunParallel`)
- Measures requests per second

### BenchmarkDataRetrieval

- Tests data retrieval performance
- Pre-creates test data for consistent measurement
- Measures response times

## Test Configuration

### Test Server

- **Host**: `localhost`
- **Port**: `8082` (different from default to avoid conflicts)
- **Timeout**: 30 seconds
- **Max Data Size**: 10MB

### Test Wallet

The test wallet (`test-wallet.json`) contains:

- RSA key pair for Arweave signing
- Pre-configured for testing (not for production use)
- Enables deterministic test results

### Environment Variables

- `RUN_INTEGRATION_TESTS=true` - Required to run integration tests
- `CI` - Automatically detected, skips tests unless explicitly enabled

## Running the Tests

### Prerequisites

```bash
# Install dependencies
go mod tidy
```

### Unit Tests Only

```bash
make test
# or
npx nx test remote-signing
```

### Integration Tests Only

```bash
make test-integration
# or
npx nx test-integration remote-signing
```

### All Tests

```bash
make test-all
# or
npx nx test-all remote-signing
```

### Benchmarks

```bash
make benchmark
# or
npx nx benchmark remote-signing
```

## Test Features

### Goar Integration

- **Wallet Loading**: Loads Arweave wallet from JSON file
- **Data Item Creation**: Creates ANS-104 compatible data items
- **Mock Signing**: Simulates signing process for testing
- **Signature Verification**: Validates wallet ownership

### Server Testing

- **Lifecycle Management**: Start/stop server for each test
- **Health Monitoring**: Waits for server ready state
- **Cleanup**: Graceful shutdown after tests

### WebSocket Testing

- **Connection Management**: Establishes and manages connections
- **Message Handling**: Tests different message types
- **Real-time Verification**: Confirms notifications arrive promptly

### Error Scenarios

- **Network Errors**: Connection failures, timeouts
- **Data Validation**: Invalid formats, oversized data
- **Business Logic**: Missing resources, invalid states

## CI/CD Integration

### GitHub Actions

The tests are designed to run in CI environments:

- Automatically skip in CI unless `RUN_INTEGRATION_TESTS=true`
- No external dependencies required
- Self-contained test environment

### Nx Integration

- Integrated with Nx build system
- Supports affected testing
- Caching for improved performance

## Test Data Management

### Mock Data

- Deterministic test data for consistent results
- Binary data generation for edge case testing
- JSON payloads with various structures

### Cleanup

- Automatic cleanup after each test
- No persistent state between tests
- Memory-based storage (no database required)

## Performance Characteristics

### Response Times

- Data submission: ~1-5ms
- Data retrieval: ~1-3ms
- WebSocket notifications: ~100-200ms
- Concurrent handling: 10+ requests/second

### Memory Usage

- Minimal memory footprint
- Efficient data storage
- Proper garbage collection

## Troubleshooting

### Common Issues

1. **Port Conflicts**

   - Tests use port 8082 by default
   - Change `testServerPort` if needed

2. **Timing Issues**

   - WebSocket tests include retry logic
   - Adjust timeouts for slower systems

3. **Goar Dependencies**
   - Requires proper Go module setup
   - Run `go mod tidy` if import errors occur

### Debug Mode

Set environment variables for additional logging:

```bash
GIN_MODE=debug RUN_INTEGRATION_TESTS=true go test -v .
```

## Future Enhancements

- **Real Arweave Integration**: Test against live Arweave network
- **Performance Testing**: Load testing with higher concurrency
- **Security Testing**: Authentication and authorization scenarios
- **Network Resilience**: Connection failure and recovery testing

## Example Output

```bash
✅ Integration tests completed

=== RUN   TestIntegrationSuite
=== RUN   TestIntegrationSuite/CompleteSigningWorkflow
    integration_test.go:147: Created bundle item with wallet nNlFKCv8sRE4DKe8vWKs...: 60 bytes
    integration_test.go:171: Received UUID: d6e36a43-b9ca-4cde-bd06-86cbd8f8cbbd
    integration_test.go:185: Successfully retrieved 60 bytes
    integration_test.go:202: Signed data item: 95 bytes with mock signature
    integration_test.go:222: Server response: Data signed successfully
    integration_test.go:237: ✅ Signature verification successful! Signer: nNlFKCv8sRE4DKe8vWKs...
--- PASS: TestIntegrationSuite/CompleteSigningWorkflow (0.00s)
--- PASS: TestIntegrationSuite/JSONDataSubmission (0.00s)
--- PASS: TestIntegrationSuite/BinaryDataSubmission (0.00s)
--- PASS: TestIntegrationSuite/WebSocketNotifications (0.10s)
--- PASS: TestIntegrationSuite/ErrorHandling (0.01s)
--- PASS: TestIntegrationSuite/MultipleRequests (0.00s)
PASS
```

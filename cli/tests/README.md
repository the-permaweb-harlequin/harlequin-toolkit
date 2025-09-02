# Harlequin CLI Tests

This directory contains comprehensive tests for the Harlequin CLI organized by test type.

## Test Structure

### 🔬 Unit Tests (`unit/`)

Tests individual functions and components in isolation:

- **Bundling logic** - Lua bundling, dependency resolution
- **Docker operations** - Container management, image building
- **Configuration handling** - YAML parsing, validation
- **Utility functions** - File operations, formatting, etc.

### 🔗 Integration Tests (`integration/`)

Tests component interactions and TUI workflows using **bubbles/test** and **goldie**:

- **TUI components** - Interactive components using bubbles/test patterns
- **Command workflows** - Multi-step processes, state transitions
- **Snapshot testing** - UI output validation with goldie golden files
- **Component integration** - Progress, result, and selector components
- **Workflow simulation** - Complete user interaction flows

### 🎭 End-to-End Tests (`e2e/`)

Tests complete CLI commands as a user would run them:

- **CLI compilation** - Build the binary for testing
- **Command execution** - Run actual CLI commands
- **File system effects** - Verify outputs, artifacts
- **Process exit codes** - Ensure proper error handling

## Running Tests

```bash
# From the cli/ directory:

# Run all tests
make test

# Run specific test types
make test-unit
make test-integration
make test-e2e

# Run tests with coverage
make test-coverage

# Run tests in verbose mode
make test-verbose
```

## Integration Test Features

### Bubbles/Test Integration

- **TUI Model Testing**: Create and test TUI models with realistic interactions
- **Key Event Simulation**: Test keyboard navigation and input handling
- **State Transition Testing**: Validate workflow state changes
- **Component Interaction**: Test list selectors, progress bars, result displays

### Goldie Snapshot Testing

- **UI Output Validation**: Compare current UI output with golden files
- **Regression Detection**: Catch unexpected UI changes
- **Cross-platform Consistency**: Ensure UI looks the same across environments
- **Update Golden Files**: `go test -update` to refresh snapshots

### Test Organization

```
integration/
├── tui_bubbles_test.go          # Core TUI interaction tests
├── workflow_integration_test.go  # Workflow and component tests
├── tui_test.go                  # Basic integration tests
└── testdata/                   # Golden files for snapshot tests
    ├── initial_command_selection.golden
    ├── help_view.golden
    ├── build_workflow.golden
    └── upload_workflow.golden
```

## Test Guidelines

### Unit Tests

- Mock external dependencies
- Test edge cases and error conditions
- Keep tests fast and isolated
- Use table-driven tests for multiple scenarios

### Integration Tests

- **bubbles/test patterns**: Test TUI interactions, key handling, state management
- **goldie snapshot testing**: Validate UI output consistency with golden files
- **Component integration**: Test progress, result, and selector components together
- **Workflow simulation**: Test complete user interaction flows end-to-end
- **Mock external APIs** but test real component interactions
- **State transition testing**: Validate navigation and workflow states

### E2E Tests

- Compile the actual CLI binary
- Test real command execution
- Use temporary directories for file operations
- Test both success and failure scenarios
- Verify actual file outputs and side effects

## Test Utilities

Each test type has its own helper utilities in the `helpers/` subdirectory:

- **Fixtures** - Sample files, configurations, expected outputs
- **Mocks** - Mock implementations of external services
- **Assertions** - Custom test assertions and validation helpers
- **Setup/Teardown** - Test environment preparation and cleanup

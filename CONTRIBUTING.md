# Contributing to specforge

Thank you for your interest in contributing to `specforge`! Follow these guidelines to set up your environment, write tests, and submit changes.

---

## Development Setup

1. **Go Version:** Verify that you are running Go 1.24 or later:
   ```bash
   go version
   ```
2. **Clone and Install Dependencies:**
   ```bash
   git clone https://github.com/SSOHEB/SpecForge.git
   cd specforge
   go mod download
   ```

---

## Verification & Testing Tier

We use a testing pyramid containing unit tests, golden tests, integration tests, and benchmarks. Ensure all tests pass before proposing modifications.

### 1. Unit Tests (Isolated Package Verification)
Verifies individual packages inside `internal/` and `cmd/`:
```bash
make test
```

### 2. Golden Tests (Code & Schema Generator Output)
Asserts that changes to generators do not cause unintended deviations in output files.
```bash
make test-golden
```
To intentionally regenerate golden files after changing generator outputs:
```bash
go test ./tests/golden/... -update
```

### 3. Integration Tests (CLI E2E Checks)
Drives the CLI as a black box executing `validate` and `generate` subcommands on valid and invalid templates:
```bash
make test-integration
```

### 4. Benchmarks
Measures execution performance and memory allocations:
```bash
make bench
```

## Pre-commit checks

To ensure code quality before pushing, you can set up pre-commit hooks. We support two options:

**Option 1: Using the pre-commit framework (Recommended)**
If you have [pre-commit](https://pre-commit.com) installed, run:
```bash
pre-commit install
```
This will configure hooks for `go fmt`, `go vet`, and `golangci-lint` to run automatically before every commit.

**Option 2: Plain Git Hook**
If you don't use the pre-commit framework, you can use the provided fallback script:
```bash
cp scripts/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

---

## Coding Style & Linters

We enforce strict Go formatting and lint checking. Run formatting and lint checks locally:

```bash
# Format source code
make fmt

# Run linters
golangci-lint run ./...
```

* **Doc Comments:** Every exported constant, type, method, function, and variable must have a doc comment starting with its own name.
* **Imports:** Use `goimports` to sort and manage import blocks automatically.
* **Error Handling:** Always check returned errors (e.g. do not ignore errors returned by `os.Setenv` or `f.Close` inside production or test files).

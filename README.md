# configforge

A semantic configuration management and generation framework for Go applications.

**Project Status: Feature-Complete (All 18 stages fully implemented and verified).**

---

## Installation

### From Binary Releases
Download the pre-compiled binary for your platform from the GitHub Releases page.

### From Source
```bash
go install github.com/SSOHEB/configforge/cmd/configforge@latest
```

### Via Docker
Run the command directly inside the container:
```bash
docker run --rm -v ${PWD}:/workspace -w /workspace ghcr.io/ssoheb/configforge:latest defaults -m metadata.yaml
```

---

## CLI Usage

Every command supports `--metadata` (`-m`), `--config` (`-c`), and `--out` (`-o`) flags.

### 1. defaults
Prints a list of all fields, their Go types, and default values or requirements:
```bash
configforge defaults --metadata examples/http-server/metadata.yaml
```

### 2. validate
Validates configuration files against semantic rules (e.g. min, max, enum, regex pattern):
```bash
configforge validate --metadata examples/http-server/metadata.yaml --config examples/http-server/config.yaml
```

### 3. generate
Generates draft-07 JSON Schema (`schema.json`) and typed Go config structures (`generated_config.go`):
```bash
configforge generate --metadata examples/http-server/metadata.yaml --out examples/http-server --functional-api
```

### 4. schema
Generates JSON Schema. Prints to standard output, or writes to `--out` if `-w`/`--write` is provided:
```bash
configforge schema --metadata examples/http-server/metadata.yaml
```

### 5. docs
Generates reference Markdown documentation from a metadata specification:
```bash
configforge docs --metadata examples/http-server/metadata.yaml --out examples/http-server
```

---

## Testing

configforge features a testing pyramid containing unit tests, golden tests, black-box integration tests, and benchmarks.

### 1. Unit Tests (Fast)
Verifies individual components (parser, AST builder, generators, validator, defaults engine, watcher) in isolation:
```bash
make test
```

### 2. Golden Tests
Ensures output generators (JSON Schema, Go structures, Markdown docs) do not drift from canonical expected outputs.
```bash
make test-golden
```
To deliberately regenerate golden files when generator logic is updated:
```bash
go test ./tests/golden/... -update
```

### 3. Integration Tests
Drives the CLI as a black box using `go run` to test end-to-end generating and validating on valid and invalid configuration templates:
```bash
make test-integration
```

### 4. Benchmarks
Benchmarks configuration loading and AST compiling to measure performance and track memory allocations:
```bash
make bench
```

---

## Building from Source & Platform Execution Quirks

To compile the `configforge` CLI tool from source:
```bash
go build -o bin/configforge ./cmd/configforge
```

> [!NOTE]
> On some Windows machines, local Application Control policies or Windows Defender heuristics might block the execution of freshly compiled application binaries that implement folder-watching (`fsnotify`) or connection retry loops (such as the E2E example servers). If you run into execution blocks when executing compiled example binaries locally, use `go run` or execute them in isolated container environments (Docker/WSL2) for validation. The `configforge` CLI binary itself performs standard file parsing and generation and executes normally on Windows.

---

## Project Status

`configforge` has evolved from initial scaffolding to a comprehensive, enterprise-ready configuration management framework:
1. Normalized metadata spec parser (YAML).
2. Dynamic Abstract Syntax Tree (AST) compiler with validator flags.
3. Deterministic JSON Schema Generator.
4. Go struct generator with Custom Marshal/Unmarshal tags.
5. Functional Getter API generator.
6. Build-time schema generator command.
7. Declarative defaults CLI table command.
8. CLI command for validating files against schemas.
9. Runtime loader with automatic default injection.
10. Runtime environment variable overrides.
11. Real-time config watcher with debounced hot-reload and atomic swaps.
12. Comprehensive CLI runner (`cobra`).
13. Automatic reference Markdown documentation generator.
14. Unit, golden, and integration test pyramids.
15. Dynamic benchmarking suite.
16. Cross-platform compilation and release registry.
17. Docker multi-stage container packaging.
18. Automating release pipeline (GitHub Actions).

# configforge

A semantic configuration management and generation framework for Go applications.

Status: active development

## CLI Usage

Build the CLI:
```bash
go build -o bin/configforge ./cmd/configforge
```

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

### 5. docs (Stub)
Generates configuration documentation (stubbed for now):
```bash
configforge docs
```

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

## Building from Source & Platform Execution Quirks

To compile the `configforge` CLI tool from source:
```bash
go build -o bin/configforge ./cmd/configforge
```

> [!NOTE]
> On some Windows machines, local Application Control policies or Windows Defender heuristics might block the execution of freshly compiled application binaries that implement folder-watching (`fsnotify`) or connection retry loops (such as the E2E example servers). If you run into execution blocks when executing compiled example binaries locally, use `go run` or execute them in isolated container environments (Docker/WSL2) for validation. The `configforge` CLI binary itself performs standard file parsing and generation and executes normally on Windows.


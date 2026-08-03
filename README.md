# configforge

`configforge` is a semantic configuration management and generation framework for Go applications.

**Project Status: Feature-Complete (All 18 stages fully implemented and verified).**

---

## Overview

`configforge` separates configuration **specification** from configuration **data**. By defining your configuration schema, validation rules, default values, and description comments in a single YAML file (`metadata.yaml`), `configforge` compiles this specification into an AST. This AST is then used to generate type-safe Go structs, JSON schemas for IDE autocomplete, and Markdown documentation, while also providing a runtime loader with default values, environment overrides, and file-watching hot reloads.

---

## Why configforge?

In traditional Go application architectures, configuration management often falls into two patterns:
1. **Plain Struct Unmarshaling:** Minimal validation, boilerplate code, zero defaults handling, and silent failures when required fields are missing.
2. **Dynamic Maps/Loose Types (Viper):** Prone to typos, lacks compile-time type-safety, and bypasses Go's type compiler.

`configforge` is inspired by the **OpenTelemetry Configuration Specification** philosophy. It allows developers to:
* Declare all configuration rules in a metadata file.
* Compile and validate the metadata structure at build-time.
* Compile strongly typed Go config structures and functional APIs.
* Enforce semantic validation checks (required fields, value ranges, enums, regex patterns) at application startup and reload, cleanly decoupled from runtime serialization formats.

---

## Installation

### 1. Pre-built Binaries
Download the compiled release binary for your OS and architecture from the GitHub Releases page.

### 2. From Source
```bash
go install github.com/SSOHEB/configforge/cmd/configforge@latest
```

### 3. Via Docker
Run `configforge` inside a container (sharing the working directory):
```bash
docker run --rm -v ${PWD}:/workspace -w /workspace ghcr.io/ssoheb/configforge:latest defaults -m metadata.yaml
```

---

## Quick Start

Get up and running in under 2 minutes:

### 1. Declare the Specification (`metadata.yaml`)
Create a namespace `server` with nested fields and value constraints:

```yaml
server:
  host:
    type: string
    default: "localhost"
  port:
    type: int
    default: 8080
    min: 1
    max: 65535
    description: "Port to listen on"
  api_key:
    type: string
    pattern: "^[A-Z0-9]{16}$"
    required: true
```

### 2. Generate Structs & JSON Schema
Compile the metadata specification into typed Go structures with a Functional Getter API:

```bash
configforge generate --metadata metadata.yaml --out ./config --functional-api
```

This generates `config/generated_config.go` containing:
```go
type ServerConfig struct {
    HostField string `yaml:"host"`
    PortField int    `yaml:"port"`
}

func (s *ServerConfig) Host() string { return s.HostField }
func (s *ServerConfig) Port() int    { return s.PortField }
```

### 3. Provide Configuration (`config.yaml`)
Create a local configuration file containing overrides:

```yaml
server:
  port: 9000
  api_key: "MYSECRETAPIKEY12"
```

### 4. Load, Validate, and Access in Go
Load the configuration, automatically injecting default values (like `host: "localhost"`) and checking validation constraints:

```go
package main

import (
	"fmt"
	"log"

	"configforge/internal/runtime"
	"configforge/internal/schema"
	"configforge/internal/parser"
)

func main() {
	// Parse metadata spec to build the AST schema
	rawMeta, _ := parser.ParseFile("metadata.yaml")
	ast, _ := schema.Build(rawMeta)

	// Load configuration, apply defaults, env overrides, and validate
	cfg, _, err := runtime.LoadAndPrepareFile[Config](ast, "config.yaml")
	if err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Access configuration values using the safe Functional API
	fmt.Printf("Starting server on %s:%d...\n", cfg.Server().Host(), cfg.Server().Port())
}
```

---

## CLI Reference

Every command supports `--metadata` (`-m`), `--config` (`-c`), and `--out` (`-o`) flags.

### `defaults`
Prints a formatted table listing all defined configuration paths, Go types, and default values or required flags:
```bash
configforge defaults --metadata metadata.yaml
```

### `validate`
Validates application configuration files against semantic constraints (ranges, enums, regexes, required keys):
```bash
configforge validate --metadata metadata.yaml --config config.yaml
```

### `generate`
Compiles both the Draft-07 JSON Schema (`schema.json`) and the typed Go structs (`generated_config.go`):
```bash
configforge generate --metadata metadata.yaml --out ./generated --functional-api
```

### `schema`
Generates and prints the JSON Schema document. Use `-w` to write to file:
```bash
configforge schema --metadata metadata.yaml
```

### `docs`
Generates structured Reference Markdown documentation:
```bash
configforge docs --metadata metadata.yaml --out ./docs
```

### `version`
Displays version details, Git commit hash, and build timestamp:
```bash
configforge version
```

---

## Architecture

`configforge` operates as a compiler pipeline separating frontend parsing from backend code generation and runtime layers:

```
                  +-----------------------+
                  |  spec/metadata.yaml   |
                  +-----------------------+
                              |
                              v
                  +-----------------------+
                  |    internal/parser    | (YAML parsing)
                  +-----------------------+
                              |
                              v
                  +-----------------------+
                  |    internal/schema    | (AST Compilation & Checks)
                  +-----------------------+
                              |
         +--------------------+--------------------+
         |                                         |
         v                                         v
+------------------+                      +------------------+
|internal/generator| (Go/JSON/Markdown)   |internal/validator| (Semantic checks)
+------------------+                      +------------------+
         |                                         |
         +--------------------+--------------------+
                              |
                              v
                  +-----------------------+
                  |   internal/runtime    | (YAML Loader, Watcher, Envs)
                  +-----------------------+
```

---

## Testing

configforge features a testing pyramid containing unit tests, golden tests, black-box integration tests, and benchmarks.

### 1. Unit Tests
Verifies internal business rules of individual packages:
```bash
make test
```

### 2. Golden Tests
Prevents generator output drift against pre-recorded expected outputs:
```bash
make test-golden
```
To intentionally regenerate golden files:
```bash
go test ./tests/golden/... -update
```

### 3. Integration Tests
Verifies CLI command executions and error statuses as a black box:
```bash
make test-integration
```

### 4. Benchmarks
Measures execution performance and memory footprint:
```bash
make bench
```

---

## Contributing

For guidelines on setting up local environments, running the testing tiers, and code conventions, see [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

This project is licensed under the MIT License. See the LICENSE file for details.

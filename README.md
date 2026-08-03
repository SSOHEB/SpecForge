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

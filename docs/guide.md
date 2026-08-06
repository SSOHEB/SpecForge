# codrao Narrative Developer Guide

Welcome to `codrao`! This guide walks you through defining, generating, loading, and dynamically reloading configurations in Go applications.

---

## 1. Defining Your Specification (`metadata.yaml`)

Configuration is declared in a single source-of-truth metadata file. Namespaces represent nested structs, and fields define concrete config values and validation constraints.

```yaml
# Examples of all allowed field types and constraints
server:
  host:
    type: string
    default: "localhost"
    description: "Server host address"
  port:
    type: int
    default: 8080
    min: 1
    max: 65535
    description: "Server port number"
  ssl_mode:
    type: string
    enum: [disable, require, verify-ca]
    default: "disable"
  api_key:
    type: string
    pattern: "^[A-Z0-9]{16}$"
    required: true
  allowed_ips:
    type: string[]
    description: "List of authorized IP addresses"
```

### Supported Validation Constraints
* **`required`**: If `true`, the field must be physically present in the input configuration map. Absence triggers validation errors.
* **`min` / `max`**: Applicable to `int` and `float` fields. Enforces ranges.
* **`enum`**: Applicable to `string` fields. Value must match one of the listed options.
* **`pattern`**: Applicable to `string` fields. Checks matching against a regular expression.

---

### YAML Aliases and Anchors

Codrao parses `metadata.yaml` with `gopkg.in/yaml.v3`, but it does not expand aliases or merge keys into schema structure. Keep the schema explicit.

What works today:

- An anchor on a mapping that is used directly, for example `server: &server`, is parsed as a normal namespace. The anchor itself has no effect.
- A scalar anchor/alias inside a leaf value, for example `default: &d 8080` followed by `default: *d`, may decode through the YAML library, but Codrao does not rely on it. Treat it as outside the supported contract.

What fails:

- Reusing a mapping with an alias instead of spelling it out returns `expected mapping node`.
- YAML merge keys such as `<<: *base` return `expected mapping node`.
- An alias in place of a namespace mapping returns `expected mapping node`.

Example that is not supported:

```yaml
base: &base
  host:
    type: string
server:
  <<: *base
```

Use explicit values instead. Defaults and environment overrides already cover the common reuse cases without anchors.

---

## 2. Generating Code & Schemas

Use the CLI to compile your metadata file:

```bash
codrao generate --metadata metadata.yaml --out ./config --functional-api
```

### Functional API vs Plain Structs
By default, the CLI generates standard public fields. However, using `--functional-api` generates getter methods (e.g., `Port()`). 

To resolve name collisions in Go (since a struct cannot have a field and a method with the same name), codrao automatically appends a `Field` suffix to the struct properties:
```go
// Generated output when --functional-api is enabled
type ServerConfig struct {
    HostField string `yaml:"host"`
    PortField int    `yaml:"port"`
}

func (s *ServerConfig) Host() string { return s.HostField }
func (s *ServerConfig) Port() int    { return s.PortField }
```

---

## 3. Runtime Configuration Loading

The `internal/runtime` package processes the configuration lifecycle in three sequential steps:

1. **Defaults Injection:** Missing fields in the YAML are populated using AST-defined defaults.
2. **Environment Variable Overrides:** Environment overrides are mapped. Dotted paths are joined with `_`, capitalized, and prefixed with `CODRAO_` (e.g., `server.port` becomes `CODRAO_SERVER_PORT`).
3. **Semantic Validation:** The configuration is validated against the AST semantic rules (required checks, ranges, enums, regex patterns).

```go
cfg, rawMap, err := runtime.LoadAndPrepareFile[Config](ast, "config.yaml")
if err != nil {
    log.Fatal("Invalid config configuration:", err)
}
```

---

## 4. Live Configuration Watching (Hot Reload)

To watch for configuration changes dynamically, initialize a `Watcher`.

The watcher monitors the file's parent directory rather than the file itself. This prevents issues with editors that save changes by writing to a temporary file and atomically renaming/moving it, which changes the file's OS inode.

```go
w, err := runtime.NewWatcher[Config](ast, "config.yaml")
if err != nil {
    log.Fatalf("failed to initialize watcher: %v", err)
}

// Register success callback
w.OnReload(func(newCfg *Config) {
    fmt.Printf("Config reloaded! New Port: %d\n", newCfg.Server().Port())
})

// Register error callback
w.OnError(func(err error) {
    log.Printf("Reload failed: %v", err)
})

// Run watcher
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go w.Start(ctx)
```

Updates are debounced (100ms window) to prevent rapid successive saves from triggering redundant compiles, and reloaded configurations are atomically swapped to prevent partial states.

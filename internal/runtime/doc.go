// Package runtime handles configuration loading, default value injection, environment overrides, and file-watching hot reloads.
//
// This package is imported directly by applications to manage their configuration lifecycle at runtime.
//
// # Pipeline Role
//
// Runtime prepares the configuration data for application consumption:
//
//	[File/Env] -> [Load] -> [Apply Defaults] -> [Apply Env Overrides] -> [Validate] -> [Unmarshal Struct]
//
// # Key Components
//
//   - Loader (LoadAndPrepare): Reads configuration bytes from a file or reader, injects AST-defined default values
//     for missing keys, parses and overrides values using environment variables, validates the result, and unmarshals
//     it into the target struct.
//   - Overrides (ApplyEnvOverrides): Translates configuration paths to environment variables (e.g.
//     "instrumentation.http.port" becomes "SPECFORGE_INSTRUMENTATION_HTTP_PORT"), parsing them into the correct Go types.
//   - Watcher: Monitors filesystem events via fsnotify. Debounces rapid-save events (e.g., 100ms), compiles the updated
//     configuration, validates it, and atomically swaps the active configuration reference if valid.
package runtime

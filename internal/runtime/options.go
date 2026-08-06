package runtime

// RuntimeOptions provides configuration for the runtime loading process.
type RuntimeOptions struct {
	EnvPrefix    string // if empty and env overrides enabled, defaults to "CODRAO_"
	DisableEnv   bool
	MetadataPath string // if empty, caller must handle discovery (see pkg/config below)
}

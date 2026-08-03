package runtime

import "fmt"

// ConfigFileNotFoundError is returned when the config file cannot be found.
type ConfigFileNotFoundError struct {
	Path string
	Err  error
}

func (e *ConfigFileNotFoundError) Error() string {
	return fmt.Sprintf("config file not found at %s: %v", e.Path, e.Err)
}

// ConfigEmptyError is returned when the config file/reader is empty.
type ConfigEmptyError struct{}

func (e *ConfigEmptyError) Error() string {
	return "config is empty"
}

// ConfigSyntaxError is returned when YAML parsing fails.
type ConfigSyntaxError struct {
	Err error
}

func (e *ConfigSyntaxError) Error() string {
	return fmt.Sprintf("config YAML syntax error: %v", e.Err)
}

// EnvParseError is returned when an environment variable value fails to parse.
type EnvParseError struct {
	EnvVar       string
	Value        string
	ExpectedType string
	Err          error
}

func (e *EnvParseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("failed to parse environment variable %s=%q as %s: %v", e.EnvVar, e.Value, e.ExpectedType, e.Err)
	}
	return fmt.Sprintf("failed to parse environment variable %s=%q as %s", e.EnvVar, e.Value, e.ExpectedType)
}


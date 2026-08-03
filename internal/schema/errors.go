package schema

import "fmt"

// DefaultTypeMismatchError is returned when a default value does not match the field type.
type DefaultTypeMismatchError struct {
	Path     string
	Expected string
	Actual   string
	Err      error
}

func (e *DefaultTypeMismatchError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("default type mismatch at %s: expected %s, got actual value with error: %v", e.Path, e.Expected, e.Err)
	}
	return fmt.Sprintf("default type mismatch at %s: expected %s, got %s", e.Path, e.Expected, e.Actual)
}

// SiblingNameCollisionError is returned when two sibling fields/nodes resolve to the same Go name.
type SiblingNameCollisionError struct {
	Path   string
	GoName string
	Keys   []string
}

func (e *SiblingNameCollisionError) Error() string {
	return fmt.Sprintf("name collision at %s: Go name %q resolved from multiple YAML keys: %v", e.Path, e.GoName, e.Keys)
}

// InvalidRegexPatternError is returned when a pattern regex fails compilation at build time.
type InvalidRegexPatternError struct {
	Path    string
	Pattern string
	Err     error
}

func (e *InvalidRegexPatternError) Error() string {
	return fmt.Sprintf("invalid regex pattern %q at %s: %v", e.Pattern, e.Path, e.Err)
}

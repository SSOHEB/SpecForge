package parser

import "fmt"

// ValidationError represents a structural validation error with a path and message.
type ValidationError struct {
	Path string
	Msg  string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Msg)
}

// InvalidYAMLSyntaxError is returned when YAML is malformed.
type InvalidYAMLSyntaxError struct {
	Err error
}

func (e *InvalidYAMLSyntaxError) Error() string {
	return fmt.Sprintf("invalid YAML syntax: %v", e.Err)
}

// UnknownTypeError is returned when a field has an unsupported type.
type UnknownTypeError struct {
	Path string
	Type string
}

func (e *UnknownTypeError) Error() string {
	return fmt.Sprintf("%s: unknown type %q", e.Path, e.Type)
}

// InvalidAttributeCombinationError is returned when attributes don't match the type.
type InvalidAttributeCombinationError struct {
	Path string
	Msg  string
}

func (e *InvalidAttributeCombinationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Msg)
}

// MissingTypeError is returned when a field has no type.
type MissingTypeError struct {
	Path string
}

func (e *MissingTypeError) Error() string {
	return fmt.Sprintf("%s: missing required 'type' field", e.Path)
}

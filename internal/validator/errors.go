package validator

import (
	"fmt"
	"strings"
)

// ValidationError represents a single validation rule violation.
type ValidationError struct {
	Path string
	Msg  string
	Rule string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s (rule: %s)", e.Path, e.Msg, e.Rule)
}

// ValidationErrors is a collection of ValidationErrors that implements the error interface.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var sb strings.Builder
	for i, err := range e {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

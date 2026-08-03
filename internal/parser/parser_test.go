package parser

import (
	"errors"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name          string
		yamlInput     string
		expectErr     bool
		errTypeCheck  func(error) bool
		errStringSub  string
	}{
		{
			name: "valid nested metadata",
			yamlInput: `
instrumentation:
  http:
    enabled:
      type: bool
      default: true
      description: "Enables HTTP instrumentation"
      required: false
    capture_headers:
      type: string[]
      description: "Headers to capture"
    port:
      type: int
      default: 8080
      min: 1
      max: 65535
    log_level:
      type: string
      enum: [debug, info, warn, error]
      default: info
`,
			expectErr: false,
		},
		{
			name: "missing type",
			yamlInput: `
instrumentation:
  http:
    enabled:
      default: true
`,
			expectErr: true,
			errTypeCheck: func(err error) bool {
				var target *MissingTypeError
				return errors.As(err, &target)
			},
			errStringSub: "instrumentation.http.enabled: missing required 'type' field",
		},
		{
			name: "unknown type",
			yamlInput: `
instrumentation:
  http:
    enabled:
      type: invalid_type
`,
			expectErr: true,
			errTypeCheck: func(err error) bool {
				var target *UnknownTypeError
				return errors.As(err, &target)
			},
			errStringSub: `instrumentation.http.enabled: unknown type "invalid_type"`,
		},
		{
			name: "invalid attribute combinations - enum on bool",
			yamlInput: `
instrumentation:
  http:
    enabled:
      type: bool
      enum: [true, false]
`,
			expectErr: true,
			errTypeCheck: func(err error) bool {
				var target *InvalidAttributeCombinationError
				return errors.As(err, &target)
			},
			errStringSub: "instrumentation.http.enabled: enum is not valid for type bool",
		},
		{
			name: "invalid attribute combinations - min on string",
			yamlInput: `
instrumentation:
  http:
    log_level:
      type: string
      min: 1
`,
			expectErr: true,
			errTypeCheck: func(err error) bool {
				var target *InvalidAttributeCombinationError
				return errors.As(err, &target)
			},
			errStringSub: "instrumentation.http.log_level: min is not valid for type string",
		},
		{
			name: "invalid attribute combinations - pattern on int",
			yamlInput: `
instrumentation:
  http:
    port:
      type: int
      pattern: "^[0-9]+$"
`,
			expectErr: true,
			errTypeCheck: func(err error) bool {
				var target *InvalidAttributeCombinationError
				return errors.As(err, &target)
			},
			errStringSub: "instrumentation.http.port: pattern is not valid for type int",
		},
		{
			name: "empty file",
			yamlInput: ``,
			expectErr: false,
		},
		{
			name: "malformed YAML syntax",
			yamlInput: `
instrumentation:
  http:
    port
      type: int
`,
			expectErr: true,
			errTypeCheck: func(err error) bool {
				var target *InvalidYAMLSyntaxError
				return errors.As(err, &target)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.yamlInput)
			meta, err := Parse(r)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errTypeCheck != nil && !tt.errTypeCheck(err) {
					t.Errorf("error did not match expected type: %v", err)
				}
				if tt.errStringSub != "" && !strings.Contains(err.Error(), tt.errStringSub) {
					t.Errorf("expected error to contain %q, got %q", tt.errStringSub, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				if meta == nil {
					t.Fatalf("expected metadata object, got nil")
				}
			}
		})
	}
}

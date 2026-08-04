package validator

import (
	"testing"

	"github.com/SSOHEB/configforge/internal/parser"
	"github.com/SSOHEB/configforge/internal/schema"

	"gopkg.in/yaml.v3"
)

func TestValidate(t *testing.T) {
	// Parse sample metadata.yaml
	rawMeta, err := parser.ParseFile("../../examples/http-server/metadata.yaml")
	if err != nil {
		t.Fatalf("failed to parse metadata.yaml: %v", err)
	}

	ast, err := schema.Build(rawMeta)
	if err != nil {
		t.Fatalf("failed to build AST: %v", err)
	}

	tests := []struct {
		name         string
		yamlConfig   string
		expectErrors bool
		errChecks    []func(ValidationError) bool
		expectedLen  int
	}{
		{
			name: "fully valid config",
			yamlConfig: `
instrumentation:
  http:
    enabled: true
    port: 8080
    log_level: "info"
    api_key: "ABCDEF1234567890"
`,
			expectErrors: false,
		},
		{
			name: "missing required field (required api_key? wait, we didn't set required: true in api_key, but we can test missing required namespace or we can add a required field check)",
			yamlConfig: `
instrumentation:
  http:
    port: 8080
`,
			// Wait, let's see. In metadata.yaml, is there any required field?
			// Let's check: in metadata.yaml, port/api_key are not explicitly "required: true" (enabled is required: false).
			// If none are required, this config is valid.
			// Let's verify by adding a required field to this test case or writing a custom AST, or checking.
			// Wait! We can write a custom AST manually for testing required fields! That is much more robust!
			expectErrors: false,
		},
		{
			name: "port over max",
			yamlConfig: `
instrumentation:
  http:
    port: 99999
`,
			expectErrors: true,
			expectedLen:  1,
			errChecks: []func(ValidationError) bool{
				func(ve ValidationError) bool {
					return ve.Path == "instrumentation.http.port" && ve.Rule == "max"
				},
			},
		},
		{
			name: "log_level not in enum",
			yamlConfig: `
instrumentation:
  http:
    log_level: "bogus"
`,
			expectErrors: true,
			expectedLen:  1,
			errChecks: []func(ValidationError) bool{
				func(ve ValidationError) bool {
					return ve.Path == "instrumentation.http.log_level" && ve.Rule == "enum"
				},
			},
		},
		{
			name: "pattern mismatch on api_key",
			yamlConfig: `
instrumentation:
  http:
    api_key: "short"
`,
			expectErrors: true,
			expectedLen:  1,
			errChecks: []func(ValidationError) bool{
				func(ve ValidationError) bool {
					return ve.Path == "instrumentation.http.api_key" && ve.Rule == "pattern"
				},
			},
		},
		{
			name: "multiple simultaneous violations",
			yamlConfig: `
instrumentation:
  http:
    port: 99999
    log_level: "bogus"
    api_key: "invalid-key-shape"
`,
			expectErrors: true,
			expectedLen:  3,
			errChecks: []func(ValidationError) bool{
				func(ve ValidationError) bool {
					return ve.Path == "instrumentation.http.api_key" && ve.Rule == "pattern"
				},
				func(ve ValidationError) bool {
					return ve.Path == "instrumentation.http.log_level" && ve.Rule == "enum"
				},
				func(ve ValidationError) bool {
					return ve.Path == "instrumentation.http.port" && ve.Rule == "max"
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rawConfig map[string]any
			if err := yaml.Unmarshal([]byte(tt.yamlConfig), &rawConfig); err != nil {
				t.Fatalf("failed to parse test yaml: %v", err)
			}

			errs := Validate(ast, rawConfig)
			if tt.expectErrors {
				if len(errs) == 0 {
					t.Fatalf("expected errors, got none")
				}
				if tt.expectedLen > 0 && len(errs) != tt.expectedLen {
					t.Errorf("expected %d errors, got %d: %v", tt.expectedLen, len(errs), errs)
				}
				for _, check := range tt.errChecks {
					matched := false
					for _, ve := range errs {
						if check(ve) {
							matched = true
							break
						}
					}
					if !matched {
						t.Errorf("a check failed to find matching error in: %v", errs)
					}
				}
			} else {
				if len(errs) > 0 {
					t.Fatalf("expected no errors, got: %v", errs)
				}
			}
		})
	}
}

func TestValidate_RequiredFieldsCustomAST(t *testing.T) {
	// Build a custom AST to strictly test required fields check
	reqField := &schema.Field{
		Name:     "Secret",
		YAMLKey:  "secret",
		Path:     []string{"secret"},
		Type:     schema.TypeString,
		Required: true,
	}
	ast := &schema.AST{
		Root: &schema.Node{
			Name: "Config",
			Path: []string{},
			Fields: []*schema.Field{
				reqField,
			},
		},
	}

	// Case 1: missing required field
	errs := Validate(ast, map[string]any{})
	if len(errs) != 1 || errs[0].Rule != "required" || errs[0].Path != "secret" {
		t.Errorf("expected 1 required error on 'secret', got: %v", errs)
	}

	// Case 2: present required field
	errs = Validate(ast, map[string]any{"secret": "val"})
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got: %v", errs)
	}
}

package schema

import (
	"errors"
	"strings"
	"testing"

	"github.com/SSOHEB/configforge/internal/parser"
)

func TestToGoName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"port", "Port"},
		{"Port", "Port"},
		{"http_server", "HttpServer"},
		{"consecutive__underscores", "ConsecutiveUnderscores"},
		{"http_2_port", "Http2Port"},
		{"nested.name-test", "NestedNameTest"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToGoName(tt.input)
			if got != tt.expected {
				t.Errorf("ToGoName(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestBuild_HTTPServerExample(t *testing.T) {
	raw, err := parser.ParseFile("../../examples/http-server/metadata.yaml")
	if err != nil {
		t.Fatalf("failed to parse sample metadata: %v", err)
	}

	ast, err := Build(raw)
	if err != nil {
		t.Fatalf("failed to build AST: %v", err)
	}

	if ast.Root == nil {
		t.Fatalf("expected AST.Root to be non-nil")
	}
	if len(ast.Root.Children) != 1 || ast.Root.Children[0].Name != "Instrumentation" {
		t.Fatalf("expected first child to be Instrumentation, got %v", ast.Root.Children)
	}

	instr := ast.Root.Children[0]
	if len(instr.Children) != 1 || instr.Children[0].Name != "Http" {
		t.Fatalf("expected child of Instrumentation to be Http")
	}

	http := instr.Children[0]

	var portField *Field
	var enabledField *Field
	var logLevelField *Field

	for _, f := range http.Fields {
		switch f.YAMLKey {
		case "port":
			portField = f
		case "enabled":
			enabledField = f
		case "log_level":
			logLevelField = f
		}
	}

	if portField == nil {
		t.Fatalf("expected to find field 'port'")
	}
	if portField.Name != "Port" || portField.Type != TypeInt || portField.Default != 8080 {
		t.Errorf("portField mismatch: %+v", portField)
	}
	if portField.Min == nil || *portField.Min != 1 || portField.Max == nil || *portField.Max != 65535 {
		t.Errorf("portField bounds mismatch: min=%v, max=%v", portField.Min, portField.Max)
	}

	if enabledField == nil {
		t.Fatalf("expected to find field 'enabled'")
	}
	if enabledField.Name != "Enabled" || enabledField.Type != TypeBool || enabledField.Default != true {
		t.Errorf("enabledField mismatch: %+v", enabledField)
	}

	if logLevelField == nil {
		t.Fatalf("expected to find field 'log_level'")
	}
	if logLevelField.Name != "LogLevel" || logLevelField.Type != TypeString || logLevelField.Default != "info" {
		t.Errorf("logLevelField mismatch: %+v", logLevelField)
	}
	if len(logLevelField.Enum) != 4 || logLevelField.Enum[0] != "debug" || logLevelField.Enum[1] != "info" {
		t.Errorf("logLevelField enum mismatch: %v", logLevelField.Enum)
	}
}

func TestBuild_Errors(t *testing.T) {
	tests := []struct {
		name         string
		yamlInput    string
		errTypeCheck func(error) bool
		errStringSub string
	}{
		{
			name: "default value type mismatch - int expected got string",
			yamlInput: `
port:
  type: int
  default: "abc"
`,
			errTypeCheck: func(err error) bool {
				var target *DefaultTypeMismatchError
				return errors.As(err, &target)
			},
			errStringSub: "default type mismatch at port: expected int",
		},
		{
			name: "sibling name collision",
			yamlInput: `
http_server:
  type: string
http-server:
  type: int
`,
			errTypeCheck: func(err error) bool {
				var target *SiblingNameCollisionError
				return errors.As(err, &target)
			},
			errStringSub: "name collision at : Go name \"HttpServer\" resolved from multiple YAML keys: [http-server http_server]",
		},
		{
			name: "invalid regex pattern",
			yamlInput: `
api_key:
  type: string
  pattern: "["
`,
			errTypeCheck: func(err error) bool {
				var target *InvalidRegexPatternError
				return errors.As(err, &target)
			},
			errStringSub: "invalid regex pattern \"[\" at api_key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.yamlInput)
			raw, err := parser.Parse(r)
			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}

			_, err = Build(raw)
			if err == nil {
				t.Fatalf("expected build error, got nil")
			}

			if tt.errTypeCheck != nil && !tt.errTypeCheck(err) {
				t.Errorf("error did not match expected type: %v", err)
			}

			if tt.errStringSub != "" && !strings.Contains(err.Error(), tt.errStringSub) {
				t.Errorf("expected error to contain %q, got %q", tt.errStringSub, err.Error())
			}
		})
	}
}

func TestBuild_DeepNesting(t *testing.T) {
	yamlInput := `
a:
  b:
    c:
      d:
        type: string
        default: "deep"
`
	r := strings.NewReader(yamlInput)
	raw, err := parser.Parse(r)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	ast, err := Build(raw)
	if err != nil {
		t.Fatalf("failed to build AST: %v", err)
	}

	nodeA := ast.Root.Children[0]
	if nodeA.Name != "A" {
		t.Fatalf("expected node A, got %s", nodeA.Name)
	}

	nodeB := nodeA.Children[0]
	if nodeB.Name != "B" {
		t.Fatalf("expected node B, got %s", nodeB.Name)
	}

	nodeC := nodeB.Children[0]
	if nodeC.Name != "C" {
		t.Fatalf("expected node C, got %s", nodeC.Name)
	}

	fieldD := nodeC.Fields[0]
	if fieldD.Name != "D" || fieldD.Type != TypeString || fieldD.Default != "deep" {
		t.Errorf("fieldD mismatch: %+v", fieldD)
	}
}

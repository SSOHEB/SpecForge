package runtime

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"configforge/internal/parser"
	"configforge/internal/schema"
	"configforge/internal/validator"
)

type testEnvConfig struct {
	BooleanField bool              `yaml:"boolean_field"`
	IntegerField int               `yaml:"integer_field"`
	FloatField   float64           `yaml:"float_field"`
	StringField  string            `yaml:"string_field"`
	StringSlice  []string          `yaml:"string_slice"`
	IntSlice     []int             `yaml:"int_slice"`
	StringMap    map[string]string `yaml:"string_map"`
	Nested       struct {
		Value int `yaml:"value"`
	} `yaml:"nested"`
}

func buildTestEnvAST() *schema.AST {
	ast := &schema.AST{
		Root: &schema.Node{
			Name: "Config",
			Fields: []*schema.Field{
				{Name: "BooleanField", YAMLKey: "boolean_field", Path: []string{"boolean_field"}, Type: schema.TypeBool},
				{Name: "IntegerField", YAMLKey: "integer_field", Path: []string{"integer_field"}, Type: schema.TypeInt},
				{Name: "FloatField", YAMLKey: "float_field", Path: []string{"float_field"}, Type: schema.TypeFloat},
				{Name: "StringField", YAMLKey: "string_field", Path: []string{"string_field"}, Type: schema.TypeString},
				{Name: "StringSlice", YAMLKey: "string_slice", Path: []string{"string_slice"}, Type: schema.TypeStringSlice},
				{Name: "IntSlice", YAMLKey: "int_slice", Path: []string{"int_slice"}, Type: schema.TypeIntSlice},
				{Name: "StringMap", YAMLKey: "string_map", Path: []string{"string_map"}, Type: schema.TypeStringMap},
			},
			Children: []*schema.Node{
				{
					Name:    "Nested",
					YAMLKey: "nested",
					Path:    []string{"nested"},
					Fields: []*schema.Field{
						{Name: "Value", YAMLKey: "value", Path: []string{"nested", "value"}, Type: schema.TypeInt},
					},
				},
			},
		},
	}
	return ast
}

func TestApplyEnvOverrides(t *testing.T) {
	ast := buildTestEnvAST()

	// Ensure clean environment before tests
	cleanup := func() {
		os.Unsetenv("CONFIGFORGE_BOOLEAN_FIELD")
		os.Unsetenv("CONFIGFORGE_INTEGER_FIELD")
		os.Unsetenv("CONFIGFORGE_FLOAT_FIELD")
		os.Unsetenv("CONFIGFORGE_STRING_FIELD")
		os.Unsetenv("CONFIGFORGE_STRING_SLICE")
		os.Unsetenv("CONFIGFORGE_INT_SLICE")
		os.Unsetenv("CONFIGFORGE_STRING_MAP")
		os.Unsetenv("CONFIGFORGE_NESTED_VALUE")
	}
	t.Cleanup(cleanup)

	t.Run("no relevant env vars set", func(t *testing.T) {
		cleanup()
		raw := map[string]any{
			"string_field": "yaml",
		}
		rawWithOverrides, err := ApplyEnvOverrides(ast, raw, raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rawWithOverrides["string_field"] != "yaml" {
			t.Errorf("expected string_field unchanged, got %v", rawWithOverrides["string_field"])
		}
	})

	t.Run("env overrides YAML value", func(t *testing.T) {
		cleanup()
		os.Setenv("CONFIGFORGE_STRING_FIELD", "env")
		os.Setenv("CONFIGFORGE_NESTED_VALUE", "42")

		raw := map[string]any{
			"string_field": "yaml",
			"nested": map[string]any{
				"value": 10,
			},
		}

		rawWithOverrides, err := ApplyEnvOverrides(ast, raw, raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rawWithOverrides["string_field"] != "env" {
			t.Errorf("expected string_field overridden to 'env', got %v", rawWithOverrides["string_field"])
		}

		nested := convertToMapStringAny(rawWithOverrides["nested"])
		if nested["value"] != 42 {
			t.Errorf("expected nested.value overridden to 42, got %v", nested["value"])
		}
	})

	t.Run("env supplies value for absent field", func(t *testing.T) {
		cleanup()
		os.Setenv("CONFIGFORGE_INTEGER_FIELD", "100")

		raw := make(map[string]any)
		rawWithOverrides, err := ApplyEnvOverrides(ast, raw, raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rawWithOverrides["integer_field"] != 100 {
			t.Errorf("expected integer_field to be set to 100, got %v", rawWithOverrides["integer_field"])
		}
	})

	t.Run("slice and map parsing", func(t *testing.T) {
		cleanup()
		os.Setenv("CONFIGFORGE_STRING_SLICE", "x , y, z")
		os.Setenv("CONFIGFORGE_INT_SLICE", "1,2,3")
		os.Setenv("CONFIGFORGE_STRING_MAP", "a=1, b=2")

		raw := make(map[string]any)
		rawWithOverrides, err := ApplyEnvOverrides(ast, raw, raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !reflect.DeepEqual(rawWithOverrides["string_slice"], []string{"x", "y", "z"}) {
			t.Errorf("expected string_slice [x y z], got %v", rawWithOverrides["string_slice"])
		}
		if !reflect.DeepEqual(rawWithOverrides["int_slice"], []int{1, 2, 3}) {
			t.Errorf("expected int_slice [1 2 3], got %v", rawWithOverrides["int_slice"])
		}
		if !reflect.DeepEqual(rawWithOverrides["string_map"], map[string]string{"a": "1", "b": "2"}) {
			t.Errorf("expected string_map map[a:1 b:2], got %v", rawWithOverrides["string_map"])
		}
	})

	t.Run("malformed values return typed error", func(t *testing.T) {
		cleanup()
		os.Setenv("CONFIGFORGE_BOOLEAN_FIELD", "invalid_bool")

		raw := make(map[string]any)
		_, err := ApplyEnvOverrides(ast, raw, raw)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		var parseErr *EnvParseError
		if !errors.As(err, &parseErr) {
			t.Errorf("expected EnvParseError, got %T: %v", err, err)
		}
		if parseErr.EnvVar != "CONFIGFORGE_BOOLEAN_FIELD" || parseErr.ExpectedType != "bool" {
			t.Errorf("mismatched error details: %+v", parseErr)
		}
	})
}

func TestE2ESmoke(t *testing.T) {
	t.Cleanup(func() {
		os.Unsetenv("CONFIGFORGE_CONFIG_PATH")
		os.Unsetenv("CONFIGFORGE_INSTRUMENTATION_HTTP_PORT")
	})

	os.Setenv("CONFIGFORGE_CONFIG_PATH", "../../examples/http-server/config_minimal.yaml")
	os.Setenv("CONFIGFORGE_INSTRUMENTATION_HTTP_PORT", "9090")

	rawMeta, err := parser.ParseFile("../../examples/http-server/metadata.yaml")
	if err != nil {
		t.Fatalf("failed to parse metadata: %v", err)
	}

	ast, err := schema.Build(rawMeta)
	if err != nil {
		t.Fatalf("failed to build AST: %v", err)
	}

	type httpConfig struct {
		Instrumentation struct {
			Http struct {
				Port           int      `yaml:"port"`
				Enabled        bool     `yaml:"enabled"`
				Host           string   `yaml:"host"`
				Timeout        int      `yaml:"timeout"`
				LogLevel       string   `yaml:"log_level"`
				ApiKey         string   `yaml:"api_key"`
				CaptureHeaders []string `yaml:"capture_headers"`
				RedactQuery    []string `yaml:"redact_query"`
			} `yaml:"http"`
		} `yaml:"instrumentation"`
	}

	cfg, rawConfig, err := LoadAndPrepare[httpConfig](ast)
	if err != nil {
		t.Fatalf("failed to load and prepare: %v", err)
	}

	valErrs := validator.Validate(ast, rawConfig)
	if len(valErrs) > 0 {
		t.Fatalf("validation failed unexpectedly: %v", valErrs)
	}

	if cfg.Instrumentation.Http.Port != 9090 {
		t.Errorf("expected Port 9090, got %d", cfg.Instrumentation.Http.Port)
	}
	if cfg.Instrumentation.Http.Host != "127.0.0.1" {
		t.Errorf("expected Host 127.0.0.1, got %s", cfg.Instrumentation.Http.Host)
	}
	if cfg.Instrumentation.Http.Enabled != true {
		t.Errorf("expected Enabled true, got %v", cfg.Instrumentation.Http.Enabled)
	}
	if cfg.Instrumentation.Http.ApiKey != "MINIMALAPIKEY123" {
		t.Errorf("expected ApiKey MINIMALAPIKEY123, got %s", cfg.Instrumentation.Http.ApiKey)
	}

	t.Logf("E2E Validation Passed!")
	t.Logf("HTTP Server Port: %d (overridden via env)", cfg.Instrumentation.Http.Port)
	t.Logf("HTTP Server Host: %s (default)", cfg.Instrumentation.Http.Host)
	t.Logf("HTTP Server Enabled: %v (default)", cfg.Instrumentation.Http.Enabled)
	t.Logf("HTTP Server ApiKey: %s (configured)", cfg.Instrumentation.Http.ApiKey)
}


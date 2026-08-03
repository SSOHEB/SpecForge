package runtime

import (
	"reflect"
	"testing"

	"configforge/internal/schema"
)

type testDefaultsConfig struct {
	BooleanField   bool              `yaml:"boolean_field"`
	StringField    string            `yaml:"string_field"`
	IntegerField   int               `yaml:"integer_field"`
	FloatField     float64           `yaml:"float_field"`
	StringSlice    []string          `yaml:"string_slice"`
	IntSlice       []int             `yaml:"int_slice"`
	StringMap      map[string]string `yaml:"string_map"`
	NoDefaultField string            `yaml:"no_default_field"`
	Nested         struct {
		Enabled bool `yaml:"enabled"`
		Port    int  `yaml:"port"`
	} `yaml:"nested"`
}

func buildTestDefaultsAST() *schema.AST {
	minPort := float64(1)
	maxPort := float64(65535)

	ast := &schema.AST{
		Root: &schema.Node{
			Name: "Config",
			Fields: []*schema.Field{
				{Name: "BooleanField", YAMLKey: "boolean_field", Type: schema.TypeBool, Default: true},
				{Name: "StringField", YAMLKey: "string_field", Type: schema.TypeString, Default: "hello"},
				{Name: "IntegerField", YAMLKey: "integer_field", Type: schema.TypeInt, Default: 123},
				{Name: "FloatField", YAMLKey: "float_field", Type: schema.TypeFloat, Default: 3.14},
				{Name: "StringSlice", YAMLKey: "string_slice", Type: schema.TypeStringSlice, Default: []string{"a", "b"}},
				{Name: "IntSlice", YAMLKey: "int_slice", Type: schema.TypeIntSlice, Default: []int{1, 2}},
				{Name: "StringMap", YAMLKey: "string_map", Type: schema.TypeStringMap, Default: map[string]string{"k": "v"}},
				{Name: "NoDefaultField", YAMLKey: "no_default_field", Type: schema.TypeString, Default: nil},
			},
			Children: []*schema.Node{
				{
					Name:    "Nested",
					YAMLKey: "nested",
					Path:    []string{"nested"},
					Fields: []*schema.Field{
						{Name: "Enabled", YAMLKey: "enabled", Path: []string{"nested", "enabled"}, Type: schema.TypeBool, Default: true},
						{Name: "Port", YAMLKey: "port", Path: []string{"nested", "port"}, Type: schema.TypeInt, Default: 80, Min: &minPort, Max: &maxPort},
					},
				},
			},
		},
	}
	return ast
}

func TestApplyDefaults(t *testing.T) {
	ast := buildTestDefaultsAST()

	t.Run("fields entirely absent", func(t *testing.T) {
		raw := make(map[string]any)
		rawWithDefaults := ApplyDefaults(ast, raw)

		// Assert raw map values
		if rawWithDefaults["boolean_field"] != true {
			t.Errorf("expected boolean_field true, got %v", rawWithDefaults["boolean_field"])
		}
		if rawWithDefaults["string_field"] != "hello" {
			t.Errorf("expected string_field 'hello', got %v", rawWithDefaults["string_field"])
		}
		if rawWithDefaults["integer_field"] != 123 {
			t.Errorf("expected integer_field 123, got %v", rawWithDefaults["integer_field"])
		}
		if rawWithDefaults["float_field"] != 3.14 {
			t.Errorf("expected float_field 3.14, got %v", rawWithDefaults["float_field"])
		}
		if !reflect.DeepEqual(rawWithDefaults["string_slice"], []string{"a", "b"}) {
			t.Errorf("expected string_slice [a b], got %v", rawWithDefaults["string_slice"])
		}
		if !reflect.DeepEqual(rawWithDefaults["int_slice"], []int{1, 2}) {
			t.Errorf("expected int_slice [1 2], got %v", rawWithDefaults["int_slice"])
		}
		if !reflect.DeepEqual(rawWithDefaults["string_map"], map[string]string{"k": "v"}) {
			t.Errorf("expected string_map map[k:v], got %v", rawWithDefaults["string_map"])
		}
		if _, exists := rawWithDefaults["no_default_field"]; exists {
			t.Errorf("expected no_default_field to be absent, got %v", rawWithDefaults["no_default_field"])
		}

		// Assert nested namespace defaults
		nestedVal, exists := rawWithDefaults["nested"]
		if !exists {
			t.Fatalf("missing nested map")
		}
		nestedMap := convertToMapStringAny(nestedVal)
		if nestedMap["enabled"] != true {
			t.Errorf("expected nested.enabled true, got %v", nestedMap["enabled"])
		}
		if nestedMap["port"] != 80 {
			t.Errorf("expected nested.port 80, got %v", nestedMap["port"])
		}

		// Convert to typed struct
		cfg, err := convertMapToStruct[testDefaultsConfig](rawWithDefaults)
		if err != nil {
			t.Fatalf("failed to convert to struct: %v", err)
		}

		if cfg.BooleanField != true {
			t.Errorf("cfg: expected BooleanField true")
		}
		if cfg.StringField != "hello" {
			t.Errorf("cfg: expected StringField 'hello'")
		}
		if cfg.IntegerField != 123 {
			t.Errorf("cfg: expected IntegerField 123")
		}
		if cfg.FloatField != 3.14 {
			t.Errorf("cfg: expected FloatField 3.14")
		}
		if !reflect.DeepEqual(cfg.StringSlice, []string{"a", "b"}) {
			t.Errorf("cfg: expected StringSlice [a b]")
		}
		if !reflect.DeepEqual(cfg.IntSlice, []int{1, 2}) {
			t.Errorf("cfg: expected IntSlice [1 2]")
		}
		if cfg.StringMap["k"] != "v" {
			t.Errorf("cfg: expected StringMap[k] 'v'")
		}
		if cfg.NoDefaultField != "" {
			t.Errorf("cfg: expected NoDefaultField empty, got %q", cfg.NoDefaultField)
		}
		if cfg.Nested.Enabled != true {
			t.Errorf("cfg: expected Nested.Enabled true")
		}
		if cfg.Nested.Port != 80 {
			t.Errorf("cfg: expected Nested.Port 80")
		}
	})

	t.Run("explicit zero values survive", func(t *testing.T) {
		raw := map[string]any{
			"boolean_field": false,
			"nested": map[string]any{
				"enabled": false,
			},
		}

		rawWithDefaults := ApplyDefaults(ast, raw)

		// Assert raw map values
		if rawWithDefaults["boolean_field"] != false {
			t.Errorf("expected boolean_field false to survive, got %v", rawWithDefaults["boolean_field"])
		}

		nestedVal := rawWithDefaults["nested"]
		nestedMap := convertToMapStringAny(nestedVal)
		if nestedMap["enabled"] != false {
			t.Errorf("expected nested.enabled false to survive, got %v", nestedMap["enabled"])
		}
		if nestedMap["port"] != 80 {
			t.Errorf("expected nested.port default to be applied, got %v", nestedMap["port"])
		}

		// Convert to typed struct
		cfg, err := convertMapToStruct[testDefaultsConfig](rawWithDefaults)
		if err != nil {
			t.Fatalf("failed to convert to struct: %v", err)
		}

		if cfg.BooleanField != false {
			t.Errorf("cfg: expected BooleanField false")
		}
		if cfg.Nested.Enabled != false {
			t.Errorf("cfg: expected Nested.Enabled false")
		}
		if cfg.Nested.Port != 80 {
			t.Errorf("cfg: expected Nested.Port default 80")
		}
	})
}

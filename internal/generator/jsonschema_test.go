package generator

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/SSOHEB/SpecForge/internal/parser"
	"github.com/SSOHEB/SpecForge/internal/schema"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

func TestJSONSchemaGeneration(t *testing.T) {
	// Parse metadata
	raw, err := parser.ParseFile("../../examples/http-server/metadata.yaml")
	if err != nil {
		t.Fatalf("failed to parse metadata.yaml: %v", err)
	}

	// Build AST
	ast, err := schema.Build(raw)
	if err != nil {
		t.Fatalf("failed to build AST: %v", err)
	}

	// Generate JSON Schema
	schemaBytes, err := GenerateJSONSchema(ast)
	if err != nil {
		t.Fatalf("failed to generate JSON Schema: %v", err)
	}

	// 1. Verify exact structure of generated schema
	var schemaMap map[string]any
	if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
		t.Fatalf("failed to unmarshal generated schema: %v", err)
	}

	// Verify $schema property
	if schemaMap["$schema"] != "http://json-schema.org/draft-07/schema#" {
		t.Errorf("unexpected $schema value: %v", schemaMap["$schema"])
	}

	// Traverse to: instrumentation.http.port
	instr, ok := schemaMap["properties"].(map[string]any)["instrumentation"].(map[string]any)
	if !ok {
		t.Fatalf("missing instrumentation in schema properties")
	}
	httpNode, ok := instr["properties"].(map[string]any)["http"].(map[string]any)
	if !ok {
		t.Fatalf("missing http inside instrumentation")
	}
	portField, ok := httpNode["properties"].(map[string]any)["port"].(map[string]any)
	if !ok {
		t.Fatalf("missing port inside http")
	}

	if portField["type"] != "integer" {
		t.Errorf("expected port type to be 'integer', got %v", portField["type"])
	}
	if portField["minimum"].(float64) != 1 {
		t.Errorf("expected port minimum to be 1, got %v", portField["minimum"])
	}
	if portField["maximum"].(float64) != 65535 {
		t.Errorf("expected port maximum to be 65535, got %v", portField["maximum"])
	}

	// Traverse to: instrumentation.http.log_level
	logLevelField, ok := httpNode["properties"].(map[string]any)["log_level"].(map[string]any)
	if !ok {
		t.Fatalf("missing log_level inside http")
	}
	if logLevelField["type"] != "string" {
		t.Errorf("expected log_level type to be 'string', got %v", logLevelField["type"])
	}
	enumSlice, ok := logLevelField["enum"].([]any)
	if !ok || len(enumSlice) != 4 || enumSlice[0].(string) != "debug" {
		t.Errorf("expected log_level enum values, got %v", logLevelField["enum"])
	}

	// 2. Deterministic output verification (byte-identical)
	schemaBytes2, err := GenerateJSONSchema(ast)
	if err != nil {
		t.Fatalf("failed to generate JSON Schema again: %v", err)
	}
	if !bytes.Equal(schemaBytes, schemaBytes2) {
		t.Errorf("schema generation is not deterministic (output mismatch)")
	}

	// 3. Validate actual valid config.yaml passes validation
	configBytes, err := os.ReadFile("../../examples/http-server/config.yaml")
	if err != nil {
		t.Fatalf("failed to read config.yaml: %v", err)
	}

	var rawConfig any
	if err := yaml.Unmarshal(configBytes, &rawConfig); err != nil {
		t.Fatalf("failed to parse config.yaml: %v", err)
	}

	// Convert YAML mapping to JSON compatible representation (map[string]any instead of map[any]any)
	jsonConfig := convertYAMLToJSON(rawConfig)

	// Validate config using santhosh-tekuri/jsonschema/v5
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", bytes.NewReader(schemaBytes)); err != nil {
		t.Fatalf("failed to add schema resource: %v", err)
	}
	compiledSchema, err := compiler.Compile("schema.json")
	if err != nil {
		t.Fatalf("failed to compile schema: %v", err)
	}

	if err := compiledSchema.Validate(jsonConfig); err != nil {
		t.Fatalf("valid config failed schema validation: %v", err)
	}

	// 4. Validate invalid configs fail validation
	// Case A: invalid port
	invalidPortConfigYAML := `
instrumentation:
  http:
    port: 99999
`
	var rawPortConfig any
	if err := yaml.Unmarshal([]byte(invalidPortConfigYAML), &rawPortConfig); err != nil {
		t.Fatalf("failed to parse invalid port config: %v", err)
	}
	jsonPortConfig := convertYAMLToJSON(rawPortConfig)
	errPort := compiledSchema.Validate(jsonPortConfig)
	if errPort == nil {
		t.Fatalf("expected validation error for invalid port config, got nil")
	}
	if !strings.Contains(errPort.Error(), "minimum") && !strings.Contains(errPort.Error(), "maximum") && !strings.Contains(errPort.Error(), "99999") {
		t.Errorf("expected error string to mention port out of bounds, got: %s", errPort.Error())
	}

	// Case B: invalid log_level
	invalidLogLevelConfigYAML := `
instrumentation:
  http:
    log_level: "bogus"
`
	var rawLogLevelConfig any
	if err := yaml.Unmarshal([]byte(invalidLogLevelConfigYAML), &rawLogLevelConfig); err != nil {
		t.Fatalf("failed to parse invalid log_level config: %v", err)
	}
	jsonLogLevelConfig := convertYAMLToJSON(rawLogLevelConfig)
	errLogLevel := compiledSchema.Validate(jsonLogLevelConfig)
	if errLogLevel == nil {
		t.Fatalf("expected validation error for invalid log level config, got nil")
	}
	if !strings.Contains(errLogLevel.Error(), "enum") && !strings.Contains(errLogLevel.Error(), "bogus") {
		t.Errorf("expected error string to mention enum, got: %s", errLogLevel.Error())
	}

	// 5. Empty AST produces minimal valid empty schema
	emptySchemaBytes, err := GenerateJSONSchema(nil)
	if err != nil {
		t.Fatalf("failed to generate empty schema: %v", err)
	}
	var emptyMap map[string]any
	if err := json.Unmarshal(emptySchemaBytes, &emptyMap); err != nil {
		t.Fatalf("failed to unmarshal empty schema: %v", err)
	}
	if emptyMap["$schema"] != "http://json-schema.org/draft-07/schema#" || emptyMap["type"] != "object" {
		t.Errorf("unexpected empty schema format: %v", emptyMap)
	}
}

func convertYAMLToJSON(val any) any {
	switch v := val.(type) {
	case map[string]any:
		m := make(map[string]any)
		for k, valVal := range v {
			m[k] = convertYAMLToJSON(valVal)
		}
		return m
	case map[any]any:
		m := make(map[string]any)
		for k, valVal := range v {
			m[strings.TrimSpace(k.(string))] = convertYAMLToJSON(valVal)
		}
		return m
	case []any:
		s := make([]any, len(v))
		for i, item := range v {
			s[i] = convertYAMLToJSON(item)
		}
		return s
	default:
		return v
	}
}

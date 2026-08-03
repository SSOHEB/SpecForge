package runtime

import (
	"fmt"
	"os"

	"configforge/internal/schema"
	"gopkg.in/yaml.v3"
)

// ApplyDefaults walks the AST and inserts declared default values into raw if absent.
func ApplyDefaults(ast *schema.AST, raw map[string]any) map[string]any {
	if ast == nil || ast.Root == nil {
		return raw
	}
	if raw == nil {
		raw = make(map[string]any)
	}
	applyNodeDefaults(ast.Root, raw)
	return raw
}

func applyNodeDefaults(node *schema.Node, raw map[string]any) {
	// Apply defaults for fields in this node
	for _, f := range node.Fields {
		if f.Default == nil {
			continue
		}
		if _, exists := raw[f.YAMLKey]; !exists {
			raw[f.YAMLKey] = f.Default
		}
	}

	// Apply defaults for nested namespaces (Nodes)
	for _, child := range node.Children {
		childVal, exists := raw[child.YAMLKey]
		if !exists || childVal == nil {
			childMap := make(map[string]any)
			raw[child.YAMLKey] = childMap
			applyNodeDefaults(child, childMap)
		} else {
			childMap := convertToMapStringAny(childVal)
			if childMap != nil {
				applyNodeDefaults(child, childMap)
				raw[child.YAMLKey] = childMap
			}
		}
	}
}

// convertMapToStruct converts raw map structure into typed Go struct using intermediate YAML serialization.
// This is the simplest and safest approach as it uses the same yaml structure tags and unmarshaling
// behavior defined on the generated Config types, completely avoiding custom reflection.
func convertMapToStruct[T any](raw map[string]any) (*T, error) {
	bytes, err := yaml.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal raw config with defaults: %w", err)
	}
	var cfg T
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config into typed struct: %w", err)
	}
	return &cfg, nil
}

// LoadAndPrepareFile loads raw config, applies defaults, overrides with environment variables, and unmarshals into a typed struct.
func LoadAndPrepareFile[T any](ast *schema.AST, path string) (*T, map[string]any, error) {
	raw, err := LoadFile[map[string]any](path)
	if err != nil {
		// If file is empty, we still apply all defaults from AST.
		if _, ok := err.(*ConfigEmptyError); ok {
			emptyMap := make(map[string]any)
			raw = &emptyMap
		} else {
			return nil, nil, err
		}
	}

	// Make a deep copy to preserve the original YAML map structure
	originalRaw := copyMap(*raw)

	// Apply defaults
	rawWithDefaults := ApplyDefaults(ast, *raw)

	// Apply env overrides
	rawWithOverrides, err := ApplyEnvOverrides(ast, rawWithDefaults, originalRaw)
	if err != nil {
		return nil, nil, err
	}

	cfg, err := convertMapToStruct[T](rawWithOverrides)
	if err != nil {
		return nil, nil, err
	}

	return cfg, rawWithOverrides, nil
}

func copyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := make(map[string]any)
	for k, v := range src {
		if m := convertToMapStringAny(v); m != nil {
			dst[k] = copyMap(m)
		} else if slice, ok := v.([]any); ok {
			dst[k] = copySlice(slice)
		} else {
			dst[k] = v
		}
	}
	return dst
}

func copySlice(src []any) []any {
	if src == nil {
		return nil
	}
	dst := make([]any, len(src))
	for i, v := range src {
		if m := convertToMapStringAny(v); m != nil {
			dst[i] = copyMap(m)
		} else if slice, ok := v.([]any); ok {
			dst[i] = copySlice(slice)
		} else {
			dst[i] = v
		}
	}
	return dst
}


// LoadAndPrepare loads raw config from default CONFIGFORGE_CONFIG_PATH, applies defaults, and unmarshals.
func LoadAndPrepare[T any](ast *schema.AST) (*T, map[string]any, error) {
	path := os.Getenv("CONFIGFORGE_CONFIG_PATH")
	if path == "" {
		path = "./config.yaml"
	}
	return LoadAndPrepareFile[T](ast, path)
}

func convertToMapStringAny(val any) map[string]any {
	if m, ok := val.(map[string]any); ok {
		return m
	}
	if m, ok := val.(map[any]any); ok {
		res := make(map[string]any)
		for k, v := range m {
			res[fmt.Sprintf("%v", k)] = v
		}
		return res
	}
	return nil
}

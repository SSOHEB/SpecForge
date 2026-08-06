package runtime

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/SSOHEB/codrao/internal/schema"
)

// EnvOverridesPrecedeYAML controls the precedence order between environment variables and YAML values.
// By default, it is set to true (environment variables override YAML values).
// Setting this to false changes the precedence so that YAML values take precedence over environment variables.
var EnvOverridesPrecedeYAML = true

// ApplyEnvOverrides walks the AST, checks for prefixed env vars corresponding
// to each field's path, parses them, and overwrites the config values in raw.
func ApplyEnvOverrides(ast *schema.AST, raw map[string]any, originalRaw map[string]any, prefix string) (map[string]any, error) {
	if ast == nil || ast.Root == nil {
		return raw, nil
	}
	if raw == nil {
		raw = make(map[string]any)
	}

	var walk func(node *schema.Node) error
	walk = func(node *schema.Node) error {
		for _, f := range node.Fields {
			// Construct prefixed uppercase env var name
			envParts := make([]string, len(f.Path))
			for i, p := range f.Path {
				envParts[i] = strings.ToUpper(p)
			}
			envName := prefix + strings.Join(envParts, "_")

			envVal, exists := os.LookupEnv(envName)
			if !exists {
				continue
			}

			// If YAML has precedence and key exists in original YAML, do not override
			if !EnvOverridesPrecedeYAML && hasValueAtPath(originalRaw, f.Path) {
				continue
			}

			parsedVal, err := parseEnvVal(envName, envVal, f.Type)
			if err != nil {
				return err
			}

			writeValueAtPath(raw, f.Path, parsedVal)
		}

		for _, child := range node.Children {
			if err := walk(child); err != nil {
				return err
			}
		}
		return nil
	}

	if err := walk(ast.Root); err != nil {
		return nil, err
	}

	return raw, nil
}

func parseEnvVal(envName, envVal string, ft schema.FieldType) (any, error) {
	switch ft {
	case schema.TypeBool:
		s := strings.ToLower(strings.TrimSpace(envVal))
		if s == "true" || s == "1" {
			return true, nil
		} else if s == "false" || s == "0" {
			return false, nil
		}
		return nil, &EnvParseError{EnvVar: envName, Value: envVal, ExpectedType: "bool"}

	case schema.TypeInt:
		v, err := strconv.Atoi(strings.TrimSpace(envVal))
		if err != nil {
			return nil, &EnvParseError{EnvVar: envName, Value: envVal, ExpectedType: "int", Err: err}
		}
		return v, nil

	case schema.TypeFloat:
		v, err := strconv.ParseFloat(strings.TrimSpace(envVal), 64)
		if err != nil {
			return nil, &EnvParseError{EnvVar: envName, Value: envVal, ExpectedType: "float", Err: err}
		}
		return v, nil

	case schema.TypeString:
		return envVal, nil

	case schema.TypeStringSlice:
		parts := strings.Split(envVal, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		return parts, nil

	case schema.TypeIntSlice:
		parts := strings.Split(envVal, ",")
		intParts := make([]int, len(parts))
		for i, part := range parts {
			v, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil {
				return nil, &EnvParseError{EnvVar: envName, Value: envVal, ExpectedType: "int[]", Err: err}
			}
			intParts[i] = v
		}
		return intParts, nil

	case schema.TypeStringMap:
		parts := strings.Split(envVal, ",")
		m := make(map[string]string)
		for _, part := range parts {
			if strings.TrimSpace(part) == "" {
				continue
			}
			kv := strings.SplitN(part, "=", 2)
			if len(kv) != 2 {
				return nil, &EnvParseError{
					EnvVar:       envName,
					Value:        envVal,
					ExpectedType: "map[string]string",
					Err:          fmt.Errorf("invalid key-value pair format: %q", part),
				}
			}
			m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
		return m, nil

	default:
		return nil, fmt.Errorf("unsupported field type: %v", ft)
	}
}

func writeValueAtPath(raw map[string]any, path []string, val any) {
	if len(path) == 0 {
		return
	}
	current := raw
	for i := 0; i < len(path)-1; i++ {
		key := path[i]
		next, exists := current[key]
		if !exists || next == nil {
			nextMap := make(map[string]any)
			current[key] = nextMap
			current = nextMap
		} else {
			nextMap := convertToMapStringAny(next)
			if nextMap == nil {
				nextMap = make(map[string]any)
				current[key] = nextMap
			}
			current = nextMap
		}
	}
	lastKey := path[len(path)-1]
	current[lastKey] = val
}

func hasValueAtPath(raw map[string]any, path []string) bool {
	if len(path) == 0 {
		return false
	}
	current := raw
	for i := 0; i < len(path)-1; i++ {
		key := path[i]
		next, exists := current[key]
		if !exists || next == nil {
			return false
		}
		nextMap := convertToMapStringAny(next)
		if nextMap == nil {
			return false
		}
		current = nextMap
	}
	lastKey := path[len(path)-1]
	_, exists := current[lastKey]
	return exists
}

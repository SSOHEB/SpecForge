package schema

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/SSOHEB/codrao/internal/parser"
)

// Build recursively walks raw parser metadata and constructs the normalized AST.
func Build(raw *parser.RawMetadata) (*AST, error) {
	if raw == nil {
		return nil, fmt.Errorf("nil metadata")
	}

	root := &Node{
		Name: "Config",
		Path: []string{},
	}

	keys := make([]string, 0, len(raw.Fields))
	for k := range raw.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	goNameMap := make(map[string][]string)

	for _, k := range keys {
		rawNode := raw.Fields[k]
		childNode, childField, err := buildNode(k, rawNode, []string{})
		if err != nil {
			return nil, err
		}
		if childField != nil {
			root.Fields = append(root.Fields, childField)
			goNameMap[childField.Name] = append(goNameMap[childField.Name], k)
		}
		if childNode != nil {
			root.Children = append(root.Children, childNode)
			goNameMap[childNode.Name] = append(goNameMap[childNode.Name], k)
		}
	}

	for resolvedName, originalKeys := range goNameMap {
		if len(originalKeys) > 1 {
			sort.Strings(originalKeys)
			return nil, &SiblingNameCollisionError{
				Path:   "",
				GoName: resolvedName,
				Keys:   originalKeys,
			}
		}
	}

	return &AST{Root: root}, nil
}

func buildNode(yamlKey string, rawNode *parser.RawNode, parentPath []string) (*Node, *Field, error) {
	goName := ToGoName(yamlKey)
	path := append(parentPath, yamlKey)
	dottedPath := strings.Join(path, ".")

	if rawNode.Children != nil {
		node := &Node{
			Name:    goName,
			YAMLKey: yamlKey,
			Path:    path,
		}

		keys := make([]string, 0, len(rawNode.Children))
		for k := range rawNode.Children {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		goNameMap := make(map[string][]string)

		for _, k := range keys {
			childRaw := rawNode.Children[k]
			childNode, childField, err := buildNode(k, childRaw, path)
			if err != nil {
				return nil, nil, err
			}
			if childField != nil {
				node.Fields = append(node.Fields, childField)
				goNameMap[childField.Name] = append(goNameMap[childField.Name], k)
			}
			if childNode != nil {
				node.Children = append(node.Children, childNode)
				goNameMap[childNode.Name] = append(goNameMap[childNode.Name], k)
			}
		}

		for resolvedName, originalKeys := range goNameMap {
			if len(originalKeys) > 1 {
				sort.Strings(originalKeys)
				return nil, nil, &SiblingNameCollisionError{
					Path:   dottedPath,
					GoName: resolvedName,
					Keys:   originalKeys,
				}
			}
		}

		return node, nil, nil
	}

	var ft FieldType
	switch *rawNode.Type {
	case "bool":
		ft = TypeBool
	case "string":
		ft = TypeString
	case "int":
		ft = TypeInt
	case "float":
		ft = TypeFloat
	case "string[]":
		ft = TypeStringSlice
	case "int[]":
		ft = TypeIntSlice
	case "map[string]string":
		ft = TypeStringMap
	default:
		return nil, nil, fmt.Errorf("unexpected type: %s", *rawNode.Type)
	}

	field := &Field{
		Name:        goName,
		YAMLKey:     yamlKey,
		Path:        path,
		Type:        ft,
		Description: rawNode.Description,
	}

	if rawNode.Required != nil {
		field.Required = *rawNode.Required
	}
	field.Min = rawNode.Min
	field.Max = rawNode.Max

	if len(rawNode.Enum) > 0 {
		field.Enum = make([]string, len(rawNode.Enum))
		for i, ev := range rawNode.Enum {
			s, ok := ev.(string)
			if !ok {
				return nil, nil, &DefaultTypeMismatchError{
					Path:     dottedPath,
					Expected: "string enum value",
					Actual:   fmt.Sprintf("%T", ev),
				}
			}
			field.Enum[i] = s
		}
	}

	if rawNode.Pattern != nil {
		field.Pattern = *rawNode.Pattern
		if _, err := regexp.Compile(field.Pattern); err != nil {
			return nil, nil, &InvalidRegexPatternError{
				Path:    dottedPath,
				Pattern: field.Pattern,
				Err:     err,
			}
		}
	}

	if rawNode.Default != nil {
		typedDefault, err := normalizeDefault(rawNode.Default, ft)
		if err != nil {
			return nil, nil, &DefaultTypeMismatchError{
				Path:     dottedPath,
				Expected: ft.String(),
				Actual:   fmt.Sprintf("%T", rawNode.Default),
				Err:      err,
			}
		}
		field.Default = typedDefault
	}

	return nil, field, nil
}

func normalizeDefault(val interface{}, ft FieldType) (interface{}, error) {
	switch ft {
	case TypeBool:
		b, ok := val.(bool)
		if !ok {
			return nil, fmt.Errorf("value is not a bool")
		}
		return b, nil

	case TypeString:
		s, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("value is not a string")
		}
		return s, nil

	case TypeInt:
		switch v := val.(type) {
		case int:
			return v, nil
		case int64:
			return int(v), nil
		case float64:
			if v == float64(int(v)) {
				return int(v), nil
			}
			return nil, fmt.Errorf("value is a float, expected int")
		default:
			return nil, fmt.Errorf("value of type %T is not an int", val)
		}

	case TypeFloat:
		switch v := val.(type) {
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		case int64:
			return float64(v), nil
		default:
			return nil, fmt.Errorf("value of type %T is not a float", val)
		}

	case TypeStringSlice:
		slice, ok := val.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected slice, got %T", val)
		}
		res := make([]string, len(slice))
		for i, item := range slice {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("expected string in slice at index %d, got %T", i, item)
			}
			res[i] = s
		}
		return res, nil

	case TypeIntSlice:
		slice, ok := val.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected slice, got %T", val)
		}
		res := make([]int, len(slice))
		for i, item := range slice {
			var intVal int
			switch v := item.(type) {
			case int:
				intVal = v
			case int64:
				intVal = int(v)
			case float64:
				if v == float64(int(v)) {
					intVal = int(v)
				} else {
					return nil, fmt.Errorf("expected int at index %d, got float %f", i, v)
				}
			default:
				return nil, fmt.Errorf("expected int at index %d, got %T", i, item)
			}
			res[i] = intVal
		}
		return res, nil

	case TypeStringMap:
		res := make(map[string]string)
		if m, ok := val.(map[string]interface{}); ok {
			for k, v := range m {
				s, ok := v.(string)
				if !ok {
					return nil, fmt.Errorf("expected string value for key %q, got %T", k, v)
				}
				res[k] = s
			}
			return res, nil
		} else if m, ok := val.(map[interface{}]interface{}); ok {
			for k, v := range m {
				kStr, ok1 := k.(string)
				vStr, ok2 := v.(string)
				if !ok1 || !ok2 {
					return nil, fmt.Errorf("expected string key and value, got %T: %T", k, v)
				}
				res[kStr] = vStr
			}
			return res, nil
		}
		return nil, fmt.Errorf("expected map, got %T", val)

	default:
		return nil, fmt.Errorf("unknown field type: %v", ft)
	}
}

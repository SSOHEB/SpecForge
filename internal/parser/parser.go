package parser

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parse reads YAML from r and returns a RawMetadata struct after structural validation.
func Parse(r io.Reader) (*RawMetadata, error) {
	var root yaml.Node
	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&root); err != nil {
		if err == io.EOF {
			return &RawMetadata{Fields: make(map[string]*RawNode)}, nil
		}
		return nil, &InvalidYAMLSyntaxError{Err: err}
	}

	var contentNode *yaml.Node
	if root.Kind == yaml.DocumentNode {
		if len(root.Content) == 0 {
			return &RawMetadata{Fields: make(map[string]*RawNode)}, nil
		}
		contentNode = root.Content[0]
	} else {
		contentNode = &root
	}

	// An empty document or null document
	if contentNode.Kind == yaml.ScalarNode && (contentNode.Value == "" || contentNode.Value == "null") {
		return &RawMetadata{Fields: make(map[string]*RawNode)}, nil
	}

	if contentNode.Kind != yaml.MappingNode {
		return nil, &InvalidYAMLSyntaxError{Err: fmt.Errorf("expected mapping at root, got %v", contentNode.Kind)}
	}

	metadata := &RawMetadata{
		Fields: make(map[string]*RawNode),
	}

	// Iterate over top-level key-values
	for i := 0; i < len(contentNode.Content); i += 2 {
		keyNode := contentNode.Content[i]
		valNode := contentNode.Content[i+1]

		key := keyNode.Value
		node, err := parseNode(valNode, []string{key})
		if err != nil {
			return nil, err
		}
		metadata.Fields[key] = node
	}

	// Validate the parsed metadata
	if err := validateMetadata(metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}

// ParseFile parses the file at the given path.
func ParseFile(path string) (*RawMetadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return Parse(f)
}

func parseNode(node *yaml.Node, path []string) (*RawNode, error) {
	if node.Kind != yaml.MappingNode {
		dottedPath := strings.Join(path, ".")
		return nil, &ValidationError{
			Path: dottedPath,
			Msg:  "expected mapping node",
		}
	}

	// Map key-values of this node
	rawMap := make(map[string]*yaml.Node)
	for i := 0; i < len(node.Content); i += 2 {
		kNode := node.Content[i]
		vNode := node.Content[i+1]
		rawMap[kNode.Value] = vNode
	}

	rawNode := &RawNode{}

	if isLeafField(rawMap) {
		if tNode, ok := rawMap["type"]; ok {
			var t string
			if err := tNode.Decode(&t); err != nil {
				return nil, &ValidationError{Path: strings.Join(path, "."), Msg: fmt.Sprintf("invalid type value: %v", err)}
			}
			rawNode.Type = &t
		}
		if dNode, ok := rawMap["default"]; ok {
			var d interface{}
			if err := dNode.Decode(&d); err != nil {
				return nil, &ValidationError{Path: strings.Join(path, "."), Msg: fmt.Sprintf("invalid default value: %v", err)}
			}
			rawNode.Default = d
		}
		if descNode, ok := rawMap["description"]; ok {
			var desc string
			if err := descNode.Decode(&desc); err != nil {
				return nil, &ValidationError{Path: strings.Join(path, "."), Msg: fmt.Sprintf("invalid description value: %v", err)}
			}
			rawNode.Description = desc
		}
		if rNode, ok := rawMap["required"]; ok {
			var req bool
			if err := rNode.Decode(&req); err != nil {
				return nil, &ValidationError{Path: strings.Join(path, "."), Msg: fmt.Sprintf("invalid required value: %v", err)}
			}
			rawNode.Required = &req
		}
		if minNode, ok := rawMap["min"]; ok {
			var minVal float64
			if err := minNode.Decode(&minVal); err != nil {
				return nil, &ValidationError{Path: strings.Join(path, "."), Msg: fmt.Sprintf("invalid min value: %v", err)}
			}
			rawNode.Min = &minVal
		}
		if maxNode, ok := rawMap["max"]; ok {
			var maxVal float64
			if err := maxNode.Decode(&maxVal); err != nil {
				return nil, &ValidationError{Path: strings.Join(path, "."), Msg: fmt.Sprintf("invalid max value: %v", err)}
			}
			rawNode.Max = &maxVal
		}
		if enumNode, ok := rawMap["enum"]; ok {
			var enumVal []interface{}
			if err := enumNode.Decode(&enumVal); err != nil {
				return nil, &ValidationError{Path: strings.Join(path, "."), Msg: fmt.Sprintf("invalid enum value: %v", err)}
			}
			rawNode.Enum = enumVal
		}
		if pNode, ok := rawMap["pattern"]; ok {
			var p string
			if err := pNode.Decode(&p); err != nil {
				return nil, &ValidationError{Path: strings.Join(path, "."), Msg: fmt.Sprintf("invalid pattern value: %v", err)}
			}
			rawNode.Pattern = &p
		}
		return rawNode, nil
	}

	// Parse as namespace (children)
	rawNode.Children = make(map[string]*RawNode)
	for k, v := range rawMap {
		childPath := append(path, k)
		childNode, err := parseNode(v, childPath)
		if err != nil {
			return nil, err
		}
		rawNode.Children[k] = childNode
	}

	return rawNode, nil
}

func isLeafField(rawMap map[string]*yaml.Node) bool {
	if t, ok := rawMap["type"]; ok && t.Kind == yaml.ScalarNode {
		return true
	}
	hasKnownAttr := false
	for k, v := range rawMap {
		switch k {
		case "type":
			if v.Kind == yaml.ScalarNode {
				hasKnownAttr = true
			} else {
				return false
			}
		case "description", "pattern":
			if v.Kind == yaml.ScalarNode {
				hasKnownAttr = true
			} else {
				return false
			}
		case "required":
			if v.Kind == yaml.ScalarNode && (v.Value == "true" || v.Value == "false") {
				hasKnownAttr = true
			} else {
				return false
			}
		case "min", "max":
			if v.Kind == yaml.ScalarNode {
				hasKnownAttr = true
			} else {
				return false
			}
		case "enum":
			if v.Kind == yaml.SequenceNode {
				hasKnownAttr = true
			} else {
				return false
			}
		case "default":
			if v.Kind == yaml.MappingNode {
				return false
			}
			hasKnownAttr = true
		default:
			return false
		}
	}
	return hasKnownAttr
}

func isSupportedType(t string) bool {
	switch t {
	case "bool", "string", "int", "float", "string[]", "int[]", "map[string]string":
		return true
	default:
		return false
	}
}

func validateMetadata(metadata *RawMetadata) error {
	for k, v := range metadata.Fields {
		if err := validateNode(v, []string{k}); err != nil {
			return err
		}
	}
	return nil
}

func validateNode(node *RawNode, path []string) error {
	dottedPath := strings.Join(path, ".")

	if node.Children != nil {
		for k, v := range node.Children {
			if err := validateNode(v, append(path, k)); err != nil {
				return err
			}
		}
		return nil
	}

	if node.Type == nil {
		return &MissingTypeError{Path: dottedPath}
	}

	t := *node.Type
	if !isSupportedType(t) {
		return &UnknownTypeError{Path: dottedPath, Type: t}
	}

	if len(node.Enum) > 0 && t != "string" {
		return &InvalidAttributeCombinationError{
			Path: dottedPath,
			Msg:  fmt.Sprintf("enum is not valid for type %s", t),
		}
	}

	if node.Min != nil && t != "int" && t != "float" {
		return &InvalidAttributeCombinationError{
			Path: dottedPath,
			Msg:  fmt.Sprintf("min is not valid for type %s", t),
		}
	}
	if node.Max != nil && t != "int" && t != "float" {
		return &InvalidAttributeCombinationError{
			Path: dottedPath,
			Msg:  fmt.Sprintf("max is not valid for type %s", t),
		}
	}

	if node.Pattern != nil && t != "string" {
		return &InvalidAttributeCombinationError{
			Path: dottedPath,
			Msg:  fmt.Sprintf("pattern is not valid for type %s", t),
		}
	}

	return nil
}

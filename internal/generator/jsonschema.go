package generator

import (
	"encoding/json"
	"sort"

	"github.com/SSOHEB/configforge/internal/schema"
)

// GenerateJSONSchema consumes the AST and produces pretty-printed, deterministic JSON Schema bytes.
func GenerateJSONSchema(ast *schema.AST) ([]byte, error) {
	if ast == nil || ast.Root == nil {
		schemaObj := map[string]any{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type":    "object",
		}
		return json.MarshalIndent(schemaObj, "", "  ")
	}

	schemaObj := nodeToSchema(ast.Root)
	schemaObj["$schema"] = "http://json-schema.org/draft-07/schema#"

	return json.MarshalIndent(schemaObj, "", "  ")
}

func nodeToSchema(node *schema.Node) map[string]any {
	schemaObj := make(map[string]any)
	schemaObj["type"] = "object"

	properties := make(map[string]any)
	var required []string

	// For child fields
	for _, f := range node.Fields {
		fieldSchema := fieldToSchema(f)
		properties[f.YAMLKey] = fieldSchema
		if f.Required {
			required = append(required, f.YAMLKey)
		}
	}

	// For child namespaces (Nodes)
	for _, child := range node.Children {
		childSchema := nodeToSchema(child)
		properties[child.YAMLKey] = childSchema
	}

	if len(properties) > 0 {
		schemaObj["properties"] = properties
	}

	if len(required) > 0 {
		sort.Strings(required)
		schemaObj["required"] = required
	}

	return schemaObj
}

func fieldToSchema(f *schema.Field) map[string]any {
	schemaObj := make(map[string]any)

	if f.Description != "" {
		schemaObj["description"] = f.Description
	}

	switch f.Type {
	case schema.TypeBool:
		schemaObj["type"] = "boolean"

	case schema.TypeString:
		schemaObj["type"] = "string"
		if len(f.Enum) > 0 {
			schemaObj["enum"] = f.Enum
		}
		if f.Pattern != "" {
			schemaObj["pattern"] = f.Pattern
		}

	case schema.TypeInt:
		schemaObj["type"] = "integer"
		if f.Min != nil {
			schemaObj["minimum"] = *f.Min
		}
		if f.Max != nil {
			schemaObj["maximum"] = *f.Max
		}

	case schema.TypeFloat:
		schemaObj["type"] = "number"
		if f.Min != nil {
			schemaObj["minimum"] = *f.Min
		}
		if f.Max != nil {
			schemaObj["maximum"] = *f.Max
		}

	case schema.TypeStringSlice:
		schemaObj["type"] = "array"
		schemaObj["items"] = map[string]any{
			"type": "string",
		}

	case schema.TypeIntSlice:
		schemaObj["type"] = "array"
		schemaObj["items"] = map[string]any{
			"type": "integer",
		}

	case schema.TypeStringMap:
		schemaObj["type"] = "object"
		schemaObj["additionalProperties"] = map[string]any{
			"type": "string",
		}
	}

	if f.Default != nil {
		schemaObj["default"] = f.Default
	}

	return schemaObj
}

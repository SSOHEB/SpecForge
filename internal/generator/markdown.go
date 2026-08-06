package generator

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/SSOHEB/codrao/internal/schema"
)

// GenerateMarkdownDocs converts the AST into a structured Markdown configuration reference.
// The document uses H2 for root children, H3/H4 for deeper nesting, and lists
// fields in a Markdown table with an explicit 'Constraints' column for clean formatting.
func GenerateMarkdownDocs(ast *schema.AST) ([]byte, error) {
	if ast == nil || ast.Root == nil {
		return []byte("# Configuration Reference\n"), nil
	}

	var buf bytes.Buffer
	buf.WriteString("# Configuration Reference\n\n")

	// Render the root fields first, if any
	if len(ast.Root.Fields) > 0 {
		buf.WriteString("## Root Configuration\n\n")
		renderFieldsTable(&buf, ast.Root.Fields)
		buf.WriteString("\n")
	}

	// Helper to recursively render nodes
	var renderNode func(node *schema.Node, headingLevel int)
	renderNode = func(node *schema.Node, headingLevel int) {
		hashes := strings.Repeat("#", headingLevel)
		buf.WriteString(fmt.Sprintf("%s %s\n\n", hashes, node.Name))

		pathStr := strings.Join(node.Path, ".")
		buf.WriteString(fmt.Sprintf("**Path:** `%s`\n\n", pathStr))

		if len(node.Fields) > 0 {
			renderFieldsTable(&buf, node.Fields)
			buf.WriteString("\n")
		}

		// Sort child nodes alphabetically by Name
		children := make([]*schema.Node, len(node.Children))
		copy(children, node.Children)
		sort.Slice(children, func(i, j int) bool {
			return children[i].Name < children[j].Name
		})

		for _, child := range children {
			renderNode(child, headingLevel+1)
		}
	}

	// Sort root child nodes alphabetically by Name
	rootChildren := make([]*schema.Node, len(ast.Root.Children))
	copy(rootChildren, ast.Root.Children)
	sort.Slice(rootChildren, func(i, j int) bool {
		return rootChildren[i].Name < rootChildren[j].Name
	})

	for _, child := range rootChildren {
		renderNode(child, 2)
	}

	return buf.Bytes(), nil
}

func renderFieldsTable(buf *bytes.Buffer, fields []*schema.Field) {
	// Sort fields alphabetically by YAMLKey for deterministic output
	sortedFields := make([]*schema.Field, len(fields))
	copy(sortedFields, fields)
	sort.Slice(sortedFields, func(i, j int) bool {
		return sortedFields[i].YAMLKey < sortedFields[j].YAMLKey
	})

	buf.WriteString("| Name | Type | Default | Required | Constraints | Description |\n")
	buf.WriteString("| --- | --- | --- | --- | --- | --- |\n")

	for _, f := range sortedFields {
		typeStr := fieldTypeString(f.Type)

		defaultStr := "-"
		if f.Default != nil {
			defaultStr = fmt.Sprintf("`%v`", f.Default)
		}

		requiredStr := "No"
		if f.Required {
			requiredStr = "Yes"
		}

		var constraints []string
		if f.Min != nil {
			constraints = append(constraints, fmt.Sprintf("Min: %v", *f.Min))
		}
		if f.Max != nil {
			constraints = append(constraints, fmt.Sprintf("Max: %v", *f.Max))
		}
		if len(f.Enum) > 0 {
			constraints = append(constraints, fmt.Sprintf("Enum: [%s]", strings.Join(f.Enum, ", ")))
		}
		if f.Pattern != "" {
			constraints = append(constraints, fmt.Sprintf("Pattern: `%s`", f.Pattern))
		}

		constraintsStr := "-"
		if len(constraints) > 0 {
			constraintsStr = strings.Join(constraints, ", ")
		}

		desc := f.Description
		if desc == "" {
			desc = "-"
		}

		buf.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s | %s | %s |\n",
			f.YAMLKey, typeStr, defaultStr, requiredStr, constraintsStr, desc))
	}
}

func fieldTypeString(ft schema.FieldType) string {
	switch ft {
	case schema.TypeBool:
		return "boolean"
	case schema.TypeString:
		return "string"
	case schema.TypeInt:
		return "integer"
	case schema.TypeFloat:
		return "number"
	case schema.TypeStringSlice:
		return "array of strings"
	case schema.TypeIntSlice:
		return "array of integers"
	case schema.TypeStringMap:
		return "object (map)"
	default:
		return "unknown"
	}
}

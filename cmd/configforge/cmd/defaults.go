package cmd

import (
	"fmt"
	"os"
	"strings"

	"configforge/internal/parser"
	"configforge/internal/schema"
	"github.com/spf13/cobra"
)

var defaultsCmd = &cobra.Command{
	Use:   "defaults",
	Short: "Print a list of all configuration fields, types, and defaults",
	Long: `Defaults parses the metadata spec, builds the AST, and prints
a discoverable, human-readable list of every field's dotted path,
type, and default value or whether it is required.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			return newUserError("metadata spec file not found at %s — use --metadata to specify a valid path", metadataPath)
		}

		rawMeta, err := parser.ParseFile(metadataPath)
		if err != nil {
			return newUserError("failed to parse metadata spec at %s: %v", metadataPath, err)
		}

		ast, err := schema.Build(rawMeta)
		if err != nil {
			return newUserError("invalid metadata spec layout: %v", err)
		}

		cmd.Printf("%-40s %-20s %s\n", "FIELD PATH", "TYPE", "DEFAULT / REQUIREMENT")
		cmd.Println(strings.Repeat("-", 90))

		var printNode func(node *schema.Node)
		printNode = func(node *schema.Node) {
			for _, f := range node.Fields {
				dottedPath := strings.Join(f.Path, ".")
				typeStr := fieldTypeString(f.Type)

				var defaultStr string
				if f.Required {
					defaultStr = "REQUIRED"
				} else if f.Default == nil {
					defaultStr = "no default"
				} else {
					defaultStr = fmt.Sprintf("default: %v", f.Default)
				}

				cmd.Printf("%-40s %-20s %s\n", dottedPath, typeStr, defaultStr)
			}
			for _, child := range node.Children {
				printNode(child)
			}
		}

		printNode(ast.Root)
		return nil
	},
}

func fieldTypeString(ft schema.FieldType) string {
	switch ft {
	case schema.TypeBool:
		return "bool"
	case schema.TypeString:
		return "string"
	case schema.TypeInt:
		return "int"
	case schema.TypeFloat:
		return "float"
	case schema.TypeStringSlice:
		return "string[]"
	case schema.TypeIntSlice:
		return "int[]"
	case schema.TypeStringMap:
		return "map[string]string"
	default:
		return "unknown"
	}
}

func init() {
	rootCmd.AddCommand(defaultsCmd)
}

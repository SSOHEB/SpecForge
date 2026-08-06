package cmd

import (
	"os"
	"path/filepath"

	"github.com/SSOHEB/codrao/internal/generator"
	"github.com/SSOHEB/codrao/internal/parser"
	"github.com/SSOHEB/codrao/internal/schema"
	"github.com/spf13/cobra"
)

var (
	writeSchema bool
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Generate JSON Schema document from metadata spec",
	Long: `Schema parses the metadata specification, builds the AST,
and prints the draft-07 JSON Schema document directly to standard output,
or writes it to the output directory if --write is specified.`,
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

		schemaBytes, err := generator.GenerateJSONSchema(ast)
		if err != nil {
			return newUserError("failed to generate JSON Schema: %v", err)
		}

		if writeSchema {
			if err := os.MkdirAll(outputPath, 0755); err != nil {
				return newUserError("failed to create output directory %s: %v", outputPath, err)
			}
			schemaOut := filepath.Join(outputPath, "schema.json")
			if err := os.WriteFile(schemaOut, schemaBytes, 0644); err != nil {
				return newUserError("failed to write JSON Schema file %s: %v", schemaOut, err)
			}
			cmd.Printf("wrote JSON Schema: %s\n", schemaOut)
		} else {
			cmd.Println(string(schemaBytes))
		}

		return nil
	},
}

func init() {
	schemaCmd.Flags().BoolVarP(&writeSchema, "write", "w", false, "Write the JSON Schema to the output directory instead of printing to stdout")
	rootCmd.AddCommand(schemaCmd)
}

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SSOHEB/configforge/internal/generator"
	"github.com/SSOHEB/configforge/internal/parser"
	"github.com/SSOHEB/configforge/internal/schema"
	"github.com/spf13/cobra"
)

var (
	functionalAPI bool
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Go structs and JSON Schema from metadata spec",
	Long: `Generate parses the metadata specification, builds the AST,
and writes both draft-07 JSON Schema (schema.json) and typed Go configuration code
(generated_config.go) to the specified output directory.`,
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

		if err := os.MkdirAll(outputPath, 0755); err != nil {
			return newUserError("failed to create output directory %s: %v", outputPath, err)
		}

		schemaBytes, err := generator.GenerateJSONSchema(ast)
		if err != nil {
			return newUserError("failed to generate JSON Schema: %v", err)
		}
		schemaOut := filepath.Join(outputPath, "schema.json")
		if err := os.WriteFile(schemaOut, schemaBytes, 0644); err != nil {
			return newUserError("failed to write JSON Schema file %s: %v", schemaOut, err)
		}
		fmt.Printf("wrote JSON Schema: %s\n", schemaOut)

		opts := generator.GenOptions{WithFunctionalAPI: functionalAPI}
		goBytes, err := generator.GenerateGoCode(ast, "main", opts)
		if err != nil {
			return newUserError("failed to generate Go code: %v", err)
		}
		goOut := filepath.Join(outputPath, "generated_config.go")
		if err := os.WriteFile(goOut, goBytes, 0644); err != nil {
			return newUserError("failed to write Go config file %s: %v", goOut, err)
		}
		fmt.Printf("wrote Go config structs: %s\n", goOut)

		return nil
	},
}

func init() {
	generateCmd.Flags().BoolVar(&functionalAPI, "functional-api", false, "Generate alternative method-style configuration getters")
	rootCmd.AddCommand(generateCmd)
}

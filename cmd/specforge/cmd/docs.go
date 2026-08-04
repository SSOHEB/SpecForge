package cmd

import (
	"os"
	"path/filepath"

	"github.com/SSOHEB/configforge/internal/generator"
	"github.com/SSOHEB/configforge/internal/parser"
	"github.com/SSOHEB/configforge/internal/schema"
	"github.com/spf13/cobra"
)

var (
	stdoutDocs bool
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate Markdown documentation from metadata spec",
	Long: `Docs parses the metadata specification, builds the AST,
and writes the reference Markdown documentation to --out/CONFIGURATION.md,
or prints it directly to stdout if --stdout is specified.`,
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

		mdBytes, err := generator.GenerateMarkdownDocs(ast)
		if err != nil {
			return newUserError("failed to generate Markdown documentation: %v", err)
		}

		if stdoutDocs {
			cmd.Println(string(mdBytes))
		} else {
			if err := os.MkdirAll(outputPath, 0755); err != nil {
				return newUserError("failed to create output directory %s: %v", outputPath, err)
			}
			outFilePath := filepath.Join(outputPath, "CONFIGURATION.md")
			if err := os.WriteFile(outFilePath, mdBytes, 0644); err != nil {
				return newUserError("failed to write documentation file %s: %v", outFilePath, err)
			}
			cmd.Printf("wrote Markdown documentation: %s\n", outFilePath)
		}

		return nil
	},
}

func init() {
	docsCmd.Flags().BoolVar(&stdoutDocs, "stdout", false, "Print Markdown documentation to stdout instead of writing to --out/CONFIGURATION.md")
	rootCmd.AddCommand(docsCmd)
}

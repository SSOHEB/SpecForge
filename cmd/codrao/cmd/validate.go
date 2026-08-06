package cmd

import (
	"os"

	"github.com/SSOHEB/codrao/internal/parser"
	"github.com/SSOHEB/codrao/internal/runtime"
	"github.com/SSOHEB/codrao/internal/schema"
	"github.com/SSOHEB/codrao/internal/validator"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate application configuration against metadata schema",
	Long: `Validate parses the metadata specification, builds the semantic AST,
loads the configuration file applying defaults and environment overrides,
and runs the validator checking all rules.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			return newUserError("metadata spec file not found at %s — use --metadata to specify a valid path", metadataPath)
		}
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return newUserError("configuration file not found at %s — use --config to specify a valid path", configPath)
		}

		rawMeta, err := parser.ParseFile(metadataPath)
		if err != nil {
			return newUserError("failed to parse metadata spec at %s: %v", metadataPath, err)
		}

		ast, err := schema.Build(rawMeta)
		if err != nil {
			return newUserError("invalid metadata spec layout: %v", err)
		}

		_, rawConfig, err := runtime.LoadAndPrepareFile[map[string]any](ast, configPath, runtime.RuntimeOptions{})
		if err != nil {
			return newUserError("failed to load configuration at %s: %v", configPath, err)
		}

		valErrs := validator.Validate(ast, rawConfig)
		if len(valErrs) > 0 {
			for _, valErr := range valErrs {
				cmd.PrintErrln(valErr.Error())
			}
			return newUserError("validation failed: %d error(s) found", len(valErrs))
		}

		cmd.Println("config valid")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

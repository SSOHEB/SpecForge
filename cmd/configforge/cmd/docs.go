package cmd

import (
	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate documentation from metadata spec (not yet implemented)",
	Long:  `Docs parses the metadata spec and generates configuration documentation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return newUserError("docs command is not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
}

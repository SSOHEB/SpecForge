package cmd

import (
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version details of codrao",
	Long:  `Version prints the version tag, git commit hash, and build date of the codrao binary.`,
	Run: func(cmd *cobra.Command, _ []string) {
		cmd.Printf("codrao version: %s\ncommit: %s\nbuilt at: %s\n", version, commit, date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

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
	Short: "Print the version details of specforge",
	Long:  `Version prints the version tag, git commit hash, and build date of the specforge binary.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("specforge version: %s\ncommit: %s\nbuilt at: %s\n", version, commit, date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

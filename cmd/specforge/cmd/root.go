package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	metadataPath string
	configPath   string
	outputPath   string
)

// CliUserError is a structured error representing user input/config validation failures.
type CliUserError struct {
	Msg string
}

func (e *CliUserError) Error() string {
	return e.Msg
}

func newUserError(format string, args ...any) error {
	return &CliUserError{Msg: fmt.Sprintf(format, args...)}
}

var rootCmd = &cobra.Command{
	Use:   "specforge",
	Short: "specforge is a semantic configuration management framework",
	Long: `specforge is a command-line tool that parses configuration metadata specifications,
generates typed Go structures and JSON schemas, applies defaults and environment variable overrides,
and validates configuration files semantic rules at build time and runtime.`,
}

// Execute parses commands and manages consistent exit codes:
// 0: Success
// 1: User or config specification error (e.g. invalid config, missing file)
// 2: Unexpected internal or CLI usage error
func Execute() {
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	if err := rootCmd.Execute(); err != nil {
		var uErr *CliUserError
		if errors.As(err, &uErr) {
			fmt.Fprintf(os.Stderr, "Error: %s\n", uErr.Msg)
			os.Exit(1)
		}
		// CLI usage error (e.g. unknown command or flag)
		fmt.Fprintf(os.Stderr, "Usage Error: %v\n", err)
		os.Exit(2)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&metadataPath, "metadata", "m", "./metadata.yaml", "Path to metadata specification YAML file")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "./config.yaml", "Path to application configuration YAML file")
	rootCmd.PersistentFlags().StringVarP(&outputPath, "out", "o", "./generated", "Output directory for generated files")
}

// ExecuteWithArgs runs the CLI command in-process. This is extremely useful
// for executing black-box integration tests on restricted OS platforms.
func ExecuteWithArgs(args []string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)

	// Reset persistent flags to avoid state leakage
	metadataPath = ""
	configPath = ""
	outputPath = ""

	// Reset subcommand flags
	writeSchema = false
	functionalAPI = false
	stdoutDocs = false

	err := rootCmd.Execute()
	return buf.String(), err
}

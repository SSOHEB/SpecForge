package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	initDir   string
	initForce bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new metadata and config file",
	Long: `Init generates a starter metadata.yaml and config.yaml in the specified directory.
Use this to bootstrap a new configuration schema for your application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		metaPath := filepath.Join(initDir, "metadata.yaml")
		configPath := filepath.Join(initDir, "config.yaml")

		if !initForce {
			var exists []string
			if _, err := os.Stat(metaPath); err == nil {
				exists = append(exists, "metadata.yaml")
			}
			if _, err := os.Stat(configPath); err == nil {
				exists = append(exists, "config.yaml")
			}
			if len(exists) > 0 {
				return newUserError("initialization failed: the following files already exist in %s: %v — use --force to overwrite", initDir, exists)
			}
		}

		if err := os.MkdirAll(initDir, 0755); err != nil {
			return newUserError("failed to create directory %s: %v", initDir, err)
		}

		metaYAML := `server:
  port:
    type: int
    default: 8080
    min: 1024
    max: 65535
    description: "Port for the HTTP server to listen on"
  host:
    type: string
    default: "127.0.0.1"
    description: "Host address to bind to"
`
		if err := os.WriteFile(metaPath, []byte(metaYAML), 0644); err != nil {
			return newUserError("failed to write metadata.yaml: %v", err)
		}

		configYAML := `server:
  port: 8080
  host: "127.0.0.1"
`
		if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
			return newUserError("failed to write config.yaml: %v", err)
		}

		cmd.Printf("Successfully generated metadata.yaml and config.yaml in %s\n\n", initDir)
		cmd.Println("Next steps:")
		cmd.Printf("  codrao validate -m %s -c %s\n", metaPath, configPath)
		cmd.Printf("  codrao generate -m %s -o %s\n", metaPath, initDir)

		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initDir, "dir", ".", "Directory to initialize the files in")
	initCmd.Flags().BoolVar(&initForce, "force", false, "Overwrite existing files")
	rootCmd.AddCommand(initCmd)
}

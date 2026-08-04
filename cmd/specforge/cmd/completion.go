package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `To load completions:

Bash:
$ source <(specforge completion bash)

# To load completions for each session, execute once:
# Linux:
$ specforge completion bash > /etc/bash_completion.d/specforge
# macOS:
$ specforge completion bash > /usr/local/etc/bash_completion.d/specforge

Zsh:
# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:
$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ specforge completion zsh > "${fpath[1]}/_specforge"

Fish:
$ specforge completion fish | source
# To load completions for each session, execute once:
$ specforge completion fish > ~/.config/fish/completions/specforge.fish

PowerShell:
PS> specforge completion powershell | Out-String | Invoke-Expression
# To load completions for every new session, run:
PS> specforge completion powershell > specforge.ps1
# and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			return newUserError("unsupported shell: %s", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

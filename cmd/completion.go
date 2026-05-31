package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for cloudsql.

To load completions:

Bash:
  $ source <(cloudsql completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ cloudsql completion bash > /etc/bash_completion.d/cloudsql
  # macOS:
  $ cloudsql completion bash > $(brew --prefix)/etc/bash_completion.d/cloudsql

Zsh:
  $ source <(cloudsql completion zsh)
  # To load completions for each session, execute once:
  $ cloudsql completion zsh > "${fpath[1]}/_cloudsql"
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ cloudsql completion fish | source
  # To load completions for each session, execute once:
  $ cloudsql completion fish > ~/.config/fish/completions/cloudsql.fish

PowerShell:
  PS> cloudsql completion powershell | Out-String | Invoke-Expression
  # To load completions for each session, execute once:
  PS> cloudsql completion powershell > cloudsql.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

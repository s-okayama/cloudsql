package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for cloudsql.

Use --install to automatically install completions for your current shell.

  $ cloudsql completion --install

Or manually generate and source:

  Zsh:  source <(cloudsql completion zsh)
  Bash: source <(cloudsql completion bash)
  Fish: cloudsql completion fish | source
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		install, _ := cmd.Flags().GetBool("install")
		if install {
			installCompletion()
			return
		}
		if len(args) == 0 {
			cmd.Help()
			return
		}
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

func init() {
	completionCmd.Flags().Bool("install", false, "install completion for current shell")
}

func installCompletion() {
	shell := detectShell()

	switch shell {
	case "zsh":
		installZshCompletion()
	case "bash":
		installBashCompletion()
	case "fish":
		installFishCompletion()
	default:
		fmt.Printf("Unsupported shell: %s\n", shell)
		fmt.Println("Manually run: source <(cloudsql completion zsh)")
	}
}

func detectShell() string {
	shellEnv := os.Getenv("SHELL")
	if strings.Contains(shellEnv, "zsh") {
		return "zsh"
	}
	if strings.Contains(shellEnv, "bash") {
		return "bash"
	}
	if strings.Contains(shellEnv, "fish") {
		return "fish"
	}
	return filepath.Base(shellEnv)
}

func installZshCompletion() {
	var buf bytes.Buffer
	rootCmd.GenZshCompletion(&buf)

	// Try fpath first
	out, err := exec.Command("zsh", "-c", "echo ${fpath[1]}").Output()
	if err == nil {
		dir := strings.TrimSpace(string(out))
		if dir != "" {
			path := filepath.Join(dir, "_cloudsql")
			if err := os.WriteFile(path, buf.Bytes(), 0644); err == nil {
				fmt.Printf("Completion installed to %s\n", path)
				fmt.Println("Run 'exec zsh' to enable")
				return
			}
		}
	}

	// Fallback: ~/.zsh/completions
	home := os.Getenv("HOME")
	dir := filepath.Join(home, ".zsh", "completions")
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, "_cloudsql")
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to write completion: %v", err)
	}

	// Add to fpath in .zshrc if not already there
	zshrc := filepath.Join(home, ".zshrc")
	content, _ := os.ReadFile(zshrc)
	fpathLine := fmt.Sprintf("fpath=(%s $fpath)", dir)
	if !strings.Contains(string(content), fpathLine) {
		f, err := os.OpenFile(zshrc, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalf("Failed to update .zshrc: %v", err)
		}
		defer f.Close()
		f.WriteString("\n" + fpathLine + "\nautoload -Uz compinit && compinit\n")
	}

	fmt.Printf("Completion installed to %s\n", path)
	fmt.Println("Run 'exec zsh' to enable")
}

func installBashCompletion() {
	var buf bytes.Buffer
	rootCmd.GenBashCompletion(&buf)

	home := os.Getenv("HOME")
	dir := filepath.Join(home, ".local", "share", "bash-completion", "completions")
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, "cloudsql")
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to write completion: %v", err)
	}
	fmt.Printf("Completion installed to %s\n", path)
	fmt.Println("Restart your shell to enable")
}

func installFishCompletion() {
	var buf bytes.Buffer
	rootCmd.GenFishCompletion(&buf, true)

	home := os.Getenv("HOME")
	dir := filepath.Join(home, ".config", "fish", "completions")
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, "cloudsql.fish")
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to write completion: %v", err)
	}
	fmt.Printf("Completion installed to %s\n", path)
	fmt.Println("Restart your shell to enable")
}

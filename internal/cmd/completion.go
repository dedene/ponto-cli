package cmd

import (
	"fmt"
)

// CompletionCmd generates shell completions.
type CompletionCmd struct {
	Shell string `arg:"" help:"Shell type (bash, zsh, fish, powershell)" enum:"bash,zsh,fish,powershell"`
}

func (c *CompletionCmd) Run() error {
	// Kong has built-in completion support
	// For now, print a message directing users to Kong's completion
	fmt.Printf("Shell completions for %s:\n\n", c.Shell)

	switch c.Shell {
	case "bash":
		fmt.Println("Add to ~/.bashrc:")
		fmt.Println(`  eval "$(ponto --help-completions=bash)"`)
	case "zsh":
		fmt.Println("Add to ~/.zshrc:")
		fmt.Println(`  eval "$(ponto --help-completions=zsh)"`)
	case "fish":
		fmt.Println("Add to ~/.config/fish/config.fish:")
		fmt.Println(`  ponto --help-completions=fish | source`)
	case "powershell":
		fmt.Println("Add to your PowerShell profile:")
		fmt.Println(`  ponto --help-completions=powershell | Invoke-Expression`)
	}

	return nil
}

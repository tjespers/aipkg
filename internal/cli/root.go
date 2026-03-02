package cli

import (
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// SetVersionInfo sets the version information for the CLI.
// Called from main.go with ldflags-injected values.
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "aipkg",
		Short:         "Package manager for AI artifacts",
		Long:          "aipkg is a package manager for AI artifacts: skills, prompts, commands, agents, and MCP server configs.",
		Version:       version + " (" + commit + ", " + date + ")",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newCreateCmd())

	return cmd
}

// Execute runs the root command.
func Execute() error {
	return newRootCmd().Execute()
}

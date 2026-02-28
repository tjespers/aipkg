package cli

import (
	"github.com/spf13/cobra"
)

// Version info set via ldflags at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// SetVersionInfo injects build-time version metadata.
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}

// NewRootCmd creates the top-level aipkg command.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "aipkg",
		Short:         "AI package manager",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       version + " (" + commit + ", " + date + ")",
	}
	cmd.AddCommand(newInitCmd())
	return cmd
}

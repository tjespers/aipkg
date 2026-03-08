package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tjespers/aipkg/internal/project"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new aipkg project",
		Long:  "Initialize a new aipkg project by creating an aipkg-project.json file in the current directory.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInit(cmd)
		},
	}
	return cmd
}

func runInit(cmd *cobra.Command) error {
	// Check for existing aipkg.json (FR-017, mutual exclusivity).
	if _, err := os.Stat("aipkg.json"); err == nil {
		return fmt.Errorf("package manifest (aipkg.json) already exists in this directory; use aipkg require or aipkg install instead")
	}

	// Check for existing aipkg-project.json (FR-016).
	if _, err := os.Stat("aipkg-project.json"); err == nil {
		return fmt.Errorf("project already initialized (aipkg-project.json exists)")
	}

	if err := project.Create("."); err != nil {
		return fmt.Errorf("cannot initialize project: %w", err)
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Initialized project in aipkg-project.json")
	return nil
}

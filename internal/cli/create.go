package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/tjespers/aipkg/internal/license"
	"github.com/tjespers/aipkg/internal/manifest"
	"github.com/tjespers/aipkg/internal/naming"
	"github.com/tjespers/aipkg/internal/scaffold"
	"github.com/tjespers/aipkg/internal/schema"
)

func newCreateCmd() *cobra.Command {
	var (
		flagName        string
		flagVersion     string
		flagDescription string
		flagLicense     string
		flagPath        string
	)

	cmd := &cobra.Command{
		Use:   "create [@scope/package-name]",
		Short: "Create a new aipkg package",
		Long:  "Create a new aipkg package directory with a valid aipkg.json manifest and well-known artifact directories.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve name: positional arg takes precedence over --name flag.
			pkgName := flagName
			if len(args) > 0 {
				pkgName = args[0]
			}

			return runCreate(cmd, &createOptions{
				name:        pkgName,
				version:     flagVersion,
				description: flagDescription,
				license:     flagLicense,
				path:        flagPath,
			})
		},
	}

	cmd.Flags().StringVarP(&flagName, "name", "n", "", "package name (alternative to positional arg)")
	cmd.Flags().StringVarP(&flagVersion, "version", "v", "", "package version (strict semver)")
	cmd.Flags().StringVarP(&flagDescription, "description", "d", "", "short package description")
	cmd.Flags().StringVarP(&flagLicense, "license", "l", "", "SPDX license identifier or \"proprietary\"")
	cmd.Flags().StringVarP(&flagPath, "path", "p", "", "target directory (default: derived from package name)")

	return cmd
}

type createOptions struct {
	name        string
	version     string
	description string
	license     string
	path        string
}

func runCreate(cmd *cobra.Command, opts *createOptions) error {
	isTTY := term.IsTerminal(int(os.Stdin.Fd())) //nolint:gosec // fd 0 is always in int range

	// Determine which fields still need values.
	needName := opts.name == ""
	needVersion := !cmd.Flags().Changed("version")
	needDescription := !cmd.Flags().Changed("description")
	needLicense := !cmd.Flags().Changed("license")

	// Non-interactive mode: if no TTY and fields are missing, error out.
	if !isTTY && (needName || needVersion) {
		var missing []string
		if needName {
			missing = append(missing, "--name")
		}
		if needVersion {
			missing = append(missing, "--version")
		}
		return fmt.Errorf("missing required flags for non-interactive mode: %s", strings.Join(missing, ", "))
	}

	// Set version default for prompting.
	if needVersion {
		opts.version = "0.1.0"
	}

	// Resolve target directory early if possible (for license detection).
	targetDir, err := resolveTargetDir(opts.path, opts.name)
	if err != nil {
		return err
	}

	// Check for existing aipkg.json if targetDir is known.
	if targetDir != "" {
		if err := checkExistingManifest(targetDir); err != nil {
			return err
		}
	}

	// Detect license from existing LICENSE file.
	// Check target directory first, then fall back to cwd (covers the common
	// case of creating a new subdirectory inside a repo that has a LICENSE).
	detectedLicense := ""
	if needLicense {
		if targetDir != "" {
			if id, ok := license.Detect(targetDir); ok {
				detectedLicense = id
			}
		}
		if detectedLicense == "" {
			if id, ok := license.Detect("."); ok {
				detectedLicense = id
			}
		}
		opts.license = detectedLicense
	}

	// Build and run interactive prompts for missing fields.
	if isTTY && (needName || needVersion || needDescription || needLicense) {
		if err := runPrompts(opts, needName, needVersion, needDescription, needLicense, detectedLicense); err != nil {
			if err.Error() == "user aborted" {
				return nil
			}
			return err
		}
	}

	// Validate the name.
	parsed, err := naming.Parse(opts.name)
	if err != nil {
		return err
	}
	if rule, reserved := naming.IsReservedScope(parsed.Scope); reserved {
		return fmt.Errorf("scope %q is reserved (%s)", parsed.Scope, rule)
	}

	// Validate version via schema bridge.
	if err := schema.ValidateField("version")(opts.version); err != nil {
		return fmt.Errorf("version must be in MAJOR.MINOR.PATCH format (e.g., 1.0.0)")
	}

	// Resolve target directory from name if not yet set.
	if targetDir == "" {
		targetDir = parsed.Package
	}

	// Check for existing aipkg.json again if directory was derived from name.
	if err := checkExistingManifest(targetDir); err != nil {
		return err
	}

	// Build manifest.
	m := &manifest.PackageManifest{
		SpecVersion: 1,
		Name:        opts.name,
		Version:     opts.version,
		Description: opts.description,
		License:     opts.license,
	}

	// Create directory structure.
	if err := scaffold.Create(targetDir); err != nil {
		return fmt.Errorf("cannot create package: %w", err)
	}

	// Write aipkg.json last (atomic: either all files created or manifest not written).
	if err := m.WriteFile(targetDir); err != nil {
		return fmt.Errorf("cannot create package: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created package %s in %s/\n", opts.name, targetDir)
	return nil
}

func resolveTargetDir(path, name string) (string, error) {
	if path != "" {
		return path, nil
	}
	if name != "" {
		parsed, err := naming.Parse(name)
		if err != nil {
			// Name will be validated later; return empty for now.
			return "", nil //nolint:nilerr // intentional: defer validation to later
		}
		return parsed.Package, nil
	}
	return "", nil
}

func checkExistingManifest(dir string) error {
	manifestPath := filepath.Join(dir, "aipkg.json")
	if _, err := os.Stat(manifestPath); err == nil {
		return fmt.Errorf("target directory already contains aipkg.json")
	}
	return nil
}

func runPrompts(opts *createOptions, needName, needVersion, needDescription, needLicense bool, detectedLicense string) error {
	var fields []huh.Field

	if needName {
		fields = append(fields, huh.NewInput().
			Title("Package name").
			Description("Scoped name like @scope/package-name").
			Value(&opts.name).
			Validate(validateName))
	}

	if needVersion {
		fields = append(fields, huh.NewInput().
			Title("Version").
			Value(&opts.version).
			Validate(func(s string) error {
				if err := schema.ValidateField("version")(s); err != nil {
					return fmt.Errorf("version must be in MAJOR.MINOR.PATCH format (e.g., 1.0.0)")
				}
				return nil
			}))
	}

	if needDescription {
		fields = append(fields, huh.NewInput().
			Title("Description").
			Description("Optional short summary (max 255 chars)").
			Value(&opts.description).
			Validate(func(s string) error {
				if err := schema.ValidateField("description")(s); err != nil {
					return fmt.Errorf("description must be at most 255 characters")
				}
				return nil
			}))
	}

	if needLicense {
		desc := "Optional SPDX identifier or \"proprietary\""
		if detectedLicense != "" {
			desc = fmt.Sprintf("Detected: %s (press Enter to accept)", detectedLicense)
		}
		fields = append(fields, huh.NewInput().
			Title("License").
			Description(desc).
			Value(&opts.license))
	}

	if len(fields) == 0 {
		return nil
	}

	form := huh.NewForm(huh.NewGroup(fields...))
	return form.Run()
}

func validateName(s string) error {
	// First validate format via schema bridge.
	if err := schema.ValidateField("name")(s); err != nil {
		if !strings.HasPrefix(s, "@") || !strings.Contains(s, "/") {
			return fmt.Errorf("package name must be scoped (e.g., @scope/package-name)")
		}
		return fmt.Errorf("package name must match @scope/package-name format with lowercase alphanumeric and hyphens only")
	}

	// Then check reserved scopes.
	parsed, err := naming.Parse(s)
	if err != nil {
		return err
	}
	if rule, reserved := naming.IsReservedScope(parsed.Scope); reserved {
		return fmt.Errorf("scope %q is reserved (%s)", parsed.Scope, rule)
	}

	return nil
}

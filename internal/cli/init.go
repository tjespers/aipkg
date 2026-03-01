package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/tjespers/aipkg/internal/license"
	"github.com/tjespers/aipkg/internal/manifest"
	"github.com/tjespers/aipkg/internal/schema"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a new aipkg.json manifest",
		Example: `  # Interactive — prompts for all fields
  aipkg init

  # Non-interactive package
  aipkg init --type package --name @myorg/my-skill --version 1.0.0

  # Non-interactive project
  aipkg init --type project

  # Hybrid — provide some flags, prompted for the rest
  aipkg init --type package --name @myorg/my-skill`,
		RunE: runInit,
	}
	cmd.Flags().String("type", "", "Manifest type: project or package")
	cmd.Flags().String("name", "", "Scoped package name (@scope/name)")
	cmd.Flags().String("version", "", "Package version (MAJOR.MINOR.PATCH)")
	cmd.Flags().String("description", "", "Short description (max 255 chars)")
	cmd.Flags().String("license", "", "SPDX license identifier")
	return cmd
}

func isTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func runInit(cmd *cobra.Command, _ []string) error {
	if _, err := os.Stat("aipkg.json"); err == nil {
		return fmt.Errorf("aipkg.json already exists")
	}

	interactive := isTTY()

	flags := cmd.Flags()
	typVal, _ := flags.GetString("type")
	nameVal, _ := flags.GetString("name")
	versionVal, _ := flags.GetString("version")
	descVal, _ := flags.GetString("description")
	licenseVal, _ := flags.GetString("license")

	if typVal == "project" {
		if cmd.Flags().Changed("version") {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Warning: --version is ignored for project type")
		}
		if cmd.Flags().Changed("license") {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Warning: --license is ignored for project type")
		}
	}

	if typVal == "" {
		if !interactive {
			return fmt.Errorf("missing required flag: --type")
		}
		if err := promptType(&typVal); err != nil {
			return err
		}
	}

	if typVal != "project" && typVal != "package" {
		return fmt.Errorf("invalid type %q: must be \"project\" or \"package\"", typVal)
	}

	if typVal == "package" {
		if err := collectPackageFields(cmd, interactive, &nameVal, &versionVal, &descVal, &licenseVal); err != nil {
			return err
		}
	} else {
		if err := collectProjectFields(cmd, interactive, &nameVal, &descVal); err != nil {
			return err
		}
	}

	m := &manifest.Manifest{
		Type:        typVal,
		Name:        nameVal,
		Description: descVal,
	}
	if typVal == "package" {
		m.Version = versionVal
		m.License = licenseVal
	}

	data, err := m.Marshal()
	if err != nil {
		return fmt.Errorf("failed to serialize manifest: %w", err)
	}

	if err := schema.ValidateManifest(data); err != nil {
		return err
	}

	if err := os.WriteFile("aipkg.json", data, 0o600); err != nil { //nolint:gosec // manifest is not sensitive
		return fmt.Errorf("failed to write aipkg.json: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created aipkg.json (%s)\n", typVal)
	return nil
}

func promptType(typVal *string) error {
	return runForm([]huh.Field{
		huh.NewSelect[string]().
			Title("What type of manifest?").
			Options(
				huh.NewOption("Package — publishable bundle of AI artifacts", "package"),
				huh.NewOption("Project — local project configuration", "project"),
			).
			Value(typVal),
	})
}

func collectPackageFields(cmd *cobra.Command, interactive bool, name, ver, desc, lic *string) error {
	if *name != "" {
		if err := schema.ValidateName(*name); err != nil {
			return err
		}
	}
	if *ver != "" {
		if err := schema.ValidateVersion(*ver); err != nil {
			return err
		}
	}
	if cmd.Flags().Changed("description") && *desc != "" {
		if err := schema.ValidateDescription(*desc); err != nil {
			return err
		}
	}

	if !interactive {
		if *name == "" {
			return fmt.Errorf("missing required flags for package: --name")
		}
		if *ver == "" {
			*ver = "0.1.0"
		}
		if !cmd.Flags().Changed("license") {
			*lic = license.Detect(".")
		}
		return nil
	}

	var fields []huh.Field
	var nameBare string
	promptName := *name == ""

	if promptName {
		fields = append(fields, huh.NewInput().
			Title("Package name").
			Prompt("@ ").
			Placeholder("scope/package-name").
			Value(&nameBare).
			Validate(func(s string) error {
				return schema.ValidateName("@" + s)
			}),
		)
	}

	if *ver == "" {
		*ver = "0.1.0"
		fields = append(fields, huh.NewInput().
			Title("Version").
			Value(ver).
			Validate(schema.ValidateVersion),
		)
	}

	if !cmd.Flags().Changed("description") {
		fields = append(fields, huh.NewInput().
			Title("Description (optional)").
			Value(desc).
			Validate(optionalValidator(schema.ValidateDescription)),
		)
	}

	if !cmd.Flags().Changed("license") {
		detected := license.Detect(".")
		if detected != "" {
			*lic = detected
		}
		fields = append(fields, huh.NewInput().
			Title("License (optional)").
			Placeholder("SPDX identifier, e.g. Apache-2.0").
			Value(lic),
		)
	}

	if len(fields) == 0 {
		return nil
	}

	if err := runForm(fields); err != nil {
		return err
	}
	if promptName {
		*name = "@" + nameBare
	}
	return nil
}

func collectProjectFields(cmd *cobra.Command, interactive bool, name, desc *string) error {
	if *name != "" {
		if err := schema.ValidateName(*name); err != nil {
			return err
		}
	}
	if *desc != "" {
		if err := schema.ValidateDescription(*desc); err != nil {
			return err
		}
	}

	if !interactive {
		return nil
	}

	var fields []huh.Field
	var nameBare string
	promptName := !cmd.Flags().Changed("name")

	if promptName {
		fields = append(fields, huh.NewInput().
			Title("Project name (optional)").
			Prompt("@ ").
			Placeholder("scope/project-name").
			Value(&nameBare).
			Validate(optionalValidator(func(s string) error {
				return schema.ValidateName("@" + s)
			})),
		)
	}

	if !cmd.Flags().Changed("description") {
		fields = append(fields, huh.NewInput().
			Title("Description (optional)").
			Value(desc).
			Validate(optionalValidator(schema.ValidateDescription)),
		)
	}

	if len(fields) == 0 {
		return nil
	}

	if err := runForm(fields); err != nil {
		return err
	}
	if promptName && nameBare != "" {
		*name = "@" + nameBare
	}
	return nil
}

func runForm(fields []huh.Field) error {
	form := huh.NewForm(huh.NewGroup(fields...))
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return fmt.Errorf("cancelled")
		}
		return err
	}
	return nil
}

func optionalValidator(validate func(string) error) func(string) error {
	return func(s string) error {
		if s == "" {
			return nil
		}
		return validate(s)
	}
}

package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tjespers/aipkg/internal/archive"
	"github.com/tjespers/aipkg/internal/artifact"
	"github.com/tjespers/aipkg/internal/ignore"
	"github.com/tjespers/aipkg/internal/manifest"
	"github.com/tjespers/aipkg/internal/naming"
	"github.com/tjespers/aipkg/internal/schema"
)

func newPackCmd() *cobra.Command {
	var flagOutput string

	cmd := &cobra.Command{
		Use:   "pack [directory]",
		Short: "Create a distributable .aipkg archive",
		Long:  "Package the current directory (or the specified directory) into a distributable .aipkg archive with a SHA-256 sidecar file.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			srcDir := "."
			if len(args) > 0 {
				srcDir = args[0]
			}
			return runPack(cmd, srcDir, flagOutput)
		},
	}

	cmd.Flags().StringVarP(&flagOutput, "output", "o", "", "output path (file or directory)")

	return cmd
}

func runPack(cmd *cobra.Command, srcDir, output string) error {
	// Step 1: Load manifest.
	m, err := manifest.LoadFile(srcDir)
	if err != nil {
		return err
	}

	// Step 2: Validate manifest against schema.
	jsonData, err := m.MarshalIndent()
	if err != nil {
		return fmt.Errorf("marshaling manifest: %w", err)
	}
	if err := schema.Validate(jsonData); err != nil {
		return fmt.Errorf("manifest validation failed: %w", err)
	}

	// Parse the scoped name for filename construction.
	parsed, err := naming.Parse(m.Name)
	if err != nil {
		return err
	}

	// Resolve output path early so we can exclude the archive from itself.
	archiveName := fmt.Sprintf("%s--%s-%s.aipkg", parsed.Scope, parsed.Package, m.Version)
	archivePath, err := resolveOutputPath(output, archiveName)
	if err != nil {
		return err
	}

	// Step 3: Load ignore rules and discover artifacts.
	rules, err := ignore.LoadRules(srcDir, archivePath)
	if err != nil {
		return fmt.Errorf("loading ignore rules: %w", err)
	}
	artifacts, err := artifact.Discover(srcDir, rules.IsExcluded)
	if err != nil {
		return err
	}

	// FR-011: At least one artifact required.
	if len(artifacts) == 0 {
		return fmt.Errorf("no artifacts discovered in any well-known directory")
	}

	// Step 4: Validate each artifact.
	if err := artifact.ValidateAll(srcDir, artifacts); err != nil {
		count := countErrors(err)
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "pack: %d validation %s\n", count, pluralize("error", count))
		return err
	}

	// Step 5: Build enriched manifest (add artifacts array).
	enriched := *m
	enriched.Artifacts = artifacts
	enrichedJSON, err := enriched.MarshalIndent()
	if err != nil {
		return fmt.Errorf("marshaling enriched manifest: %w", err)
	}

	// Validate enriched manifest against schema.
	if err := schema.Validate(enrichedJSON); err != nil {
		return fmt.Errorf("enriched manifest validation failed: %w", err)
	}

	// Collect file paths for the archive.
	var paths []string
	for _, art := range artifacts {
		paths = append(paths, art.Path)
	}

	// Step 6: Create zip archive.
	f, err := os.Create(archivePath) //nolint:gosec // user-specified output path is intentional
	if err != nil {
		return fmt.Errorf("creating archive: %w", err)
	}
	defer f.Close() //nolint:errcheck // best-effort close on deferred path

	topLevelDir := parsed.Package
	if err := archive.CreateArchive(f, srcDir, topLevelDir, paths, enrichedJSON); err != nil {
		// Clean up partial archive on failure.
		_ = f.Close()
		_ = os.Remove(archivePath)
		return fmt.Errorf("creating archive: %w", err)
	}
	_ = f.Close()

	// Step 7: Write SHA-256 sidecar.
	if err := archive.WriteSidecar(archivePath); err != nil {
		return err
	}

	// Step 8: Print summary to stderr.
	info, _ := os.Stat(archivePath)
	size := formatSize(info.Size())
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s (%d %s, %s)\n",
		archivePath, len(artifacts), pluralize("artifact", len(artifacts)), size)

	return nil
}

func resolveOutputPath(output, defaultName string) (string, error) {
	if output == "" {
		return defaultName, nil
	}

	// If output is an existing directory or ends with a separator, write inside it.
	info, err := os.Stat(output)
	if err == nil && info.IsDir() {
		return filepath.Join(output, defaultName), nil
	}
	if output[len(output)-1] == filepath.Separator || output[len(output)-1] == '/' {
		return filepath.Join(output, defaultName), nil
	}

	// Otherwise treat as exact file path. Parent must exist.
	parent := filepath.Dir(output)
	if _, err := os.Stat(parent); err != nil {
		return "", fmt.Errorf("output directory does not exist: %s", parent)
	}

	return output, nil
}

func countErrors(err error) int {
	if err == nil {
		return 0
	}
	count := 1
	for _, c := range err.Error() {
		if c == '\n' {
			count++
		}
	}
	return count
}

func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	case bytes >= 1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/rhydianjenkins/apix/v3/internal/config"
)

const (
	readmeHelpStartMarker = "<!-- apix-help:start -->"
	readmeHelpEndMarker   = "<!-- apix-help:end -->"
)

// TestReadmeHelpSectionUpToDate keeps the `apix help` block in README.md in
// sync with the CLI's actual --help output, so the docs can't silently drift
// from the real command surface.
//
// If this fails after changing a command's Use/Short/Example or flags,
// regenerate the section with:
//
//	UPDATE_README=1 go test ./cmd/apix -run TestReadmeHelpSectionUpToDate
func TestReadmeHelpSectionUpToDate(t *testing.T) {
	path := readmePath(t)

	readme, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}

	section := regexp.MustCompile(
		"(?s)(" + regexp.QuoteMeta(readmeHelpStartMarker) + "\n)(.*?)(\n" + regexp.QuoteMeta(readmeHelpEndMarker) + ")",
	)

	match := section.FindSubmatchIndex(readme)
	if match == nil {
		t.Fatalf("could not find %q / %q markers in %s", readmeHelpStartMarker, readmeHelpEndMarker, path)
	}
	innerStart, innerEnd := match[4], match[5]

	got := string(readme[innerStart:innerEnd])
	want := "```\n" + generateRootHelp(t) + "\n```"

	if got == want {
		return
	}

	if os.Getenv("UPDATE_README") == "1" {
		updated := make([]byte, 0, len(readme)-len(got)+len(want))
		updated = append(updated, readme[:innerStart]...)
		updated = append(updated, want...)
		updated = append(updated, readme[innerEnd:]...)

		if err := os.WriteFile(path, updated, 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", path, err)
		}

		t.Logf("updated %s", path)
		return
	}

	t.Errorf(
		"README.md `apix help` section is out of date with `apix --help`.\n"+
			"Run `UPDATE_README=1 go test ./cmd/apix -run TestReadmeHelpSectionUpToDate` to regenerate it.\n\nwant:\n%s\n\ngot:\n%s",
		want, got,
	)
}

// generateRootHelp returns the output of `apix --help` for the full command
// tree, trimmed of trailing newlines.
//
// config.CfgPaths is pinned to a placeholder so the generated `--config`
// default doesn't bake in whatever $HOME happens to be on the machine (or CI
// runner) that regenerates it.
func generateRootHelp(t *testing.T) string {
	t.Helper()

	origCfgPaths := config.CfgPaths
	config.CfgPaths = []string{"~/.apix.yaml"}
	t.Cleanup(func() { config.CfgPaths = origCfgPaths })

	cmd := newRootCmd("dev")
	cmd.SetArgs([]string{"--help"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("failed to execute --help: %v", err)
	}

	return strings.TrimRight(out.String(), "\n")
}

func readmePath(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine caller for README path resolution")
	}

	return filepath.Join(filepath.Dir(thisFile), "..", "..", "README.md")
}

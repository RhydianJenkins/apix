package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rhydianjenkins/apix/v3/pkg/config"
)

func TestEditHandler_OpensLastConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	outFile := filepath.Join(tempDir, "edited.txt")
	editor := filepath.Join(tempDir, "editor.sh")

	script := "#!/bin/sh\nprintf '%s' \"$1\" > " + outFile + "\n"
	if err := os.WriteFile(editor, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EDITOR", editor)

	config.CfgPaths = []string{
		filepath.Join(tempDir, "tracked.yaml"),
		filepath.Join(tempDir, "private.yaml"),
	}

	EditHandler(nil, nil)

	got, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("expected editor to have run: %v", err)
	}
	if want := config.CfgPaths[1]; string(got) != want {
		t.Errorf("editor opened %q, want %q", got, want)
	}
}

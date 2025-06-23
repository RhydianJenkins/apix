package handlers

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/spf13/cobra"
)

func EditHandler(cmd *cobra.Command, args []string) {
	editor := getEditor()
	editCmd := exec.Command(editor, config.CfgPath)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr
	editCmd.Run()
}

func getEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual
	}

	switch runtime.GOOS {
	case "windows":
		return "notepad"
	case "darwin":
		return "open"
	default:
		return "nano"
	}
}

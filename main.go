package main

import (
	_ "embed"
	"os"

	"github.com/rhydianjenkins/apix/cmd"
)

//go:embed VERSION
var version string

func main() {
	if err := cmd.Execute(version); err != nil {
		os.Exit(1)
	}
}

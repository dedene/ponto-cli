package cmd

import (
	"fmt"
	"runtime"
)

// Build-time variables injected via ldflags.
var (
	version = "dev"
	commit  = ""
	date    = ""
)

// VersionCmd prints version information.
type VersionCmd struct{}

func (c *VersionCmd) Run() error {
	fmt.Printf("ponto %s\n", VersionString())

	return nil
}

// VersionString returns the version string for display.
func VersionString() string {
	v := version
	if commit != "" {
		v += " (" + commit + ")"
	}

	if date != "" {
		v += " " + date
	}

	v += " " + runtime.GOOS + "/" + runtime.GOARCH

	return v
}

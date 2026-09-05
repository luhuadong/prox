package shell

import (
	_ "embed"
	"strings"
)

//go:embed prox.bash
var bashHook string

// BashHook returns the Bash integration script for the current prox version.
func BashHook(version string) string {
	return strings.ReplaceAll(bashHook, "@VERSION@", version)
}

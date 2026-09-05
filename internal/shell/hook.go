package shell

import (
	_ "embed"
	"strings"
)

//go:embed prox.bash
var bashHook string

//go:embed completion.bash
var bashCompletion string

// BashHook returns the Bash integration script for the current prox version.
func BashHook(version string) string {
	return strings.ReplaceAll(bashHook, "@VERSION@", version) + "\n" + bashCompletion
}

// BashCompletion returns the standalone Bash completion script.
func BashCompletion() string {
	return bashCompletion
}

package main

import (
	"os"
	"runtime/debug"

	"github.com/luhuadong/prox/internal/app"
)

var version string

func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, resolvedVersion()))
}

func resolvedVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

package main

import (
	"os"

	"prox/internal/app"
)

var version = "0.1.0"

func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, version))
}

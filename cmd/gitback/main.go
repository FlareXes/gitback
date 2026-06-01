// Package main is the entry point for the gitback CLI.
// cmd/gitback/main.go

package main

import (
	"fmt"
	"os"

	"github.com/flarexes/gitback/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

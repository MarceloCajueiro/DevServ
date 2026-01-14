// DevServ - Local Development Server Manager
//
// A CLI tool for managing multiple local development services with an
// interactive TUI dashboard, structured logging, and intelligent process
// management.
package main

import (
	"fmt"
	"os"

	"github.com/marcelocajueiro/devserv/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

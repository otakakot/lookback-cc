package main

import (
	"fmt"
	"os"

	"github.com/otakakot/lookback-cc/internal/cli"
	"github.com/otakakot/lookback-cc/internal/version"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "install":
		os.Exit(cli.RunInstall())
	case "uninstall":
		os.Exit(cli.RunUninstall())
	case "version", "--version", "-v":
		fmt.Println("lookback-cc", version.Version)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: lookback-cc <command>

Commands:
  install      Install debrief, summarize, and report commands
  uninstall    Remove installed commands and hooks
  version      Show version information
`)
}

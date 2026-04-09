package main

import (
	"fmt"
	"os"

	"github.com/otakakot/lookback-cc/internal/cli"
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
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: lookback-cc <command>

Commands:
  install      Install debrief hook, summarize, and report commands
  uninstall    Remove installed hooks and commands
`)
}

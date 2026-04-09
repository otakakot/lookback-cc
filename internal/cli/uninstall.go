package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

func RunUninstall() int {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")
	debriefBinary := filepath.Join(home, ".claude", "hooks", "debrief")
	gobin := goBinDir()

	// Remove debrief hook.
	fmt.Println("==> Removing debrief hook...")

	if removed, err := removeIfExists(debriefBinary); err != nil {
		fmt.Fprintf(os.Stderr, "    Error: %v\n", err)
		return 1
	} else if removed {
		fmt.Printf("    Removed: %s\n", debriefBinary)
	} else {
		fmt.Println("    Skipped: not found")
	}

	// Remove SessionEnd hook from settings.
	fmt.Println()
	fmt.Println("==> Removing SessionEnd hook from settings...")

	if backup, err := backupSettings(settingsPath); err != nil {
		fmt.Fprintf(os.Stderr, "    Error: backup: %v\n", err)
		return 1
	} else if backup != "" {
		fmt.Printf("    Backup: %s\n", backup)
	}

	result, err := settingsUninstall(settingsPath, debriefBinary)
	if err != nil {
		fmt.Fprintf(os.Stderr, "    Error: %v\n", err)
		return 1
	}

	switch result {
	case "not_found":
		fmt.Println("    Skipped: not found")
	case "uninstalled":
		fmt.Printf("    Removed: debrief hook from %s\n", settingsPath)
	}

	// Remove summarize command.
	fmt.Println()
	fmt.Println("==> Removing summarize command...")

	summarizePath := filepath.Join(gobin, "summarize")

	if removed, err := removeIfExists(summarizePath); err != nil {
		fmt.Fprintf(os.Stderr, "    Error: %v\n", err)
		return 1
	} else if removed {
		fmt.Printf("    Removed: %s\n", summarizePath)
	} else {
		fmt.Println("    Skipped: not found")
	}

	// Remove report command.
	fmt.Println()
	fmt.Println("==> Removing report command...")

	reportPath := filepath.Join(gobin, "report")

	if removed, err := removeIfExists(reportPath); err != nil {
		fmt.Fprintf(os.Stderr, "    Error: %v\n", err)
		return 1
	} else if removed {
		fmt.Printf("    Removed: %s\n", reportPath)
	} else {
		fmt.Println("    Skipped: not found")
	}

	fmt.Println()
	fmt.Println("Done! Summaries in ~/.claude/debrief/ and reports in ~/.claude/report/ are preserved.")

	return 0
}

func removeIfExists(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	}

	return true, os.Remove(path)
}

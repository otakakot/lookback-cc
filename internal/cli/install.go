package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/otakakot/lookback-cc/internal/version"
)

func RunInstall() int {
	fmt.Println("==> Checking prerequisites...")

	for _, cmd := range []string{"go", "claude"} {
		path, err := exec.LookPath(cmd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "    Error: '%s' not found. Please install it first.\n", cmd)
			return 1
		}

		fmt.Printf("    %s: %s\n", cmd, path)
	}

	modPath, modVer := moduleInfo()
	if modPath == "" {
		fmt.Fprintln(os.Stderr, "    Error: could not determine module info")
		return 1
	}

	fmt.Printf("    module: %s@%s\n", modPath, modVer)

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "    Error: %v\n", err)
		return 1
	}

	gobin := goBinDir()
	settingsPath := filepath.Join(home, ".claude", "settings.json")
	debriefBinary := filepath.Join(gobin, "debrief")
	outputDir := filepath.Join(home, ".claude", "debrief")

	// Install debrief command.
	fmt.Println()
	fmt.Println("==> Installing debrief command...")

	if err := goInstall(modPath+"/cmd/debrief", modVer, ""); err != nil {
		fmt.Fprintf(os.Stderr, "    Error: %v\n", err)
		return 1
	}

	fmt.Printf("    Installed: %s\n", debriefBinary)

	// Create output directory.
	fmt.Println()
	fmt.Println("==> Creating output directory...")

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "    Error: mkdir: %v\n", err)
		return 1
	}

	fmt.Printf("    Created: %s\n", outputDir)

	// Install summarize command.
	fmt.Println()
	fmt.Println("==> Installing summarize command...")

	if err := goInstall(modPath+"/cmd/summarize", modVer, ""); err != nil {
		fmt.Fprintf(os.Stderr, "    Error: %v\n", err)
		return 1
	}

	fmt.Printf("    Installed: %s\n", filepath.Join(gobin, "summarize"))

	// Install report command.
	fmt.Println()
	fmt.Println("==> Installing report command...")

	if err := goInstall(modPath+"/cmd/report", modVer, ""); err != nil {
		fmt.Fprintf(os.Stderr, "    Error: %v\n", err)
		return 1
	}

	fmt.Printf("    Installed: %s\n", filepath.Join(gobin, "report"))

	// Configure SessionEnd hook.
	fmt.Println()
	fmt.Println("==> Configuring SessionEnd hook...")

	if backup, err := backupSettings(settingsPath); err != nil {
		fmt.Fprintf(os.Stderr, "    Error: backup: %v\n", err)
		return 1
	} else if backup != "" {
		fmt.Printf("    Backup: %s\n", backup)
	}

	result, err := settingsInstall(settingsPath, debriefBinary)
	if err != nil {
		fmt.Fprintf(os.Stderr, "    Error: %v\n", err)
		return 1
	}

	switch result {
	case "already_configured":
		fmt.Println("    Skipped: already configured")
	case "installed":
		fmt.Printf("    Configured: %s\n", settingsPath)
	}

	// Verify installed versions.
	fmt.Println()
	fmt.Println("==> Verifying installed versions...")

	for _, cmd := range []struct{ name, path string }{
		{"debrief", filepath.Join(gobin, "debrief")},
		{"summarize", filepath.Join(gobin, "summarize")},
		{"report", filepath.Join(gobin, "report")},
	} {
		out, err := exec.Command(cmd.path, "--version").Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "    %s: failed to get version: %v\n", cmd.name, err)
		} else {
			fmt.Printf("    %s\n", strings.TrimSpace(string(out)))
		}
	}

	fmt.Println()
	fmt.Printf("lookback-cc %s installed successfully!\n", version.Version)
	fmt.Printf("Summaries will be saved to: %s\n", outputDir)

	return 0
}

func moduleInfo() (string, string) {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "", ""
	}

	return bi.Main.Path, version.Version
}

func goInstall(pkg, version, gobin string) error {
	target := pkg + "@" + version

	cmd := exec.Command("go", "install", target)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if gobin != "" {
		cmd.Env = append(os.Environ(), "GOBIN="+gobin)
	}

	return cmd.Run()
}

func goBinDir() string {
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		return gobin
	}

	out, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return ""
	}

	return filepath.Join(strings.TrimSpace(string(out)), "bin")
}

func backupSettings(settingsPath string) (string, error) {
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return "", nil
	}

	backup := settingsPath + ".bak." + time.Now().Format("20060102150405")

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return "", err
	}

	return backup, os.WriteFile(backup, data, 0o600)
}

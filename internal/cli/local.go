package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/otakakot/lookback-cc/internal/version"
)

const versionPkgPath = "github.com/otakakot/lookback-cc/internal/version"

type localInstaller struct {
	projectRoot  string
	localVersion string
}

func (l *localInstaller) label() string {
	return "project: " + l.projectRoot
}

func (l *localInstaller) version() string {
	return l.localVersion
}

func (l *localInstaller) install(cmd, gobin string) error {
	return goInstallLocal(l.projectRoot, "./cmd/"+cmd, l.localVersion, gobin)
}

func RunLocal() int {
	projectRoot, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	localVersion := version.Version + "-local"

	return runInstall(&localInstaller{
		projectRoot:  projectRoot,
		localVersion: localVersion,
	}, true)
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get working directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found in any parent directory")
		}

		dir = parent
	}
}

func goInstallLocal(projectRoot, pkg, localVersion, gobin string) error {
	ldflags := fmt.Sprintf("-X %s.Version=%s", versionPkgPath, localVersion)

	cmd := exec.Command("go", "install", "-ldflags", ldflags, pkg)
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if gobin != "" {
		cmd.Env = append(os.Environ(), "GOBIN="+gobin)
	}

	return cmd.Run()
}

package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	tmpFile := os.Getenv("CCCALL_TMP")
	outFile := os.Getenv("CCCALL_OUT")
	prompt := os.Getenv("CCCALL_PROMPT")
	cwd := os.Getenv("CCCALL_CWD")

	if tmpFile == "" || outFile == "" || prompt == "" {
		fmt.Fprintln(os.Stderr, "cccall: missing required environment variables")
		os.Exit(1)
	}

	// Ensure temp file cleanup on signals.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigCh
		os.Remove(tmpFile)
		os.Exit(1)
	}()
	defer os.Remove(tmpFile)

	in, err := os.Open(tmpFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cccall: open tmp: %v\n", err)
		os.Exit(1)
	}
	defer in.Close()

	cmd := exec.Command("claude", "--print", "--bare", prompt)
	cmd.Stdin = in

	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cccall: claude: %v\n", err)
		os.Exit(1)
	}
	if len(out) == 0 {
		fmt.Fprintln(os.Stderr, "cccall: claude returned empty output")
		return
	}

	f, err := os.Create(outFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cccall: create output: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	fmt.Fprintf(f, "# 作業ディレクトリ: %s\n\n", cwd)
	if _, err := f.Write(out); err != nil {
		fmt.Fprintf(os.Stderr, "cccall: write output: %v\n", err)
		os.Exit(1)
	}
}

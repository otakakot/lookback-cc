package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	tmpFile := os.Getenv("SUMMARIZE_TMP")
	outFile := os.Getenv("SUMMARIZE_OUT")
	prompt := os.Getenv("SUMMARIZE_PROMPT")
	cwd := os.Getenv("SUMMARIZE_CWD")

	if tmpFile == "" || outFile == "" || prompt == "" {
		fmt.Fprintln(os.Stderr, "summarize: missing required environment variables")
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
		fmt.Fprintf(os.Stderr, "summarize: open tmp: %v\n", err)
		os.Exit(1)
	}
	defer in.Close()

	cmd := exec.Command("claude", "--print", "--bare", prompt)
	cmd.Stdin = in

	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "summarize: claude: %v\n", err)
		os.Exit(1)
	}
	if len(out) == 0 {
		fmt.Fprintln(os.Stderr, "summarize: claude returned empty output")
		return
	}

	f, err := os.Create(outFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "summarize: create output: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	fmt.Fprintf(f, "# 作業ディレクトリ: %s\n\n", cwd)
	if _, err := f.Write(out); err != nil {
		fmt.Fprintf(os.Stderr, "summarize: write output: %v\n", err)
		os.Exit(1)
	}
}

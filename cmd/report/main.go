package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const promptTemplate = `以下は1日のClaude Codeセッション要約です。
作業ディレクトリごとに統合し、簡潔なデイリーレポートを生成してください。

# フォーマット

## %s のまとめ

やったことを1-2文で。

## やったこと

作業ディレクトリごとに、やったことを箇条書きで簡潔に列挙。
冗長な説明は不要。各項目は1行で。同じディレクトリの複数セッションは統合。

### <作業ディレクトリ>
- 項目1
- 項目2

出力はマークダウンのみ（コードブロックで囲まない）。日本語で記述してください。`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: report [date]

Generates a daily report from lookback session summaries.

Arguments:
  date    Target date in YYYY-MM-DD format (default: today)

Input:  ~/.claude/debrief/<date>/*.md
Output: ~/.claude/report/<YYYY>/<MM>/<DD>.md
`)
	}

	flag.Parse()

	date := time.Now().Format("2006-01-02")
	if flag.NArg() > 0 {
		date = flag.Arg(0)
	}

	dayDir := filepath.Join(os.Getenv("HOME"), ".claude", "debrief", date)

	entries, err := filepath.Glob(filepath.Join(dayDir, "*.md"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "report: glob: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Fprintf(os.Stderr, "report: no markdown files found in %s\n", dayDir)
		os.Exit(1)
	}

	sort.Strings(entries)

	var sb strings.Builder

	for _, path := range entries {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "report: read %s: %v\n", path, err)
			os.Exit(1)
		}

		if len(data) == 0 {
			continue
		}

		fmt.Fprintf(&sb, "--- セッション: %s ---\n", filepath.Base(path))
		sb.Write(data)
		sb.WriteString("\n\n")
	}

	if sb.Len() == 0 {
		fmt.Fprintln(os.Stderr, "report: all session files are empty")
		os.Exit(1)
	}

	prompt := fmt.Sprintf(promptTemplate, date)

	cmd := exec.Command("claude", "--print", "--bare", prompt)
	cmd.Stdin = strings.NewReader(sb.String())

	var stderrBuf strings.Builder

	cmd.Stderr = &stderrBuf

	parts := strings.SplitN(date, "-", 3)
	if len(parts) != 3 {
		fmt.Fprintf(os.Stderr, "report: invalid date format: %s (expected YYYY-MM-DD)\n", date)
		os.Exit(1)
	}

	outDir := filepath.Join(os.Getenv("HOME"), ".claude", "report", parts[0], parts[1])
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "report: mkdir: %v\n", err)
		os.Exit(1)
	}

	outFile := filepath.Join(outDir, parts[2]+".md")

	stop := make(chan struct{})
	done := make(chan struct{})

	go func() {
		defer close(done)

		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		i := 0

		for {
			select {
			case <-stop:
				fmt.Fprintf(os.Stderr, "\r\033[K")
				return
			case <-ticker.C:
				fmt.Fprintf(os.Stderr, "\r%s レポート生成中...", frames[i%len(frames)])
				i++
			}
		}
	}()

	out, err := cmd.Output()

	close(stop)
	<-done

	if stderrBuf.Len() > 0 {
		fmt.Fprint(os.Stderr, stderrBuf.String())
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "report: claude: %v\n", err)
		os.Exit(1)
	}

	if len(out) == 0 {
		fmt.Fprintln(os.Stderr, "report: claude returned empty output")
		os.Exit(1)
	}

	if err := os.WriteFile(outFile, out, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "report: write: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("report: saved to %s\n", outFile)
}

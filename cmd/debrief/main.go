package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/otakakot/lookback-cc/internal/transcript"
)

const promptTemplate = `以下はClaude Codeのセッション中の会話履歴です。この会話を振り返り、以下の形式でマークダウンの要約を生成してください。

# フォーマット

## 概要
セッション全体で何を行ったかの簡潔な説明（2-3文）

## 作業内容
具体的に行った作業のリスト（箇条書き）

## 主な議論・決定事項
セッション中に議論されたこと、決定されたことのリスト

## 残課題・次のステップ
未完了の作業や次にやるべきことがあれば記載

---
作業ディレクトリ: %s
出力はマークダウンのみ（コードブロックで囲まない）。日本語で記述してください。`

// HookInput is the JSON payload received on stdin from the SessionEnd hook.
type HookInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
}

func main() {
	var input HookInput
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "debrief: failed to read hook input: %v\n", err)
		os.Exit(1)
	}

	if input.TranscriptPath == "" {
		fmt.Fprintln(os.Stderr, "debrief: no transcript_path in hook input")
		os.Exit(1)
	}

	if input.Cwd == "" {
		fmt.Fprintln(os.Stderr, "debrief: no cwd in hook input")
		os.Exit(1)
	}

	home := os.Getenv("HOME")
	if home == "" {
		fmt.Fprintln(os.Stderr, "debrief: HOME is not set")
		os.Exit(1)
	}

	// Parse the conversation transcript.
	turns, err := transcript.Parse(input.TranscriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "debrief: %v\n", err)
		os.Exit(1)
	}

	if len(turns) == 0 {
		os.Exit(0)
	}

	conversationText := transcript.FormatForSummary(turns)

	// Prepare output directory and file.
	now := time.Now()

	outDir := filepath.Join(home, ".claude", "debrief", now.Format("2006-01-02"))
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "debrief: mkdir: %v\n", err)
		os.Exit(1)
	}

	outFile := filepath.Join(outDir, now.Format("15-04-05")+".md")

	prompt := fmt.Sprintf(promptTemplate, input.Cwd)

	// Spawn summarize in background to run claude and write the summary.
	spawnCccall(conversationText, outFile, prompt, input.Cwd)
}

// spawnCccall writes conversation text to a temp file and launches the summarize
// command in the background. summarize handles running claude, writing the output
// file, and cleaning up the temp file. The parent returns immediately.
func spawnCccall(conversation, outFile, prompt, cwd string) {
	tmpFile, err := os.CreateTemp("", "debrief-*.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "debrief: tmpfile: %v\n", err)
		os.Exit(1)
	}

	if _, err := tmpFile.WriteString(conversation); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		fmt.Fprintf(os.Stderr, "debrief: write tmp: %v\n", err)
		os.Exit(1)
	}
	tmpFile.Close()

	cmd := exec.Command("summarize")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // detach from parent session
	}
	cmd.Env = append(os.Environ(),
		"SUMMARIZE_PROMPT="+prompt,
		"SUMMARIZE_TMP="+tmpFile.Name(),
		"SUMMARIZE_OUT="+outFile,
		"SUMMARIZE_CWD="+cwd,
	)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		os.Remove(tmpFile.Name())
		fmt.Fprintf(os.Stderr, "debrief: spawn summarize: %v\n", err)
		os.Exit(1)
	}

	cmd.Process.Release()
}

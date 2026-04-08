package transcript

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Entry represents a single line in the JSONL transcript.
type Entry struct {
	Type    string   `json:"type"`
	Message *Message `json:"message,omitempty"`
}

// Message represents a Claude API message.
type Message struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// ContentBlock represents one block inside a message's content array.
type ContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Thinking string `json:"thinking,omitempty"`
	Name     string `json:"name,omitempty"` // tool_use name
}

// Turn is a simplified representation of one conversation turn.
type Turn struct {
	Role string // "user" or "assistant"
	Text string
}

// Parse reads a JSONL transcript file and returns conversation turns
// containing only the textual user/assistant messages (no tool calls/results).
func Parse(path string) ([]Turn, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open transcript: %w", err)
	}
	defer f.Close()

	var turns []Turn

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // up to 10MB per line

	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		if entry.Message == nil {
			continue
		}

		if entry.Type != "user" && entry.Type != "assistant" {
			continue
		}

		text := extractText(entry.Message.Content)
		if text == "" {
			continue
		}

		turns = append(turns, Turn{Role: entry.Type, Text: text})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan transcript: %w", err)
	}

	return turns, nil
}

// FormatForSummary converts turns into a readable conversation text
// suitable for feeding into a summarizer.
func FormatForSummary(turns []Turn) string {
	var sb strings.Builder

	for _, t := range turns {
		switch t.Role {
		case "user":
			sb.WriteString("## User\n")
		case "assistant":
			sb.WriteString("## Assistant\n")
		}

		sb.WriteString(t.Text)
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func extractText(raw json.RawMessage) string {
	// content can be a string or an array of content blocks
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}

	var blocks []ContentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return ""
	}

	var parts []string

	for _, b := range blocks {
		if b.Type == "text" && b.Text != "" {
			parts = append(parts, b.Text)
		}
	}

	return strings.TrimSpace(strings.Join(parts, "\n"))
}

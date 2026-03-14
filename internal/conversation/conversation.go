package conversation

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Line represents a single displayable conversation line.
type Line struct {
	Icon      string
	Text      string
	Timestamp time.Time
}

// jsonlEntry represents a single line in a Claude Code JSONL conversation file.
type jsonlEntry struct {
	Type      string       `json:"type"`
	Timestamp time.Time    `json:"timestamp"`
	Message   *jsonlMessage `json:"message,omitempty"`
}

type jsonlMessage struct {
	Content json.RawMessage `json:"content"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ProjectDir derives the Claude Code project directory name from a CWD path.
// Claude Code replaces "/" with "-" and removes "." characters.
func ProjectDir(cwd string) string {
	s := strings.ReplaceAll(cwd, "/", "-")
	s = strings.ReplaceAll(s, ".", "-")
	return s
}

// JSONLPath returns the full path to a Claude Code JSONL conversation file.
func JSONLPath(claudeDir, cwd, sessionID string) string {
	return filepath.Join(claudeDir, "projects", ProjectDir(cwd), sessionID+".jsonl")
}

// ReadLines reads conversation lines from a Claude Code JSONL file.
// It returns only user and assistant text messages, skipping tool calls,
// tool results, thinking blocks, and system messages.
// limit controls the maximum number of lines returned (0 = unlimited);
// when limited, returns the most recent lines.
func ReadLines(path string, limit int) ([]Line, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []Line
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		var entry jsonlEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		if entry.Type != "user" && entry.Type != "assistant" {
			continue
		}
		if entry.Message == nil {
			continue
		}
		if line := parseLine(entry); line != nil {
			lines = append(lines, *line)
		}
	}

	if limit > 0 && len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}

	return lines, scanner.Err()
}

func parseLine(entry jsonlEntry) *Line {
	raw := entry.Message.Content

	// Content can be a string.
	var strContent string
	if err := json.Unmarshal(raw, &strContent); err == nil {
		text := cleanText(strContent)
		if text == "" {
			return nil
		}
		return &Line{Icon: iconFor(entry.Type), Text: text, Timestamp: entry.Timestamp}
	}

	// Content is an array of blocks.
	var blocks []json.RawMessage
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return nil
	}

	var hasToolResult bool
	var textParts []string

	for _, b := range blocks {
		var block contentBlock
		if err := json.Unmarshal(b, &block); err != nil {
			continue
		}
		switch block.Type {
		case "text":
			if t := strings.TrimSpace(block.Text); t != "" {
				textParts = append(textParts, t)
			}
		case "tool_result":
			hasToolResult = true
		}
	}

	if len(textParts) == 0 {
		return nil
	}
	// Tool result messages are API plumbing, not real user input.
	if hasToolResult {
		return nil
	}

	text := cleanText(strings.Join(textParts, " "))
	return &Line{Icon: iconFor(entry.Type), Text: text, Timestamp: entry.Timestamp}
}

func iconFor(msgType string) string {
	if msgType == "assistant" {
		return "🤖"
	}
	return "👤"
}

func cleanText(s string) string {
	return strings.TrimSpace(s)
}

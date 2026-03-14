package conversation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectDir(t *testing.T) {
	tests := []struct {
		cwd  string
		want string
	}{
		{"/home/fred/git/cc-queue", "-home-fred-git-cc-queue"},
		{"/home/fred/.dotfiles/zsh", "-home-fred--dotfiles-zsh"},
		{"/home/fred/.dotfiles/kitty/.config/kitty", "-home-fred--dotfiles-kitty--config-kitty"},
		{"/tmp/myproject", "-tmp-myproject"},
	}
	for _, tt := range tests {
		got := ProjectDir(tt.cwd)
		if got != tt.want {
			t.Errorf("ProjectDir(%q) = %q, want %q", tt.cwd, got, tt.want)
		}
	}
}

func TestReadLines_Empty(t *testing.T) {
	path := writeTempJSONL(t, "")
	lines, err := ReadLines(path, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("got %d lines, want 0", len(lines))
	}
}

func TestReadLines_UserAndAssistantText(t *testing.T) {
	jsonl := `{"type":"user","timestamp":"2026-03-13T10:00:00Z","message":{"content":[{"type":"text","text":"hello world"}]}}
{"type":"assistant","timestamp":"2026-03-13T10:00:01Z","message":{"content":[{"type":"text","text":"hi there"}]}}
`
	path := writeTempJSONL(t, jsonl)
	lines, err := ReadLines(path, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}
	if lines[0].Icon != "👤" || lines[0].Text != "hello world" {
		t.Errorf("line 0 = %+v, want 👤 'hello world'", lines[0])
	}
	if lines[1].Icon != "🤖" || lines[1].Text != "hi there" {
		t.Errorf("line 1 = %+v, want 🤖 'hi there'", lines[1])
	}
}

func TestReadLines_StringContent(t *testing.T) {
	jsonl := `{"type":"user","timestamp":"2026-03-13T10:00:00Z","message":{"content":"plain string message"}}
`
	path := writeTempJSONL(t, jsonl)
	lines, err := ReadLines(path, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}
	if lines[0].Icon != "👤" || lines[0].Text != "plain string message" {
		t.Errorf("line 0 = %+v, want 👤 'plain string message'", lines[0])
	}
}

func TestReadLines_SkipsToolUseOnly(t *testing.T) {
	jsonl := `{"type":"assistant","timestamp":"2026-03-13T10:00:00Z","message":{"content":[{"type":"tool_use","name":"Read"}]}}
`
	path := writeTempJSONL(t, jsonl)
	lines, err := ReadLines(path, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("got %d lines, want 0 (tool_use only should be skipped)", len(lines))
	}
}

func TestReadLines_SkipsToolResult(t *testing.T) {
	jsonl := `{"type":"user","timestamp":"2026-03-13T10:00:00Z","message":{"content":[{"type":"tool_result","content":"file contents"}]}}
`
	path := writeTempJSONL(t, jsonl)
	lines, err := ReadLines(path, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("got %d lines, want 0 (tool_result should be skipped)", len(lines))
	}
}

func TestReadLines_SkipsThinkingOnly(t *testing.T) {
	jsonl := `{"type":"assistant","timestamp":"2026-03-13T10:00:00Z","message":{"content":[{"type":"thinking","text":"let me think..."}]}}
`
	path := writeTempJSONL(t, jsonl)
	lines, err := ReadLines(path, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("got %d lines, want 0 (thinking only should be skipped)", len(lines))
	}
}

func TestReadLines_AssistantTextWithToolUse(t *testing.T) {
	// Assistant says something then calls a tool — should show the text
	jsonl := `{"type":"assistant","timestamp":"2026-03-13T10:00:00Z","message":{"content":[{"type":"text","text":"Let me check that."},{"type":"tool_use","name":"Read"}]}}
`
	path := writeTempJSONL(t, jsonl)
	lines, err := ReadLines(path, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}
	if lines[0].Icon != "🤖" || lines[0].Text != "Let me check that." {
		t.Errorf("line 0 = %+v, want 🤖 'Let me check that.'", lines[0])
	}
}

func TestReadLines_SkipsNonMessageTypes(t *testing.T) {
	jsonl := `{"type":"progress","data":{"type":"hook_progress"}}
{"type":"file-history-snapshot","data":{}}
{"type":"system","message":{"content":"system msg"}}
{"type":"user","timestamp":"2026-03-13T10:00:00Z","message":{"content":[{"type":"text","text":"real message"}]}}
`
	path := writeTempJSONL(t, jsonl)
	lines, err := ReadLines(path, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1 (only user message)", len(lines))
	}
}

func TestReadLines_Limit(t *testing.T) {
	jsonl := `{"type":"user","timestamp":"2026-03-13T10:00:00Z","message":{"content":[{"type":"text","text":"first"}]}}
{"type":"assistant","timestamp":"2026-03-13T10:00:01Z","message":{"content":[{"type":"text","text":"second"}]}}
{"type":"user","timestamp":"2026-03-13T10:00:02Z","message":{"content":[{"type":"text","text":"third"}]}}
{"type":"assistant","timestamp":"2026-03-13T10:00:03Z","message":{"content":[{"type":"text","text":"fourth"}]}}
`
	path := writeTempJSONL(t, jsonl)
	lines, err := ReadLines(path, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}
	// Should be the last 2 lines
	if lines[0].Text != "third" {
		t.Errorf("line 0 = %q, want 'third'", lines[0].Text)
	}
	if lines[1].Text != "fourth" {
		t.Errorf("line 1 = %q, want 'fourth'", lines[1].Text)
	}
}

func TestReadLines_NewlinesPreserved(t *testing.T) {
	jsonl := `{"type":"user","timestamp":"2026-03-13T10:00:00Z","message":{"content":[{"type":"text","text":"line one\nline two\nline three"}]}}
`
	path := writeTempJSONL(t, jsonl)
	lines, err := ReadLines(path, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}
	if lines[0].Text != "line one\nline two\nline three" {
		t.Errorf("text = %q, want newlines preserved", lines[0].Text)
	}
}

func TestReadLines_FileNotFound(t *testing.T) {
	lines, err := ReadLines("/nonexistent/file.jsonl", 0)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if lines != nil {
		t.Errorf("got %v, want nil", lines)
	}
}

func TestReadLines_MalformedLines(t *testing.T) {
	jsonl := `not json at all
{"type":"user","timestamp":"2026-03-13T10:00:00Z","message":{"content":[{"type":"text","text":"valid"}]}}
{broken json
`
	path := writeTempJSONL(t, jsonl)
	lines, err := ReadLines(path, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1 (only valid line)", len(lines))
	}
	if lines[0].Text != "valid" {
		t.Errorf("text = %q, want 'valid'", lines[0].Text)
	}
}

func TestReadLines_SkipsEmptyTextBlocks(t *testing.T) {
	jsonl := `{"type":"assistant","timestamp":"2026-03-13T10:00:00Z","message":{"content":[{"type":"text","text":""},{"type":"text","text":"\n\n"}]}}
`
	path := writeTempJSONL(t, jsonl)
	lines, err := ReadLines(path, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("got %d lines, want 0 (empty text should be skipped)", len(lines))
	}
}

func TestJSONLPath(t *testing.T) {
	got := JSONLPath("/home/user/.claude", "/home/user/git/myproject", "abc-123")
	want := "/home/user/.claude/projects/-home-user-git-myproject/abc-123.jsonl"
	if got != want {
		t.Errorf("JSONLPath() = %q, want %q", got, want)
	}
}

func writeTempJSONL(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.jsonl")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeTempJSONL: %v", err)
	}
	return path
}

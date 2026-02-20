package cmd_test

import (
	"strings"
	"testing"

	"github.com/duboisf/cc-queue/cmd"
	"github.com/duboisf/cc-queue/internal/queue"
)

func TestPop_MarksAsWorking(t *testing.T) {
	setupQueueDir(t)
	t.Setenv("KITTY_WINDOW_ID", "42")

	seedEntry(t, "test-sess", "/tmp/project", "permission_prompt", 1)

	opts, _, _ := testOptionsWithStdin(`{"session_id":"test-sess","cwd":"/tmp/project"}`)
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "pop")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Entry should still exist but be marked as working.
	if n := entryCount(t); n != 1 {
		t.Fatalf("expected 1 entry after pop, got %d", n)
	}

	entries, _ := queue.List()
	if entries[0].Event != "working" {
		t.Errorf("expected event=working, got %q", entries[0].Event)
	}
}

func TestPop_CreatesWorkingEntryForNewSession(t *testing.T) {
	setupQueueDir(t)
	t.Setenv("KITTY_WINDOW_ID", "42")

	// No pre-existing entry — pop should create one.
	opts, _, _ := testOptionsWithStdin(`{"session_id":"new-sess","cwd":"/tmp/new-project"}`)
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "pop")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if n := entryCount(t); n != 1 {
		t.Fatalf("expected 1 entry, got %d", n)
	}

	entries, _ := queue.List()
	if entries[0].Event != "working" {
		t.Errorf("expected event=working, got %q", entries[0].Event)
	}
	if entries[0].SessionID != "new-sess" {
		t.Errorf("expected session_id=new-sess, got %q", entries[0].SessionID)
	}
	if !strings.Contains(entries[0].CWD, "new-project") {
		t.Errorf("expected cwd to contain new-project, got %q", entries[0].CWD)
	}
}

func TestPop_RemovesEntryWithoutKittyWindowID(t *testing.T) {
	setupQueueDir(t)
	// No KITTY_WINDOW_ID — falls back to removing the entry.
	t.Setenv("KITTY_WINDOW_ID", "")

	seedEntry(t, "test-sess", "/tmp/project", "permission_prompt", 1)
	if n := entryCount(t); n != 1 {
		t.Fatalf("expected 1 seeded entry, got %d", n)
	}

	opts, _, _ := testOptionsWithStdin(`{"session_id":"test-sess"}`)
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "pop")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if n := entryCount(t); n != 0 {
		t.Fatalf("expected 0 entries after pop without kitty, got %d", n)
	}
}

func TestPop_NonexistentEntryWithoutKitty(t *testing.T) {
	setupQueueDir(t)
	t.Setenv("KITTY_WINDOW_ID", "")

	opts, _, _ := testOptionsWithStdin(`{"session_id":"no-such-session"}`)
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "pop")
	if err != nil {
		t.Fatalf("expected no error for nonexistent entry, got %v", err)
	}
}

func TestPop_MalformedJSON(t *testing.T) {
	setupQueueDir(t)

	opts, _, _ := testOptionsWithStdin(`{not valid json}`)
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "pop")
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

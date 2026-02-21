package cmd_test

import (
	"testing"

	"github.com/duboisf/cc-queue/cmd"
)

func TestEnd_RemovesSession(t *testing.T) {
	setupQueueDir(t)

	seedEntry(t, "sess-end", "/tmp/project", "permission_prompt", 1001)
	if n := entryCount(t); n != 1 {
		t.Fatalf("expected 1 seeded entry, got %d", n)
	}

	input := `{"session_id":"sess-end","cwd":"/tmp/project","hook_event_name":"SessionEnd"}`
	opts, _, _ := testOptionsWithStdin(input)
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "end")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if n := entryCount(t); n != 0 {
		t.Fatalf("expected 0 entries after end, got %d", n)
	}
}

func TestEnd_NonexistentSession(t *testing.T) {
	setupQueueDir(t)

	input := `{"session_id":"no-such-session","hook_event_name":"SessionEnd"}`
	opts, _, _ := testOptionsWithStdin(input)
	root := cmd.NewRootCmd(opts)

	// Should not error even if session doesn't exist.
	_, _, err := executeCommand(root, "end")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestEnd_MalformedJSON(t *testing.T) {
	setupQueueDir(t)

	opts, _, _ := testOptionsWithStdin(`{not valid json}`)
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "end")
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

package cmd_test

import (
	"testing"

	"github.com/duboisf/cc-queue/cmd"
)

func TestPop_RemovesEntry(t *testing.T) {
	setupQueueDir(t)

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
		t.Fatalf("expected 0 entries after pop, got %d", n)
	}
}

func TestPop_NonexistentEntry(t *testing.T) {
	setupQueueDir(t)

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

package cmd_test

import (
	"testing"

	"github.com/duboisf/cc-queue/cmd"
)

func TestPush_ValidInput(t *testing.T) {
	setupQueueDir(t)
	t.Setenv("KITTY_WINDOW_ID", "42")

	input := `{"session_id":"test-sess","cwd":"/tmp/project","hook_event_name":"Notification","notification_type":"permission_prompt"}`
	opts, _, _ := testOptionsWithStdin(input)
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "push")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if n := entryCount(t); n != 1 {
		t.Fatalf("expected 1 entry, got %d", n)
	}
}

func TestPush_SkipsWithoutKittyWindowID(t *testing.T) {
	setupQueueDir(t)
	// Explicitly clear KITTY_WINDOW_ID (may be set when running inside kitty).
	t.Setenv("KITTY_WINDOW_ID", "")

	input := `{"session_id":"test-sess","cwd":"/tmp/project","hook_event_name":"Notification","notification_type":"permission_prompt"}`
	opts, _, _ := testOptionsWithStdin(input)
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "push")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if n := entryCount(t); n != 0 {
		t.Fatalf("expected 0 entries, got %d", n)
	}
}

func TestPush_MalformedJSON(t *testing.T) {
	setupQueueDir(t)
	t.Setenv("KITTY_WINDOW_ID", "42")

	opts, _, _ := testOptionsWithStdin(`{not valid json}`)
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "push")
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

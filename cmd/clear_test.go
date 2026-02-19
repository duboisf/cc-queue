package cmd_test

import (
	"testing"

	"github.com/duboisf/cc-queue/cmd"
)

func TestClear_RemovesAll(t *testing.T) {
	setupQueueDir(t)
	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	seedEntry(t, "s1", "/tmp", "Notification", 1)
	seedEntry(t, "s2", "/tmp", "Notification", 2)
	seedEntry(t, "s3", "/tmp", "Notification", 3)

	_, _, err := executeCommand(root, "clear")
	if err != nil {
		t.Fatalf("clear returned error: %v", err)
	}

	if got := stdout.String(); got != "Queue cleared\n" {
		t.Errorf("stdout = %q, want %q", got, "Queue cleared\n")
	}

	if n := entryCount(t); n != 0 {
		t.Errorf("entryCount = %d, want 0", n)
	}
}

func TestClear_EmptyQueue(t *testing.T) {
	setupQueueDir(t)
	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "clear")
	if err != nil {
		t.Fatalf("clear returned error: %v", err)
	}

	if got := stdout.String(); got != "Queue cleared\n" {
		t.Errorf("stdout = %q, want %q", got, "Queue cleared\n")
	}
}

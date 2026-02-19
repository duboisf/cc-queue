package cmd_test

import (
	"os"
	"testing"

	"github.com/duboisf/cc-queue/cmd"
)

func TestClean_RemovesStale(t *testing.T) {
	setupQueueDir(t)
	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	// Alive process (current PID)
	seedEntry(t, "alive", "/tmp", "Notification", os.Getpid())
	// Dead process (PID that should not exist)
	seedEntry(t, "dead", "/tmp", "Notification", 999999999)

	_, _, err := executeCommand(root, "clean")
	if err != nil {
		t.Fatalf("clean returned error: %v", err)
	}

	if got := stdout.String(); got != "Removed 1 stale entries\n" {
		t.Errorf("stdout = %q, want %q", got, "Removed 1 stale entries\n")
	}

	if n := entryCount(t); n != 1 {
		t.Errorf("entryCount = %d, want 1", n)
	}
}

func TestClean_EmptyQueue(t *testing.T) {
	setupQueueDir(t)
	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "clean")
	if err != nil {
		t.Fatalf("clean returned error: %v", err)
	}

	if got := stdout.String(); got != "Removed 0 stale entries\n" {
		t.Errorf("stdout = %q, want %q", got, "Removed 0 stale entries\n")
	}
}

func TestClean_AllAlive(t *testing.T) {
	setupQueueDir(t)
	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	seedEntry(t, "alive", "/tmp", "Notification", os.Getpid())

	_, _, err := executeCommand(root, "clean")
	if err != nil {
		t.Fatalf("clean returned error: %v", err)
	}

	if got := stdout.String(); got != "Removed 0 stale entries\n" {
		t.Errorf("stdout = %q, want %q", got, "Removed 0 stale entries\n")
	}

	if n := entryCount(t); n != 1 {
		t.Errorf("entryCount = %d, want 1", n)
	}
}

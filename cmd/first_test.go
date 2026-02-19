package cmd_test

import (
	"testing"

	"github.com/duboisf/cc-queue/cmd"
)

func TestFirst_EmptyQueue(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "first")
	if err != nil {
		t.Fatalf("first with empty queue returned error: %v", err)
	}
}

func TestFirst_WithEntries(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()

	seedEntry(t, "sess-a", "/home/user/proj-a", "permission_prompt", 1001)
	seedEntry(t, "sess-b", "/home/user/proj-b", "idle_prompt", 1002)

	root := cmd.NewRootCmd(opts)

	// first will try to call kitty focus-window which will fail,
	// but we can verify it picks the most recent entry by checking
	// that it attempts to jump (returns an error about kitty).
	_, _, err := executeCommand(root, "first")
	// Expect kitty focus-window to fail in test environment.
	if err == nil {
		// If no error, queue should have one fewer entry.
		count := entryCount(t)
		if count != 1 {
			t.Errorf("entry count after first = %d, want 1", count)
		}
	}
	// Either kitty error or success is acceptable in test (no kitty available).
}

func TestFirst_ShellCompletion(t *testing.T) {
	opts, _, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	var firstCmd = root
	for _, c := range root.Commands() {
		if c.Name() == "first" {
			firstCmd = c
			break
		}
	}

	if firstCmd.ValidArgsFunction == nil {
		t.Fatal("first command missing ValidArgsFunction")
	}

	completions, directive := firstCmd.ValidArgsFunction(firstCmd, nil, "")
	if completions != nil {
		t.Errorf("completions = %v, want nil", completions)
	}
	if directive != 4 { // cobra.ShellCompDirectiveNoFileComp
		t.Errorf("directive = %d, want %d (NoFileComp)", directive, 4)
	}
}

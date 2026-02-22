package cmd_test

import (
	"strings"
	"testing"

	"github.com/duboisf/cc-queue/cmd"
	"github.com/duboisf/cc-queue/internal/queue"
)

func TestKittyLaunchArgs(t *testing.T) {
	entry := &queue.Entry{
		CWD:           "/home/user/project",
		KittyWindowID: "42",
	}
	got := cmd.KittyLaunchArgs(entry)
	want := []string{"@", "launch", "--type=window", "--cwd=/home/user/project", "--match", "id:42"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("arg[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestKittyLaunchArgs_WithListenOn(t *testing.T) {
	entry := &queue.Entry{
		CWD:           "/home/user/project",
		KittyWindowID: "42",
		KittyListenOn: "unix:/tmp/kitty-sock",
	}
	got := cmd.KittyLaunchArgs(entry)
	want := []string{"@", "--to", "unix:/tmp/kitty-sock", "launch", "--type=window", "--cwd=/home/user/project", "--match", "id:42"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("arg[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestShellCmd_MissingSession(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()

	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "_shell", "nonexistent")
	if err != nil {
		t.Fatalf("expected no error for missing session, got %v", err)
	}
}

func TestShellCmd_NoKittyWindowID(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()

	seedEntryNoWindow(t, "sess-nowin", "/home/user/proj", "permission_prompt", 1001)

	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "_shell", "sess-nowin")
	if err != nil {
		t.Fatalf("expected no error for empty WID, got %v", err)
	}
	if n := entryCount(t); n != 1 {
		t.Errorf("expected entry to persist, got %d entries", n)
	}
}

func TestShellCmd_EntryPersistsOnFailure(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()

	seedEntry(t, "sess-shell", "/home/user/proj", "permission_prompt", 1001)

	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "_shell", "sess-shell")
	// kitty launch will fail in test env — that's expected.
	if err == nil {
		t.Log("kitty launch unexpectedly succeeded (kitty may be running)")
	}
	// Entry must NOT be removed on failure (unlike _jump).
	if n := entryCount(t); n != 1 {
		t.Errorf("expected entry to persist after failed launch, got %d entries", n)
	}
}

func TestShellCmd_Completion(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()

	seedEntry(t, "sess-comp-1", "/tmp/a", "idle_prompt", 1001)
	seedEntry(t, "sess-comp-2", "/tmp/b", "permission_prompt", 1002)

	root := cmd.NewRootCmd(opts)
	// Find the _shell subcommand.
	shellCmd, _, err := root.Find([]string{"_shell"})
	if err != nil {
		t.Fatalf("could not find _shell command: %v", err)
	}
	completions, directive := shellCmd.ValidArgsFunction(shellCmd, nil, "sess-comp")
	if directive != 4 { // cobra.ShellCompDirectiveNoFileComp = 4
		t.Errorf("directive = %d, want ShellCompDirectiveNoFileComp (4)", directive)
	}
	if len(completions) != 2 {
		t.Fatalf("expected 2 completions, got %d: %v", len(completions), completions)
	}
	joined := strings.Join(completions, ",")
	if !strings.Contains(joined, "sess-comp-1") || !strings.Contains(joined, "sess-comp-2") {
		t.Errorf("completions = %v, expected both sess-comp-1 and sess-comp-2", completions)
	}
}

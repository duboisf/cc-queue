package cmd_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/duboisf/cc-queue/cmd"
)

func TestList_Empty(t *testing.T) {
	setupQueueDir(t)
	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	want := "No active sessions\n"
	if got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestList_WithEntries(t *testing.T) {
	setupQueueDir(t)
	opts, stdout, _ := testOptions()

	seedEntry(t, "sess-1", "/home/user/project-a", "permission_prompt", 1001)
	seedEntry(t, "sess-2", "/home/user/project-b", "idle_prompt", 1002)

	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()

	if !strings.Contains(got, "PERM") {
		t.Errorf("output missing PERM label:\n%s", got)
	}
	if !strings.Contains(got, "IDLE") {
		t.Errorf("output missing IDLE label:\n%s", got)
	}
	if !strings.Contains(got, "project-a") {
		t.Errorf("output missing cwd project-a:\n%s", got)
	}
	if !strings.Contains(got, "project-b") {
		t.Errorf("output missing cwd project-b:\n%s", got)
	}
}

func TestList_WithWorkingEntry(t *testing.T) {
	setupQueueDir(t)
	opts, stdout, _ := testOptions()

	seedEntry(t, "sess-work", "/home/user/project-a", "working", 1001)

	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "WORK") {
		t.Errorf("output missing WORK label:\n%s", got)
	}
}

func TestList_AttentionFirstSorting(t *testing.T) {
	setupQueueDir(t)
	opts, stdout, _ := testOptions()

	// Seed working entry first (older timestamp)
	seedEntryAtTime(t, "sess-work", "/home/user/project-a", "working", 1001, -10)
	// Then attention-needed entry (newer)
	seedEntryAtTime(t, "sess-perm", "/home/user/project-b", "permission_prompt", 1002, -5)

	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), got)
	}
	// PERM should come before WORK (attention-first sorting)
	if !strings.Contains(lines[0], "PERM") {
		t.Errorf("first line should have PERM, got %q", lines[0])
	}
	if !strings.Contains(lines[1], "WORK") {
		t.Errorf("second line should have WORK, got %q", lines[1])
	}
}

func TestList_OutputFormat(t *testing.T) {
	setupQueueDir(t)
	opts, stdout, _ := testOptions()

	seedEntry(t, "sess-fmt", "/tmp/myproject", "permission_prompt", 2001)

	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d: %q", len(lines), got)
	}

	line := lines[0]
	if !strings.Contains(line, "PERM") {
		t.Errorf("line missing PERM label: %q", line)
	}
	if !strings.Contains(line, "/tmp/myproject") {
		t.Errorf("line missing path: %q", line)
	}

	parts := strings.Fields(line)
	if len(parts) < 3 {
		t.Fatalf("expected at least 3 fields, got %d: %q", len(parts), line)
	}
	if parts[1] != "PERM" {
		t.Errorf("second field = %q, want %q", parts[1], "PERM")
	}
}

func TestListFzf_Empty(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	stdout, _, err := executeCommand(root, "_list-fzf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" {
		t.Errorf("output = %q, want empty", stdout)
	}
}

func TestListFzf_WithEntries(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()

	seedEntry(t, "sess-fzf-1", "/home/user/proj-a", "permission_prompt", 1001)
	seedEntry(t, "sess-fzf-2", "/home/user/proj-b", "idle_prompt", 1002)

	root := cmd.NewRootCmd(opts)
	stdout, _, err := executeCommand(root, "_list-fzf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Each line should be: session_id\tage label  path
	if !strings.Contains(stdout, "sess-fzf-1\t") {
		t.Errorf("output missing sess-fzf-1 tab-delimited:\n%s", stdout)
	}
	if !strings.Contains(stdout, "sess-fzf-2\t") {
		t.Errorf("output missing sess-fzf-2 tab-delimited:\n%s", stdout)
	}
}

func TestPreview_ExistingEntry(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()

	seedEntryWithMessage(t, "sess-prev", "/home/user/proj", "permission_prompt", 1001, "Allow read access?")

	root := cmd.NewRootCmd(opts)
	stdout, _, err := executeCommand(root, "_preview", "sess-prev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "PERM") {
		t.Errorf("output missing PERM label:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Allow read access?") {
		t.Errorf("output missing message:\n%s", stdout)
	}
}

func TestPreview_WorkingEntry(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()

	seedEntry(t, "sess-work", "/home/user/proj", "working", 1001)

	root := cmd.NewRootCmd(opts)
	stdout, _, err := executeCommand(root, "_preview", "sess-work")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "WORK") {
		t.Errorf("output missing WORK label:\n%s", stdout)
	}
}

func TestPreview_MissingEntry(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()

	root := cmd.NewRootCmd(opts)
	stdout, _, err := executeCommand(root, "_preview", "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" {
		t.Errorf("output = %q, want empty", stdout)
	}
}

// errFullTabber is a mock FullTabber that returns an error.
type errFullTabber struct{}

func (e *errFullTabber) EnterFullTab() (func(), error) {
	return func() {}, errors.New("kitty not available")
}

// trackingFullTabber records calls for verification.
type trackingFullTabber struct {
	entered  bool
	restored bool
}

func (t *trackingFullTabber) EnterFullTab() (func(), error) {
	t.entered = true
	return func() { t.restored = true }, nil
}

func TestFirst_FullTab_ErrorReturned(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()
	opts.FullTabber = &errFullTabber{}

	seedEntry(t, "sess-ft", "/home/user/proj", "permission_prompt", 1001)

	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "first", "--full-tab")
	if err == nil {
		t.Fatal("expected error from FullTabber, got nil")
	}
	if !strings.Contains(err.Error(), "kitty not available") {
		t.Errorf("error = %q, want to contain 'kitty not available'", err.Error())
	}
}

func TestFirst_FullTab_Called(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()
	ft := &trackingFullTabber{}
	opts.FullTabber = ft

	// first needs an attention-needed entry to jump to.
	seedEntry(t, "sess-first", "/home/user/proj", "permission_prompt", 1001)

	root := cmd.NewRootCmd(opts)
	_, _, _ = executeCommand(root, "first", "--full-tab")

	if !ft.entered {
		t.Error("EnterFullTab was not called")
	}
	if !ft.restored {
		t.Error("restore was not called")
	}
}

func TestFirst_SkipsWorkingSessions(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()
	ft := &trackingFullTabber{}
	opts.FullTabber = ft

	// Only working sessions — first should do nothing.
	seedEntry(t, "sess-work", "/home/user/proj", "working", 1001)

	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "first")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// FullTabber should NOT have been called since there's nothing to jump to.
	if ft.entered {
		t.Error("EnterFullTab should not be called when only working sessions exist")
	}
}

func TestRootCmd_FullTab_Error(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()
	opts.FullTabber = &errFullTabber{}

	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "--full-tab")
	if err == nil {
		t.Fatal("expected error from FullTabber, got nil")
	}
}

func TestFirst_NoFullTab_NotCalled(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()
	ft := &trackingFullTabber{}
	opts.FullTabber = ft

	root := cmd.NewRootCmd(opts)
	_, _, _ = executeCommand(root, "first")

	if ft.entered {
		t.Error("EnterFullTab should not be called without --full-tab")
	}
}

func TestJumpInternal_KeepsEntry(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()

	// Entry with no KittyWindowID — jump should succeed without removing.
	seedEntryNoWindow(t, "sess-jump", "/home/user/proj", "permission_prompt", 1001)

	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "_jump", "sess-jump")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Entry should persist (no window ID means nothing to focus, but entry stays).
	if entryCount(t) != 1 {
		t.Error("expected entry to persist after jump (no window ID)")
	}
}

func TestJumpInternal_MissingEntry(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()

	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "_jump", "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJumpInternal_StaleWindow_RemovesEntry(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()

	// Entry with a window ID that won't exist — kitty @ focus-window will fail.
	seedEntry(t, "sess-stale", "/home/user/proj", "permission_prompt", 1001)

	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "_jump", "sess-stale")
	// Error is expected (window doesn't exist), and entry should be removed (stale).
	if err == nil {
		t.Log("kitty focus-window unexpectedly succeeded (kitty may be running)")
	}
	if entryCount(t) != 0 {
		t.Error("expected stale entry to be removed after failed jump")
	}
}

func TestListFzf_CleanStaleWindowsCalled(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()

	called := false
	opts.CleanStaleWindowsFn = func() { called = true }

	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "_list-fzf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("CleanStaleWindowsFn was not called")
	}
}

func TestListFzf_NilCleanStaleWindowsFn(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()
	opts.CleanStaleWindowsFn = nil

	seedEntry(t, "sess-nil", "/home/user/proj", "permission_prompt", 1001)

	root := cmd.NewRootCmd(opts)
	stdout, _, err := executeCommand(root, "_list-fzf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "sess-nil\t") {
		t.Errorf("output missing entry:\n%s", stdout)
	}
}
